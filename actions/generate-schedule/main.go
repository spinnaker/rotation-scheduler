package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/httpreplay"
	"github.com/ghodss/yaml"
	"github.com/spinnaker/rotation-scheduler/actions/generate-schedule/scheduler"
	"github.com/spinnaker/rotation-scheduler/actions/generate-schedule/users"
	"github.com/spinnaker/rotation-scheduler/actions/generate-schedule/users/ghteams"
	"github.com/spinnaker/rotation-scheduler/schedule"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

var (
	// Required.
	stop = flag.String("stop", "", "Required. Stop generating the schedule at or before this date. Date is inclusive in the schedule if the final shift ends on this date. Must be in the format 2006-01-02.")

	// Either --users or (--githubTeam and --githubOrg) are requried.
	userList   = flag.String("users", "", "Comma-separated list of users in the rotation. Usernames will be alphabetically sorted prior to assigning shifts.")
	githubTeam = flag.String("githubTeam", "", "Fetch the user list from the specified GitHub team. Must specify an access token with read:org permissions in the GITHUB_TOKEN env variable, as well as --githubTeam.")
	githubOrg  = flag.String("githubOrg", "", "GitHub org that owns the --githubTeam. Must specify with --githubTeam.")

	// Optional
	shiftDurationDays    = flag.Int("shiftDurationDays", 7, "The duration of each shift in days")
	start                = flag.String("start", "", "Start generating from this date. If this flag and previousSchedule are not specified, starts from tomorrow's date.")
	previousSchedulePath = flag.String("previousSchedule", "", "Path to a previous schedule's YAML file. Extends this schedule until --stop.")
	schedOutFilepath     = flag.String("scheduleOutput", "", "File to write out new schedule. Will overwrite the file if it exists. If not specified, writes to stdout.")
	record               = flag.Bool("record", false, "Record the responses from external dependencies to a file. Used for external dependency testing.")
	recordOutPath        = flag.String("recordOutput", "", "Filepath for saved HTTP responses.")

	// Used to store GitHub personal access token from environment.
	accessToken string
)

const (
	startStopFormat   = "2006-01-02"
	usersSeparator    = ","
	githubTokenEnvKey = "GITHUB_TOKEN"
)

func main() {
	flag.Parse()

	if err := validateFlags(); err != nil {
		log.Fatalf("Error validating flags: %v", err)
	}

	usersSource, closer, err := usersSource()
	if err != nil {
		log.Fatalf("error creating users source: %v", err)
	}
	defer func() {
		if closer != nil {
			_ = closer.Close()
		}
	}()

	sched, err := scheduler.NewScheduler(usersSource, *shiftDurationDays)
	if err != nil {
		log.Fatalf("Error creating scheduler: %v", err)
	}

	stopTime, _ := time.Parse(startStopFormat, *stop)

	var newSchedule *schedule.Schedule
	if *previousSchedulePath != "" {
		previousSchedule, err := previousSchedule(*previousSchedulePath)
		if err != nil {
			log.Fatalf("Error parsing previous schedule: %v", err)
		}

		err = sched.ExtendSchedule(previousSchedule, stopTime)
		if err != nil {
			log.Fatalf("Error extending schedule: %v", err)
		}
	} else {
		var startTime time.Time
		if *start != "" {
			startTime, _ = time.Parse(startStopFormat, *start)
		} else {
			startTime = time.Now().AddDate(0, 0, 1) // tomorrow.
		}

		newSchedule, err = sched.Schedule(startTime, stopTime)
		if err != nil {
			log.Fatalf("Error generating new schedule: %v", err)
		}
	}

	scheduleBytes, err := yaml.Marshal(newSchedule)
	if err != nil {
		log.Fatalf("Error marshalling schedule to yaml: %v", err)
	}

	destFilepath := os.Stdout.Name()
	if *schedOutFilepath != "" {
		destFilepath = *schedOutFilepath
	}
	if err := ioutil.WriteFile(destFilepath, scheduleBytes, 0666); err != nil {
		log.Fatalf("Error writing out new schedule: %v", err)
	}
}

func usersSource() (users.Source, io.Closer, error) {
	if *userList != "" {
		userSlice := strings.Split(*userList, usersSeparator)
		return users.NewStaticSource(userSlice...), nil, nil
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})

	var client *http.Client
	var closer io.Closer
	if *record {
		r, err := httpreplay.NewRecorder(*recordOutPath, []byte{})
		if err != nil {
			return nil, nil, fmt.Errorf("error setting up recorder: %v", err)
		}

		client, err = r.Client(context.Background(), option.WithTokenSource(ts))
		if err != nil {
			return nil, nil, fmt.Errorf("error creating HTTP client: %v", err)
		}
	} else {
		client = oauth2.NewClient(context.Background(), ts)
	}

	src, err := ghteams.NewGitHubTeamsUserSource(client, *githubOrg, *githubTeam)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get users from GitHub: %v", err)
	}

	return src, closer, nil
}

func validateFlags() error {
	if *stop == "" {
		return fmt.Errorf("--stop flag missing")
	} else if _, err := time.Parse(startStopFormat, *stop); err != nil {
		return fmt.Errorf("error parsing --stop value. Must be in the format %v: %v", startStopFormat, err)
	}

	if *shiftDurationDays <= 0 {
		return fmt.Errorf("--shiftDuration must be a positive integer")
	}

	if *previousSchedulePath != "" {
		if info, err := os.Stat(*previousSchedulePath); os.IsNotExist(err) {
			return fmt.Errorf("previousSchedule file (%v) not found: %v", *previousSchedulePath, err)
		} else if info.IsDir() {
			return fmt.Errorf("previousSchedule must be a file, got a directory")
		}
	} else if *start != "" {
		if _, err := time.Parse(startStopFormat, *start); err != nil {
			return fmt.Errorf("error parsing --start value. Must be in the format %v: %v", startStopFormat, err)
		}
	}

	if *githubTeam != "" || *githubOrg != "" {
		if *githubTeam == "" || *githubOrg == "" {
			return fmt.Errorf("must specify both --githubTeam and --githubOrg")
		} else if *userList != "" {
			return fmt.Errorf("cannot specify both --users and --github* options")
		}

		var ok bool
		accessToken, ok = os.LookupEnv(githubTokenEnvKey)
		if !ok || accessToken == "" {
			log.Fatalf("GITHUB_TOKEN environment variable cannot be empty. Get one from https://github.com/settings/tokens, and ensure it has the 'read:org' scope.")
		}
	} else if *userList == "" {
		log.Fatalf("must specify either non-empty --users or (--githubTeam and --githubOrg).")
	}

	return nil
}

func previousSchedule(previousSchedulePath string) (*schedule.Schedule, error) {
	prevBytes, err := ioutil.ReadFile(previousSchedulePath)
	if err != nil {
		return nil, err
	}

	prevSched := &schedule.Schedule{}
	if err := yaml.Unmarshal(prevBytes, prevSched); err != nil {
		return nil, err
	}

	return prevSched, nil
}

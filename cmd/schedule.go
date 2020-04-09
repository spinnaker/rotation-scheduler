package cmd

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/spinnaker/rotation-scheduler/schedule"
	"github.com/spinnaker/rotation-scheduler/users"
	"github.com/spinnaker/rotation-scheduler/users/ghteams"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

var (
	scheduleCmd = &cobra.Command{
		Use:   "schedule",
		Short: "Schedule creation and extension functions.",
		Long: "These options control the shift output, like providing the user list " +
			"(or how to get a user list) and how long to make each shift. For GitHub team integration, " +
			"users can be invited to their shift by making their email public.",
	}

	stopStr  string
	stopTime time.Time

	shiftDurationDays int

	userList    []string
	githubFlags []string

	emailDomains []string
)

func init() {
	scheduleCmd.PersistentFlags().StringVar(&stopStr, "stop", "", "Required. Generate schedule stopping on this date (inclusive). Must be in the format "+startStopFormat)
	_ = scheduleCmd.MarkPersistentFlagRequired("stop")

	scheduleCmd.PersistentFlags().IntVarP(&shiftDurationDays, "shiftDurationDays", "d", 7, "Optional. Duration in days for each shift. Defaults to 7, must be a positive integer.")

	scheduleCmd.PersistentFlags().StringSliceVarP(&userList, "users", "u", []string{}, "Set of users for the rotation. Required if --github* options are not specified.")

	scheduleCmd.PersistentFlags().StringSliceVarP(&githubFlags, "github", "g", []string{}, "Fetch the user list from GitHub. Order of args must be 'organization,team,accessToken'. Must specify an access token with read:org permissions.")

	scheduleCmd.PersistentFlags().StringSliceVar(&emailDomains, "domains", []string{"*"}, "Only include email addresses from --github that match these domains. A single value of '*' will allow any domain. Use '--domains []' to only use GitHub usernames.")

	RootCmd.AddCommand(scheduleCmd)
}

type githubDetails struct {
	org, team, accessToken string
}

func parseGithubDetails() (*githubDetails, error) {
	if githubFlags == nil || len(githubFlags) != 3 {
		return nil, fmt.Errorf("invalid --github value. Must be 'organization,team,accessToken'")
	}
	return &githubDetails{
		org:         githubFlags[0],
		team:        githubFlags[1],
		accessToken: githubFlags[2],
	}, nil
}

func userSrc() (users.Source, error) {
	var userSrc users.Source

	if len(userList) != 0 {
		userSrc = users.NewStaticSource(userList...)
	} else if len(githubFlags) != 0 {
		github, err := parseGithubDetails()
		if err != nil {
			return nil, err
		}
		client, closer, err := ghHttpClient(github)
		userSrc, err = ghteams.NewGitHubTeamsUserSource(client, github.org, github.team, emailDomains...)
		if err != nil {
			return nil, fmt.Errorf("error creating GitHub users source: %v", err)
		}
		defer func() {
			if closer != nil {
				_ = closer.Close()
			}
		}()
	}

	return userSrc, nil
}

func ghHttpClient(github *githubDetails) (*http.Client, io.Closer, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: github.accessToken})

	if recordFilepath == "" {
		return oauth2.NewClient(context.Background(), ts), nil, nil
	}

	r, err := recorder()
	if err != nil {
		return nil, nil, fmt.Errorf("error creating recorder: %v", err)
	}
	client, err := r.Client(context.Background(), option.WithTokenSource(ts))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating HTTP client: %v", err)
	}
	return client, r, nil
}

func marshalSchedule(sched *schedule.Schedule, destFilepath string) error {
	scheduleBytes, err := yaml.Marshal(sched)
	if err != nil {
		return fmt.Errorf("errror marshalling schedule to yaml: %v", err)
	}

	if err := ioutil.WriteFile(destFilepath, scheduleBytes, 0666); err != nil {
		return fmt.Errorf("error writing out new schedule: %v", err)
	}
	return nil
}

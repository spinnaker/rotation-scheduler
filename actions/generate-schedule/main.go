package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/spinnaker/rotation-scheduler/actions/generate-schedule/scheduler"
	"github.com/spinnaker/rotation-scheduler/actions/generate-schedule/users"
	"github.com/spinnaker/rotation-scheduler/schedule"
)

var (
	// Required flags
	stop              = flag.String("stop", "", "Stop generating the schedule at or before this date. Date is inclusive in the schedule if the final shift ends on this date. Must be in the format 2006-01-02.")
	shiftDurationDays = flag.Int("shiftDurationDays", 7, "The duration of each shift in days")
	// TODO(ttomsu): Flag is required now, make optional later (once GH teams integration works).
	userList = flag.String("users", "", "Comma-separated list of users in the rotation. Usernames will be alphabetically sorted prior to assigning shifts.")

	start                = flag.String("start", "", "Start generating from this date. If this flag and previousSchedule are not specified, starts from tomorrow's date.")
	previousSchedulePath = flag.String("previousSchedule", "", "Path to a previous schedule's YAML file.")

	outFilepath = flag.String("out", "", "File to write out new schedule. Will overwrite the file if it exists. If not specified, writes to stdout.")
)

const (
	startStopFormat = "2006-01-02"
	usersSeparator  = ","
)

func main() {
	flag.Parse()

	if err := validateFlags(); err != nil {
		log.Fatalf("Error validating flags: %v", err)
	}

	userSlice := strings.Split(*userList, usersSeparator)
	userSource := users.NewStaticSource(userSlice...)

	sched, err := scheduler.NewScheduler(userSource, *shiftDurationDays)
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
	if *outFilepath != "" {
		destFilepath = *outFilepath
	}
	if err := ioutil.WriteFile(destFilepath, scheduleBytes, 0666); err != nil {
		log.Fatalf("Error writing out new schedule: %v", err)
	}
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

	if *userList == "" {
		return fmt.Errorf("--users must be specified")
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

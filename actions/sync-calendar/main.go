package main

import (
	"flag"
	"fmt"
	"github.com/spinnaker/rotation-scheduler/actions/sync-calendar/gcal"
	"github.com/spinnaker/rotation-scheduler/proto/schedule"
	"log"
	"os"
)

var (
	schedPath  = flag.String("schedule", "", "Required. Path to schedule YAML file.")
	jsonKey    = flag.String("jsonKey", "", "Required. The path to the JSON key with access to the Calendar API.")
	calendarID = flag.String("calendarID", "build-cop@spinnaker.io", "Optional. The calendar ID to update. Defaults to build-cop@spinnaker.io")
)

func main() {
	flag.Parse()

	if err := validateFlags(); err != nil {
		log.Fatalf("Error validating flags: %v", err)
	}

	client, err := gcal.Client(*calendarID, *jsonKey)
	if err != nil {
		log.Fatalf("Error initializing GCal client: %v", err)
	}

	cal, err := gcal.NewGCal(*calendarID, client)
	if err != nil {
		log.Fatalf("Error initializing Calendar service: %v", err)
	}

	sched, err := schedule.FromYAMLFile(*schedPath)
	if err != nil {
		log.Fatalf("Error parsing schedule: %v", err)
	}

	stopTime, err := sched.EstimateStopTime()
	if err != nil {
		log.Fatalf("Error generating stop time: %v", stopTime)
	}

	if err := cal.Schedule(sched, stopTime); err != nil {
		log.Fatalf("Error syncing schedule: %v", err)
	}
}

func validateFlags() error {
	if *schedPath == "" {
		return fmt.Errorf("--schedule flag is required and must not be empty")
	} else if info, err := os.Stat(*schedPath); os.IsNotExist(err) {
		return fmt.Errorf("schedule file (%v) not found: %v", *schedPath, err)
	} else if info.IsDir() {
		return fmt.Errorf("schedule must be a file, got a directory")
	}

	if *jsonKey == "" {
		return fmt.Errorf("--jsonKey flag is required and must not be empty")
	} else if info, err := os.Stat(*jsonKey); os.IsNotExist(err) {
		return fmt.Errorf("jsonKey file (%v) not found: %v", *jsonKey, err)
	} else if info.IsDir() {
		return fmt.Errorf("jsonKey must be a file, got a directory")
	}

	return nil
}

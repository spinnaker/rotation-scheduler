package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spinnaker/rotation-scheduler/schedule/scheduler"
)

var (
	generateCmd = &cobra.Command{
		Use:   "generate [outputFile]",
		Short: "Generates a new schedule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  executeGenerate,
	}

	startStr  string
	startTime = time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour) // default to tomorrow, local time.
)

func init() {
	generateCmd.Flags().StringVar(&startStr, "start", "",
		"Optional. Generate schedule starting on this date (inclusive). "+
			"Defaults to tomorrow's date. Must be in the format "+startStopFormat)

	scheduleCmd.AddCommand(generateCmd)
}

func parseTimeFlags() error {
	var err error
	if startStr != "" {
		if startTime, err = time.Parse(startStopFormat, startStr); err != nil {
			return fmt.Errorf("error parsing --start: %v", err)
		}
	}
	if stopStr != "" {
		if stopTime, err = time.Parse(startStopFormat, stopStr); err != nil {
			return fmt.Errorf("error parsing --stop: %v", err)
		}
	}

	return nil
}

func executeGenerate(_ *cobra.Command, args []string) error {
	if err := parseTimeFlags(); err != nil {
		return err
	}

	userSrc, err := userSrc()
	if err != nil {
		return err
	}

	schdlr, err := scheduler.NewScheduler(userSrc, shiftDurationDays)
	if err != nil {
		return fmt.Errorf("error creating new scheduler: %v", err)
	}

	newSched, err := schdlr.Schedule(startTime, stopTime)
	if err != nil {
		return fmt.Errorf("error generating new schedule: %v", err)
	}

	destFilepath := os.Stdout.Name()
	if len(args) == 1 {
		destFilepath = args[0]
	}

	return marshalSchedule(newSched, destFilepath)
}

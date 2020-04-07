package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/spinnaker/rotation-scheduler/schedule"
	"github.com/spinnaker/rotation-scheduler/schedule/scheduler"
)

var (
	extendCmd = &cobra.Command{
		Use:   "extend [outputFile]",
		Short: "Extends a previously generated schedule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  executeExtend,
	}

	previousSchedulePath string

	prune bool
)

func init() {
	extendCmd.Flags().StringVarP(&previousSchedulePath, "schedule", "s", "", "Required. Filepath to the schedule to extend.")
	_ = extendCmd.MarkFlagRequired("schedule")
	_ = extendCmd.MarkFlagFilename("schedule", "yaml")

	extendCmd.Flags().BoolVarP(&prune, "prune", "p", false, "Prune removes all shifts before the current shift and reschedules all shifts if shift owners are no longer in rotation.")

	scheduleCmd.AddCommand(extendCmd)
}

func executeExtend(_ *cobra.Command, args []string) error {
	if err := parseTimeFlags(); err != nil {
		return err
	}

	sched, err := previousSchedule(previousSchedulePath)
	if err != nil {
		return fmt.Errorf("error parsing previous schedule: %v", err)
	}

	userSrc, err := userSrc()
	if err != nil {
		return err
	}

	schdlr, err := scheduler.NewScheduler(userSrc, shiftDurationDays)
	if err != nil {
		return fmt.Errorf("error creating new scheduler: %v", err)
	}

	err = schdlr.ExtendSchedule(sched, stopTime, prune)
	if err != nil {
		return fmt.Errorf("error generating new schedule: %v", err)
	}

	destFilepath := os.Stdout.Name()
	if len(args) == 1 {
		destFilepath = args[0]
	}

	return marshalSchedule(sched, destFilepath)
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

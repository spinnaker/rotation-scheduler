package cmd

import (
	"fmt"

	"cloud.google.com/go/httpreplay"
	"github.com/spf13/cobra"
)

const (
	startStopFormat = "2006-01-02"
)

var (
	rootCmd = &cobra.Command{
		Use:          "rotation",
		Short:        "rotation generates, extends, and syncs rotation schedules.",
		SilenceUsage: true,
	}

	recordFilepath string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&recordFilepath, "record", "r", "", "Record the responses from external dependencies to the specified file. Used for external dependency testing.")
}

func recorder() (*httpreplay.Recorder, error) {
	r, err := httpreplay.NewRecorder(recordFilepath, []byte{})
	if err != nil {
		return nil, fmt.Errorf("error setting up recorder: %v", err)
	}
	return r, nil
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

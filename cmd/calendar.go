package cmd

import (
	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Shared calender manipulation functions",
}

func init() {
	rootCmd.AddCommand(calendarCmd)
}

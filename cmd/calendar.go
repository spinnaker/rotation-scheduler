package cmd

import (
	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Shared calender manipulation functions",
	Long: "Many users prefer to have their shifts reflected on their calendar," +
		"rather than having to check a text file and make their own calendar events.",
}

func init() {
	RootCmd.AddCommand(calendarCmd)
}

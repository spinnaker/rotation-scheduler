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
	RootCmd = &cobra.Command{
		Use:   "rotation",
		Short: "`rotation` generates, extends, and syncs rotation schedules.",
		Long: `This tool generates and manipulates a human-readable and version-controllable
text file for an on-call, chore, or duty rotations among a set of individuals. Each shift only
needs a start date and the user on rotation. The stop date is implied by the start of
the next shift, except for the last shift, which is explicitly specified (the stop date is inclusive).
If a user needs to change or swap shifts, but keep the same rotation cycle, use the 'userOverride' field.

Example:
<pre>
shifts:
- startDate: Sun 01 Mar 2020
  user: abc
- startDate: Sun 08 Mar 2020
  user: lmn
  userOverride: xyz
- startDate: Sun 15 Mar 2020
  user: xyz
  userOverride: lmn
- startDate: Sun 22 Mar 2020
  stopDate: Sat 28 Mar 2020
  user: abc
</pre>

The --record option is used solely for testing. It records interactions with external services 
(like GitHub and Google Calendar) without having to write a bunch of brittle mock classes.
`,
		SilenceUsage: true,
	}

	recordFilepath string
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&recordFilepath, "record", "r", "", "Record the responses from external dependencies to the specified file. Used for external dependency testing.")
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
	return RootCmd.Execute()
}

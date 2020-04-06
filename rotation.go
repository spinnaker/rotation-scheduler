package main

import (
	"os"

	"github.com/spinnaker/rotation-scheduler/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

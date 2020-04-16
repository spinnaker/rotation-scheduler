package main

import (
	"log"
	"os"

	"github.com/spf13/cobra/doc"
	"github.com/spinnaker/rotation-scheduler/cmd"
)

func main() {
	_, err := os.Open("rotation.go")
	if err != nil {
		log.Fatalf("Must 'go run' this from the repo root. Example: go run docs/gendocs.go. Error: %v", err)
	}
	err = doc.GenMarkdownTree(cmd.RootCmd, "./docs")
	if err != nil {
		log.Fatal(err)
	}
}

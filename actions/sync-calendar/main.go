package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/httpreplay"
	"github.com/ghodss/yaml"
	"github.com/spinnaker/rotation-scheduler/actions/sync-calendar/gcal"
	"github.com/spinnaker/rotation-scheduler/schedule"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var (
	schedPath   = flag.String("schedule", "", "Required. Path to schedule YAML file.")
	jsonKeyPath = flag.String("jsonKey", "", "Required. The path to the JSON key with access to the Calendar API.")
	calendarID  = flag.String("calendarID", "spinbot@spinnaker.io", "Optional. The calendar ID to update. Defaults to spinbot@spinnaker.io, as that is the preferred test account.")

	record  = flag.Bool("record", false, "Record the responses from external dependencies to a file.")
	outPath = flag.String("out", "", "Filepath for saved HTTP responses.")
)

func main() {
	flag.Parse()

	if err := validateFlags(); err != nil {
		log.Fatalf("Error validating flags: %v", err)
	}

	client, closer, err := client()
	if err != nil {
		log.Fatalf("Error initializing HTTP client: %v", err)
	}
	defer func() {
		if closer != nil {
			_ = closer.Close()
		}
	}()

	cal, err := gcal.NewGCal(*calendarID, client)
	if err != nil {
		log.Fatalf("Error initializing Calendar service: %v", err)
	}

	schedBytes, err := ioutil.ReadFile(*schedPath)
	if err != nil {
		log.Fatalf("Error reading schedule file(%v): %v", *schedPath, err)
	}

	sched := &schedule.Schedule{}
	if err := yaml.Unmarshal(schedBytes, sched); err != nil {
		log.Fatalf("Error unmarshalling schedule: %v", err)
	}

	if err := cal.Schedule(sched); err != nil {
		log.Fatalf("Error syncing schedule: %v", err)
	}

}

func client() (*http.Client, io.Closer, error) {
	key, err := ioutil.ReadFile(*jsonKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read JSON credential file: %v", err)
	}
	jwtConfig, err := google.JWTConfigFromJSON(key, calendar.CalendarScope)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate config from JSON credential: %v", err)
	}
	// Since apparently service accounts don't have any associated quotas in GSuite,
	// we must supply a user to charge quota against, and I think they need to have
	// admin permission on the G Suite account to work.
	jwtConfig.Subject = *calendarID
	ctx := context.Background()

	if *record {
		r, err := httpreplay.NewRecorder(*outPath, []byte{})
		if err != nil {
			return nil, nil, fmt.Errorf("error intializing recorder: %v", err)
		}

		// Can't use `option.WithHTTPClient` here because the library throws an error when it already has a client.
		client, err := r.Client(ctx, option.WithTokenSource(jwtConfig.TokenSource(ctx)))
		if err != nil {
			return nil, nil, fmt.Errorf("error creating recorder client: %v", err)
		}
		return client, r, nil
	}

	return jwtConfig.Client(ctx), nil, nil
}

func validateFlags() error {
	if *schedPath == "" {
		return fmt.Errorf("--schedule flag is required and must not be empty")
	} else if info, err := os.Stat(*schedPath); os.IsNotExist(err) {
		return fmt.Errorf("schedule file (%v) not found: %v", *schedPath, err)
	} else if info.IsDir() {
		return fmt.Errorf("schedule must be a file, got a directory")
	}

	if *jsonKeyPath == "" {
		return fmt.Errorf("--jsonKeyPath flag is required and must not be empty")
	} else if info, err := os.Stat(*jsonKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("jsonKeyPath file (%v) not found: %v", *jsonKeyPath, err)
	} else if info.IsDir() {
		return fmt.Errorf("jsonKeyPath must be a file, got a directory")
	}

	return nil
}

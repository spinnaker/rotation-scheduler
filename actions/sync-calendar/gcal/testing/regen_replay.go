package main

import (
	"context"
	"flag"
	"github.com/ghodss/yaml"
	"github.com/spinnaker/rotation-scheduler/proto/schedule"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"

	"cloud.google.com/go/httpreplay"
	"github.com/spinnaker/rotation-scheduler/actions/sync-calendar/gcal"
)

var (
	schedulePath = flag.String("schedule", "", "Required. Path to schedule YAML file.")
	jsonKey      = flag.String("jsonKey", "", "Required. The path to the JSON key with access to the Calendar API.")
	outPath      = flag.String("out", "", "Required. File path to the recorded replay file.")

	calendarID = flag.String("calendarID", "spinbot@spinnaker.io", "Optional. "+
		"The calendar ID to update. Defaults to spinbot@google.com, because this account's calendar is not actually "+
		"used for anything, and it costs us to create additional users.")
)

// Records a real interaction with the GCP Calendar API for use in unit tests.
//
// Sample invocation from root of repo:
// go run actions/sync-calendar/gcal/testing/regen_replay.go \
//   --jsonKey ./rotation-scheduler.json \
//   --schedule ./actions/sync-calendar/gcal/testing/test_schedule.yaml \
//   --out ./actions/sync-calendar/gcal/testing/schedule.replay
func main() {
	flag.Parse()

	s, err := ioutil.ReadFile(*schedulePath)
	if err != nil {
		log.Fatalf("Error reading schedule file(%v): %v", *schedulePath, err)
	}

	sched := &schedule.Schedule{}
	if err := yaml.Unmarshal(s, sched); err != nil {
		log.Fatalf("Error unmarshalling schedule: %v", err)
	}
	stop, err := sched.EstimateStopTime()
	if err != nil {
		log.Fatalf("Error calculating last shift time.")
	}

	r, err := httpreplay.NewRecorder(*outPath, []byte{})
	if err != nil {
		log.Fatalf("Error intializing recorder: %v", err)
	}
	defer r.Close()

	key, err := ioutil.ReadFile(*jsonKey)
	if err != nil {
		log.Fatalf("unable to read JSON credential file: %v", err)
	}
	jwtConfig, err := google.JWTConfigFromJSON(key, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("unable to generate config from JSON credential: %v", err)
	}
	// Since apparently service accounts don't have any associated quotas in GSuite,
	// we must supply a user to charge quota against, and I think they need to have
	// admin permission on the G Suite account to work.
	jwtConfig.Subject = *calendarID

	ctx := context.Background()

	// Can't use `option.WithHTTPClient` here because the library throws an error when it already has a client.
	recordingClient, err := r.Client(ctx, option.WithTokenSource(jwtConfig.TokenSource(ctx)))
	if err != nil {
		log.Fatalf("Error creating recorder client: %v", err)
	}

	cal, err := gcal.NewGCal(*calendarID, recordingClient)
	if err != nil {
		log.Fatalf("Error initializing Calendar service: %v", err)
	}

	if err := cal.Clear(); err != nil {
		log.Fatalf("Error clearing calendar: %v", err)
	}

	if err := cal.Schedule(sched, stop.AddDate(0, 0, 7)); err != nil {
		log.Fatalf("Error creating new schedule: %v", err)
	}
}

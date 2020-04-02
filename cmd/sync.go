package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/httpreplay"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/spinnaker/rotation-scheduler/actions/sync-calendar/gcal"
	"github.com/spinnaker/rotation-scheduler/schedule"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var (
	syncCmd = &cobra.Command{
		Use:   "sync scheduleFilePath",
		Short: "Sync a schedule to a shared calendar",
		Args:  cobra.ExactValidArgs(1),
		RunE:  executeSync,
	}

	jsonKeyBase64 string
	calendarID    string
)

func init() {
	syncCmd.Flags().StringVarP(&jsonKeyBase64, "jsonKey", "j", "",
		"Required. A base64-encoded service account key with access to the Calendar API. "+
			"Service account must have domain-wide delegation. Create this value with something like "+
			"'cat key.json | base64 -w 0'")
	_ = syncCmd.MarkFlagRequired("jsonKey")

	syncCmd.Flags().StringVarP(&calendarID, "calendarID", "c", "spinbot@spinnaker.io",
		"Optional. The calendar ID to update. Must be a 'primary' user calendar.")

	calendarCmd.AddCommand(syncCmd)
}

func executeSync(_ *cobra.Command, args []string) error {
	schedPath := args[0]
	schedBytes, err := ioutil.ReadFile(schedPath)
	if err != nil {
		return fmt.Errorf("error reading schedule file(%v): %v", schedPath, err)
	}

	sched := &schedule.Schedule{}
	if err := yaml.Unmarshal(schedBytes, sched); err != nil {
		return fmt.Errorf("error unmarshalling schedule: %v", err)
	}

	client, closer, err := gcalHttpClient()
	if err != nil {
		return fmt.Errorf("error initializing HTTP client: %v", err)
	}
	defer func() {
		if closer != nil {
			_ = closer.Close()
		}
	}()

	cal, err := gcal.NewGCal(calendarID, client)
	if err != nil {
		return fmt.Errorf("error initializing Calendar service: %v", err)
	}

	if err := cal.Schedule(sched); err != nil {
		return fmt.Errorf("error syncing schedule: %v", err)
	}

	return nil
}

func gcalHttpClient() (*http.Client, io.Closer, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(jsonKeyBase64)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to decode JSON credential. Ensure the : %v", err)
	}
	jwtConfig, err := google.JWTConfigFromJSON(keyBytes, calendar.CalendarScope)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate config from JSON credential: %v", err)
	}
	// Since apparently service accounts don't have any associated quotas in GSuite,
	// we must supply a user to charge quota against, and I think they need to have
	// admin permission on the G Suite account to work.
	jwtConfig.Subject = calendarID
	ctx := context.Background()

	if recordFilepath == "" {
		return jwtConfig.Client(ctx), nil, nil
	}

	r, err := httpreplay.NewRecorder(recordFilepath, []byte{})
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

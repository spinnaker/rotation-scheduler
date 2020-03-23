// Package gcal handles Google Calendar integration.
package gcal

import (
	"context"
	"fmt"
	"github.com/spinnaker/rotation-scheduler/proto/schedule"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	UserAgent  = "github.com/spinnaker/rotation-scheduler"
	DateFormat = "2006-01-02"
)

// GCal wraps the Google Calendar service.
type GCal struct {
	CalendarID string
	svc        *calendar.Service
}

// NewGCal wraps the calendar specified using the client. CalendarID should be a user's primary calendar (not a shared,
// calendar), because the Clear() method only works on primary calendars.
func NewGCal(calendarID string, client *http.Client) (*GCal, error) {
	if calendarID == "" {
		return nil, fmt.Errorf("calendar ID cannot be empty")
	}

	svc, err := calendar.NewService(context.Background(), option.WithHTTPClient(client), option.WithUserAgent(UserAgent))
	if err != nil {
		return nil, fmt.Errorf("unable to create Calendar service: %v", err)
	}

	return &GCal{
		CalendarID: calendarID,
		svc:        svc,
	}, nil
}

// Client creates an authenticated HTTP Client from the specified service account credentials. The service account
// must have a Domain-wide Delegation to control user calendars. See
// https://developers.google.com/admin-sdk/directory/v1/guides/delegation
func Client(calendarID, jsonCredentialsPath string) (*http.Client, error) {
	key, err := ioutil.ReadFile(jsonCredentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read JSON credential file: %v", err)
	}
	jwtConfig, err := google.JWTConfigFromJSON(key, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("unable to generate config from JSON credential: %v", err)
	}
	// Since apparently service accounts don't have any associated quotas in GSuite,
	// we must supply a user to charge quota against, and I think they need to have
	// admin permission on the G Suite account to work.
	jwtConfig.Subject = calendarID

	return jwtConfig.Client(context.Background()), nil
}

// Clear all events from the calendar
func (g *GCal) Clear() error {
	return g.svc.Calendars.Clear(g.CalendarID).Do()
}

// Add all events from Schedule S to the calendar.
func (g *GCal) Schedule(s *schedule.Schedule, stop time.Time) error {
	if err := s.Validate(); err != nil {
		return fmt.Errorf("schedule is invalid: %v", err)
	}

	for i, shift := range s.Shifts {
		startDate, err := time.Parse(schedule.DateFormat, shift.StartDate)
		if err != nil {
			return fmt.Errorf("err parsing shift (index %v) start date (%v): %v", i, shift.StartDate, err)
		}

		var endDate time.Time
		if i == len(s.Shifts)-1 {
			endDate = stop
		} else {
			nextShift := s.Shifts[i+1]
			endDate, err = time.Parse(schedule.DateFormat, nextShift.StartDate)
			if err != nil {
				return fmt.Errorf("err parsing next shift (index %v) start date (%v): %v", i, nextShift.StartDate, err)
			}
		}

		u := shift.User
		if shift.UserOverride != "" {
			u = shift.UserOverride
		}
		event := &calendar.Event{
			Summary: fmt.Sprintf("%v Spinnaker OSS Build Cop", u),
			Start: &calendar.EventDateTime{
				Date: startDate.Format(DateFormat),
			},
			End: &calendar.EventDateTime{
				Date: endDate.Format(DateFormat),
			},
		}

		_, err = g.svc.Events.Insert(g.CalendarID, event).SendUpdates("none").Do()
		if err != nil {
			return fmt.Errorf("insert error: %v", err)
		} else {
			log.Printf("Added shift for %v from %v to %v", shift.User, event.Start.Date, event.End.Date)
		}
	}

	return nil
}
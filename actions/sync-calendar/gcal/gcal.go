// Package gcal handles Google Calendar integration.
package gcal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spinnaker/rotation-scheduler/schedule"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	UserAgent  = "github.com/spinnaker/rotation-scheduler"
	DateFormat = "2006-01-02"

	defaultEndpoint = "https://www.googleapis.com/calendar/v3/"
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

	svc, err := calendar.NewService(context.Background(),
		option.WithHTTPClient(client),
		option.WithUserAgent(UserAgent),
		option.WithEndpoint(defaultEndpoint))
	if err != nil {
		return nil, fmt.Errorf("unable to create Calendar service: %v", err)
	}

	return &GCal{
		CalendarID: calendarID,
		svc:        svc,
	}, nil
}

// Clears all events on the calendar and replaces them with events from Schedule sched.
func (g *GCal) Schedule(sched *schedule.Schedule) error {
	if err := sched.Validate(); err != nil {
		return fmt.Errorf("schedule is invalid: %v", err)
	}

	internalEvents := internalEvents(sched)

	// Clear all events from the calendar
	if err := g.svc.Calendars.Clear(g.CalendarID).Do(); err != nil {
		return fmt.Errorf("error clearing calendar: %v", err)
	}

	for i, ie := range internalEvents {
		_, err := g.svc.Events.Insert(g.CalendarID, ie.GcalEvent).SendUpdates("none").Do()
		if err != nil {
			return fmt.Errorf("insert error with event at index %v: %v\nEvent value:\n%+v", i, err, ie.GcalEvent)
		} else {
			log.Printf("Added shift for %v from %v to %v", ie.User, ie.GcalEvent.Start.Date, ie.StopDateIncl.Format(DateFormat))
		}
	}

	return nil
}

type internalEvent struct {
	GcalEvent    *calendar.Event
	User         string
	StopDateIncl time.Time
}

func internalEvents(sched *schedule.Schedule) []*internalEvent {
	intEvents := make([]*internalEvent, len(sched.Shifts))
	for i, shift := range sched.Shifts {
		var stopDateIncl, stopDateExcl time.Time
		if shift == sched.LastShift() {
			stopDateIncl = shift.StopDate
			stopDateExcl = shift.StopDateExclusive()
		} else {
			nextShift := sched.Shifts[i+1]
			stopDateIncl = nextShift.StartDateExclusive()
			stopDateExcl = nextShift.StartDate
		}

		u := shift.User
		if shift.UserOverride != "" {
			u = shift.UserOverride
		}
		event := &calendar.Event{
			Summary: eventSummary(u),
			Start: &calendar.EventDateTime{
				Date: shift.StartDate.Format(DateFormat), // Start.Date is inclusive.
			},
			End: &calendar.EventDateTime{
				Date: stopDateExcl.Format(DateFormat), // End.Date is exclusive
			},
		}
		intEvents[i] = &internalEvent{
			GcalEvent:    event,
			User:         u,
			StopDateIncl: stopDateIncl,
		}
	}

	return intEvents
}

func eventSummary(user string) string {
	return fmt.Sprintf("%v Spinnaker OSS Build Cop", user)
}

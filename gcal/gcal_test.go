package gcal

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/httpreplay"
	"github.com/ghodss/yaml"
	"github.com/spinnaker/rotation-scheduler/schedule"
	"google.golang.org/api/calendar/v3"
)

const (
	testCalendarID = "spinbot@spinnaker.io"
)

// schedule.replay file generated with:
// JSON_KEY=$(cat rotation-scheduler.json | base64 -w 0)
// go run rotation.go calendar sync --record ./gcal/testing/schedule.replay --jsonKey $JSON_KEY ./gcal/testing/test_schedule.yaml
func TestSchedule(t *testing.T) {
	replayer, err := httpreplay.NewReplayer("testing/schedule.replay")
	if err != nil {
		t.Fatalf("cannot initializer HTTP replayer: %v", err)
	}

	client, err := replayer.Client(context.Background())
	if err != nil {
		t.Fatalf("cannot initialize client from replayer: %v", err)
	}

	gcalUnderTest, err := NewGCal(testCalendarID, client)
	if err != nil {
		t.Fatalf("cannot create new gcal: %v", err)
	}

	testSchedBytes, err := ioutil.ReadFile("testing/test_schedule.yaml")
	if err != nil {
		t.Fatalf("cannot read test schedule: %v", err)
	}

	testSched := &schedule.Schedule{}
	err = yaml.Unmarshal(testSchedBytes, testSched)
	if err != nil {
		t.Fatalf("cannot read schedule from yaml file: %v", err)
	}

	err = gcalUnderTest.Schedule(testSched)
	if err != nil {
		t.Errorf("error during scheduling: %v", err)
	}
}

func TestInternalEvent(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		schedule *schedule.Schedule
		want     []*internalEvent
	}{
		{
			desc: "single day separation",
			schedule: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
					{
						StartDate:    time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						StopDate:     time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:         "second",
						UserOverride: "third",
					},
				},
			},
			want: []*internalEvent{
				{
					GcalEvent: &calendar.Event{
						Summary: eventSummary("first"),
						Start: &calendar.EventDateTime{
							Date: "2020-01-01",
						},
						End: &calendar.EventDateTime{
							Date: "2020-01-02",
						},
					},
					User:         "first",
					StopDateIncl: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					GcalEvent: &calendar.Event{
						Summary: eventSummary("third"),
						Start: &calendar.EventDateTime{
							Date: "2020-01-02",
						},
						End: &calendar.EventDateTime{
							Date: "2020-01-03",
						},
					},
					User:         "third",
					StopDateIncl: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			desc: "multi day separation",
			schedule: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
					{
						StartDate: time.Date(2020, 1, 10, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 19, 0, 0, 0, 0, time.UTC),
						User:      "second",
					},
				},
			},
			want: []*internalEvent{
				{
					GcalEvent: &calendar.Event{
						Summary: eventSummary("first"),
						Start: &calendar.EventDateTime{
							Date: "2020-01-01",
						},
						End: &calendar.EventDateTime{
							Date: "2020-01-10",
						},
					},
					User:         "first",
					StopDateIncl: time.Date(2020, 1, 9, 0, 0, 0, 0, time.UTC),
				},
				{
					GcalEvent: &calendar.Event{
						Summary: eventSummary("second"),
						Start: &calendar.EventDateTime{
							Date: "2020-01-10",
						},
						End: &calendar.EventDateTime{
							Date: "2020-01-20",
						},
					},
					User:         "second",
					StopDateIncl: time.Date(2020, 1, 19, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got := internalEvents(tc.schedule)

			if !reflect.DeepEqual(got, tc.want) {
				toStr := func(intEvents []*internalEvent) string {
					b, _ := yaml.Marshal(intEvents)
					return string(b)
				}
				t.Errorf("internal events do not match: want\n%v\n\ngot\n%v", toStr(tc.want), toStr(got))
			}
		})
	}
}

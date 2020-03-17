package scheduler

import (
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/proto"
	"github.com/spinnaker/rotation-scheduler/actions/generate-schedule/users"
	"github.com/spinnaker/rotation-scheduler/proto/schedule"
	"testing"
	"time"
)

func TestNewScheduler(t *testing.T) {
	if _, err := NewScheduler(nil, 1); err == nil {
		t.Error("want error on nil userSource and didn't get one.")
	}

	if _, err := NewScheduler(users.NewStaticSource("foo"), 0); err == nil {
		t.Errorf("want error on invalid shift duration, and didn't get one.")
	}
}

func TestSchedule(t *testing.T) {
	s, err := NewScheduler(users.NewStaticSource("first", "second", "third"), 1)
	if err != nil {
		t.Fatalf("error creating scheduler: %v", err)
	}

	sched1 := &schedule.Schedule{
		Shifts: []*schedule.Shift{
			{
				StartDate: "Wed 01 Jan 2020",
				User:      "first",
			},
			{
				StartDate: "Thu 02 Jan 2020",
				User:      "second",
			},
			{
				StartDate: "Fri 03 Jan 2020",
				User:      "third",
			},
		},
	}

	for _, tc := range []struct {
		desc    string
		start   time.Time
		stop    time.Time
		wantErr bool
		want    *schedule.Schedule
	}{
		{
			desc:    "happy path",
			start:   time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC),
			stop:    time.Date(2020, 01, 04, 0, 0, 0, 0, time.UTC),
			wantErr: false,
			want:    sched1,
		},
		{
			desc:    "start after stop",
			start:   time.Date(2020, 01, 04, 0, 0, 0, 0, time.UTC),
			stop:    time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC),
			wantErr: false,
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: "Sat 04 Jan 2020",
						User:      "first",
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := s.Schedule(tc.start, tc.stop)
			if tc.wantErr && err == nil {
				t.Errorf("err expected and not received.")
			} else if !tc.wantErr && err != nil {
				t.Errorf("got error from Schedule: %v:", err)
			} else if tc.wantErr {
				// Successfully invoked error condition
				return
			}

			if !proto.Equal(tc.want, got) {
				wantStr := scheduleToString(tc.want, t)
				gotStr := scheduleToString(got, t)
				t.Errorf("got schedule different from expected.\nWant:\n%v\n\nGot:\n%v\n", wantStr, gotStr)
			}
		})
	}
}

func TestExtendSchedule(t *testing.T) {
	s, err := NewScheduler(users.NewStaticSource("first", "second", "third"), 1)
	if err != nil {
		t.Fatalf("error creating scheduler: %v", err)
	}

	prevSchedule := &schedule.Schedule{
		Shifts: []*schedule.Shift{
			{
				StartDate: "Wed 01 Jan 2020",
				User:      "first",
			},
			{
				StartDate: "Thu 02 Jan 2020",
				User:      "second",
			},
		},
	}

	newSchedule := &schedule.Schedule{
		Shifts: []*schedule.Shift{
			{
				StartDate: "Wed 01 Jan 2020",
				User:      "first",
			},
			{
				StartDate: "Thu 02 Jan 2020",
				User:      "second",
			},
			{
				StartDate: "Fri 03 Jan 2020",
				User:      "third",
			},
			{
				StartDate: "Sat 04 Jan 2020",
				User:      "first",
			},
		},
	}

	for _, tc := range []struct {
		desc     string
		previous *schedule.Schedule
		stop     time.Time
		wantErr  bool
		want     *schedule.Schedule
	}{
		{
			desc:     "happy path",
			previous: prevSchedule,
			stop:     time.Date(2020, 01, 05, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
			want:     newSchedule,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := s.ExtendSchedule(tc.previous, tc.stop)
			if tc.wantErr && err == nil {
				t.Errorf("err expected and not received.")
			} else if !tc.wantErr && err != nil {
				t.Errorf("got error from Schedule: %v:", err)
			} else if tc.wantErr {
				// Successfully invoked error condition
				return
			}

			if !proto.Equal(tc.want, got) {
				wantStr := scheduleToString(tc.want, t)
				gotStr := scheduleToString(got, t)
				t.Errorf("got schedule different from expected.\nWant:\n%v\n\nGot:\n%v\n", wantStr, gotStr)
			}
		})
	}
}

func TestValidatePreviousShifts(t *testing.T) {
	shifts := []*schedule.Shift{
		{
			StartDate: "foobar",
		},
	}

	if err := validatePreviousShifts(shifts); err == nil {
		t.Errorf("expected error from invalid date")
	}

	shifts[0].StartDate = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(DateFormat)

	if err := validatePreviousShifts(shifts); err == nil {
		t.Errorf("expected error from missing user")
	}
}

func scheduleToString(s *schedule.Schedule, t *testing.T) string {
	j, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("json marshalling: %v", err)
	}

	b, err := yaml.JSONToYAML(j)
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}

	return string(b)
}

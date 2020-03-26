package scheduler

import (
	"reflect"
	"testing"
	"time"

	"github.com/spinnaker/rotation-scheduler/actions/generate-schedule/users"
	"github.com/spinnaker/rotation-scheduler/schedule"
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

	for _, tc := range []struct {
		desc    string
		start   time.Time
		stop    time.Time
		wantErr bool
		want    *schedule.Schedule
	}{
		{
			desc:  "happy path",
			start: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			stop:  time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "second",
					},
					{
						StartDate: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:      "third",
					},
				},
			},
		},
		{
			desc:    "start after stop",
			start:   time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
			stop:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr: true,
		},
		{
			desc:    "zero start",
			start:   time.Time{},
			stop:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr: true,
		}, {
			desc:    "zero stop",
			start:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			stop:    time.Time{},
			wantErr: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := s.Schedule(tc.start, tc.stop)
			if tc.wantErr && err == nil {
				t.Errorf("err expected and not received.")
				return
			} else if !tc.wantErr && err != nil {
				t.Errorf("got error from Schedule: %v:", err)
				return
			} else if tc.wantErr {
				// Successfully invoked error condition
				return
			}

			if reflect.DeepEqual(tc.want, got) {
				t.Errorf("got schedule different from expected.\nWant:\n%v\n\nGot:\n%v\n", tc.want, got)
			}
		})
	}
}

func TestExtendSchedule(t *testing.T) {
	for _, tc := range []struct {
		desc         string
		input        *schedule.Schedule
		durationDays int
		stop         time.Time
		wantErr      bool
		want         *schedule.Schedule
	}{
		{
			desc: "single day extension",
			input: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "second",
					},
				},
			},
			durationDays: 1,
			stop:         time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
			wantErr:      false,
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "second",
					},
					{
						StartDate: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:      "third",
					},
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
					{
						StartDate: time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
						User:      "second",
					},
				},
			},
		},
		{
			desc: "multi-day extension",
			input: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
					{
						StartDate: time.Date(2020, 1, 8, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 14, 0, 0, 0, 0, time.UTC),
						User:      "second",
					},
				},
			},
			durationDays: 7,
			stop:         time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
			wantErr:      false,
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
					{
						StartDate: time.Date(2020, 1, 8, 0, 0, 0, 0, time.UTC),
						User:      "second",
					},
					{
						StartDate: time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
						User:      "third",
					},
					{
						StartDate: time.Date(2020, 1, 22, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 28, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewScheduler(users.NewStaticSource("first", "second", "third"), tc.durationDays)
			if err != nil {
				t.Fatalf("error creating scheduler: %v", err)
			}

			err = s.ExtendSchedule(tc.input, tc.stop)
			if tc.wantErr && err == nil {
				t.Errorf("err expected and not received.")
				return
			} else if !tc.wantErr && err != nil {
				t.Errorf("got error from Schedule: %v:", err)
				return
			} else if tc.wantErr {
				// Successfully invoked error condition
				return
			}

			if !reflect.DeepEqual(tc.want, tc.input) {
				t.Errorf("got schedule different from expected.\nWant:\n%v\n\nGot:\n%v\n", tc.want, tc.input)
			}
		})
	}
}

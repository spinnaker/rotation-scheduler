package scheduler

import (
	"reflect"
	"testing"
	"time"

	"github.com/spinnaker/rotation-scheduler/schedule"
	"github.com/spinnaker/rotation-scheduler/users"
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
	for _, tc := range []struct {
		desc          string
		shiftDuration int
		start         time.Time
		stop          time.Time
		wantErr       bool
		want          *schedule.Schedule
	}{
		{
			desc:          "same day start and stop",
			shiftDuration: 1,
			start:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			stop:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
				},
			},
		},
		{
			desc:          "same day start and stop, longer shift duration",
			shiftDuration: 7,
			start:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			stop:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
		},
		{
			desc:          "happy path",
			shiftDuration: 1,
			start:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			stop:          time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
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
						StopDate:  time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
				},
			},
		},
		{
			desc:          "start after stop",
			shiftDuration: 1,
			start:         time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
			stop:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
		},
		{
			desc:          "zero start",
			shiftDuration: 1,
			start:         time.Time{},
			stop:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
		}, {
			desc:          "zero stop",
			shiftDuration: 1,
			start:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			stop:          time.Time{},
			wantErr:       true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewScheduler(users.NewStaticSource("first", "second", "third"), tc.shiftDuration)
			if err != nil {
				t.Fatalf("error creating scheduler: %v", err)
			}

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

			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("got schedule different from expected.\nWant:\n%v\n\nGot:\n%v\n", tc.want, got)
			}
		})
	}
}

func TestExtendSchedule(t *testing.T) {
	for _, tc := range []struct {
		desc         string
		input        *schedule.Schedule
		users        []string
		durationDays int
		stop         time.Time
		prune        bool
		today        time.Time
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
			users:        []string{"first", "second", "third"},
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
			users:        []string{"first", "second", "third"},
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
		{
			desc: "last shift user no longer in rotation",
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
			users:        []string{"first", "third"},
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
						User:      "third",
					},
				},
			},
		},
		{
			desc: "prune completely old schedule",
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
			users:        []string{"first", "second"},
			durationDays: 1,
			prune:        true,
			today:        time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
			stop:         time.Date(2020, 1, 7, 0, 0, 0, 0, time.UTC),
			wantErr:      false,
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
					{
						StartDate: time.Date(2020, 1, 6, 0, 0, 0, 0, time.UTC),
						User:      "second",
					},
					{
						StartDate: time.Date(2020, 1, 7, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 7, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
				},
			},
		},
		{
			desc: "prune & reschedule",
			input: &schedule.Schedule{
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
			users:        []string{"first", "second"},
			durationDays: 1,
			stop:         time.Date(2020, 1, 7, 0, 0, 0, 0, time.UTC),
			prune:        true,
			today:        time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
			wantErr:      false,
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						User:      "second",
					},
					{
						StartDate: time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
					{
						StartDate: time.Date(2020, 1, 6, 0, 0, 0, 0, time.UTC),
						User:      "second",
					},
					{
						StartDate: time.Date(2020, 1, 7, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 7, 0, 0, 0, 0, time.UTC),
						User:      "first",
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewScheduler(users.NewStaticSource(tc.users...), tc.durationDays)
			if err != nil {
				t.Fatalf("error creating scheduler: %v", err)
			}

			if tc.prune {
				today = func() time.Time { return tc.today }
			}

			err = s.ExtendSchedule(tc.input, tc.stop, tc.prune)
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

func TestPruneOldSchedules(t *testing.T) {
	for _, tc := range []struct {
		desc       string
		pruneStart time.Time
		users      []string
		sched      *schedule.Schedule
		want       *schedule.Schedule
	}{
		{
			desc:       "zero start (no change expected)",
			pruneStart: time.Time{},
			users:      []string{"foo", "bar"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Now().Truncate(24 * time.Hour),
						StopDate:  time.Now().Truncate(24 * time.Hour),
						User:      "foobar",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Now().Truncate(24 * time.Hour),
						StopDate:  time.Now().Truncate(24 * time.Hour),
						User:      "foobar",
					},
				},
			},
		},
		{
			desc:       "single shift, pruneStart before shift (no change expected)",
			pruneStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			users:      []string{"foo", "bar"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
				},
			},
		},
		{
			desc:       "single shift, pruneStart same as stop",
			pruneStart: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			users:      []string{"foo", "bar"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
				},
			},
		},
		{
			desc:       "single shift, pruneStart after stop (completely old schedule)",
			pruneStart: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			users:      []string{"foo", "bar"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
		},
		{
			desc:       "multi shift, pruneStart before first shift (no change expected)",
			pruneStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			users:      []string{"foo", "bar"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
		},
		{
			desc:       "multi shift, pruneStart same as first shift (no change expected)",
			pruneStart: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			users:      []string{"foo", "bar"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
		},
		{
			desc:       "multi shift, pruneStart between shifts (1) (no change expected)",
			pruneStart: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
			users:      []string{"foo", "bar"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
		},
		{
			desc:       "multi shift, pruneStart between shifts (2)",
			pruneStart: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
			users:      []string{"foo", "bar"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
		},
		{
			desc:       "multi shift, pruneStart same as last shift stop",
			pruneStart: time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
			users:      []string{"foo", "bar"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
		},
		{
			desc:       "multi shift, pruneStart after last shift stop",
			pruneStart: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
			users:      []string{"foo", "bar"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewScheduler(users.NewStaticSource(tc.users...), 1)
			if err != nil {
				t.Fatalf("error creating scheduler: %v", err)
			}

			if err := tc.sched.Validate(); err != nil {
				t.Fatalf("invalid schedule beforehand: %v", err)
			}
			s.pruneOldShifts(tc.pruneStart, tc.sched)
			if err := tc.sched.Validate(); err != nil {
				t.Errorf("invalid pruned schedule: %v", err)
			}

			if !reflect.DeepEqual(tc.want, tc.sched) {
				t.Errorf("got schedule different from expected.\nWant:\n%v\n\nGot:\n%v\n", tc.want, tc.sched)
			}
		})
	}
}

func TestPruneNotFoundUsers(t *testing.T) {
	for _, tc := range []struct {
		desc  string
		users []string
		sched *schedule.Schedule
		want  *schedule.Schedule
	}{
		{
			desc:  "first user not in rotation",
			users: []string{"bar"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
				},
			},
		},
		{
			desc:  "second user not in rotation",
			users: []string{"foo"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:      "bar",
					},
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
				},
			},
		},
		{
			desc:  "override masks missing user",
			users: []string{"foo"},
			sched: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate:    time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:         "bar",
						UserOverride: "foo",
					},
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
				},
			},
			want: &schedule.Schedule{
				Shifts: []*schedule.Shift{
					{
						StartDate: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
					{
						StartDate:    time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
						User:         "bar",
						UserOverride: "foo",
					},
					{
						StartDate: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
						User:      "foo",
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewScheduler(users.NewStaticSource(tc.users...), 1)
			if err != nil {
				t.Fatalf("error creating scheduler: %v", err)
			}

			if err := tc.sched.Validate(); err != nil {
				t.Fatalf("invalid schedule beforehand: %v", err)
			}

			s.pruneNotFoundUsers(tc.sched)
			if err := tc.sched.Validate(); err != nil {
				t.Errorf("invalid pruned schedule: %v", err)
			}

			if !reflect.DeepEqual(tc.want, tc.sched) {
				t.Errorf("got schedule different from expected.\nWant:\n%v\n\nGot:\n%v\n", tc.want, tc.sched)
			}
		})
	}
}

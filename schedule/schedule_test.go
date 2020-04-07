package schedule

import (
	"reflect"
	"testing"
	"time"

	"github.com/ghodss/yaml"
)

func TestScheduleValidate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		schedule *Schedule
		wantErr  bool
	}{
		{
			desc:     "nil schedule",
			schedule: nil,
			wantErr:  true,
		},
		{
			desc:     "empty schedule",
			schedule: &Schedule{},
			wantErr:  true,
		},
		{
			desc: "empty shifts",
			schedule: &Schedule{
				Shifts: []*Shift{},
			},
			wantErr: true,
		},
		{
			desc: "invalid shift",
			schedule: &Schedule{
				Shifts: []*Shift{
					{}, // shift won't validate
				},
			},
			wantErr: true,
		},
		{
			desc: "shifts out of order",
			schedule: &Schedule{
				Shifts: []*Shift{
					{
						User:      "bar",
						StartDate: time.Date(2020, 6, 8, 0, 0, 0, 0, time.UTC),
					},
					{
						User:      "foo",
						StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 6, 13, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			wantErr: true,
		},
		{
			desc: "stop on non-last shift",
			schedule: &Schedule{
				Shifts: []*Shift{
					{
						User:      "bar",
						StartDate: time.Date(2020, 6, 8, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 6, 13, 0, 0, 0, 0, time.UTC),
					},
					{
						User:      "foo",
						StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			wantErr: true,
		},
		{
			desc: "missing stop on last shift",
			schedule: &Schedule{
				Shifts: []*Shift{
					{
						User:      "foo",
						StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			wantErr: true,
		},
		{
			desc: "valid schedule",
			schedule: &Schedule{
				Shifts: []*Shift{
					{
						User:      "foo",
						StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						User:      "bar",
						StartDate: time.Date(2020, 6, 8, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 6, 14, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.schedule.Validate()

			if tc.wantErr && err == nil {
				t.Errorf("err expected and not received.")
				return
			} else if !tc.wantErr && err != nil {
				t.Errorf("got unexpected error: %v:", err)
				return
			} else if tc.wantErr {
				// Successfully invoked error condition
				return
			}
		})
	}
}

func TestMarshalSchedule(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		schedule *Schedule
		want     string
	}{
		{
			desc: "valid",
			schedule: &Schedule{
				Shifts: []*Shift{
					{
						User:         "foo",
						UserOverride: "baz",
						StartDate:    time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						User:      "bar",
						StartDate: time.Date(2020, 6, 8, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 6, 14, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: `shifts:
- startDate: Mon 01 Jun 2020
  user: foo
  userOverride: baz
- startDate: Mon 08 Jun 2020
  stopDate: Sun 14 Jun 2020
  user: bar
`,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := yaml.Marshal(tc.schedule)
			if err != nil {
				t.Errorf("marshal schedule error: %v", err)
				return
			}

			gotString := string(got)
			if tc.want != gotString {
				t.Errorf("want:\n%v\n\ngot:\n%v", tc.want, gotString)
			}
		})
	}
}

func TestUnmarshalSchedule(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		schedule string
		want     *Schedule
	}{
		{
			desc: "valid",
			schedule: `shifts:
- startDate: Mon 01 Jun 2020
  user: foo
  userOverride: baz
- startDate: Mon 08 Jun 2020
  stopDate: Sun 14 Jun 2020
  user: bar
`,
			want: &Schedule{
				Shifts: []*Shift{
					{
						User:         "foo",
						UserOverride: "baz",
						StartDate:    time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						User:      "bar",
						StartDate: time.Date(2020, 6, 8, 0, 0, 0, 0, time.UTC),
						StopDate:  time.Date(2020, 6, 14, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got := &Schedule{}
			err := yaml.Unmarshal([]byte(tc.schedule), got)
			if err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			if !reflect.DeepEqual(*tc.want, *got) {
				t.Errorf("schedules are not deep equal. want:\n%v\n\ngot:%v", tc.want, got)
			}
		})
	}
}

func TestMarshalShift(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		shift   *Shift
		wantErr bool
		want    string
	}{
		{
			desc: "valid",
			shift: &Shift{
				StartDate:    time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
				StopDate:     time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
				User:         "foo",
				UserOverride: "bar",
			},
			want: `startDate: Mon 01 Jun 2020
stopDate: Mon 01 Jun 2020
user: foo
userOverride: bar
`,
		},
		{
			desc: "optional fields omitted",
			shift: &Shift{
				StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
				User:      "foo",
			},
			want: `startDate: Mon 01 Jun 2020
user: foo
`,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := yaml.Marshal(tc.shift)
			if err != nil {
				t.Errorf("marshal shift error: %v", err)
				return
			}

			gotString := string(got)
			if tc.want != gotString {
				t.Errorf("want:\n%v\n\ngot:\n%v", tc.want, gotString)
			}
		})
	}
}

func TestUnmarshalShift(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		shift   string
		wantErr bool
		want    Shift
	}{
		{
			desc: "valid",
			shift: `startDate: Mon 01 Jun 2020
stopDate: Mon 01 Jun 2020
user: foo
userOverride: bar
`,
			want: Shift{
				StartDate:    time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
				StopDate:     time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
				User:         "foo",
				UserOverride: "bar",
			},
		},
		{
			desc: "optional fields omitted",
			shift: `startDate: Mon 01 Jun 2020
user: foo
`,
			want: Shift{
				StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
				User:      "foo",
			},
		},
		{
			desc:    "invalid start date",
			wantErr: true,
			shift: `startDate: Monday 01 Jun 2020
user: foo
`,
		},
		{
			desc:    "invalid stop date",
			wantErr: true,
			shift: `startDate: Mon 01 Jun 2020
stopDate: Monday 01 Jun 2020
user: foo
`,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got := &Shift{}
			err := yaml.Unmarshal([]byte(tc.shift), got)

			if tc.wantErr && err == nil {
				t.Errorf("err expected and not received.")
				return
			} else if !tc.wantErr && err != nil {
				t.Errorf("got unexpected error: %v:", err)
				return
			} else if tc.wantErr {
				// Successfully invoked error condition
				return
			}

			if tc.want != *got {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestShiftValidate(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		shift   *Shift
		wantErr bool
	}{
		{
			desc:    "nil shift",
			shift:   nil,
			wantErr: true,
		},
		{
			desc:    "empty shift",
			shift:   &Shift{},
			wantErr: true,
		},
		{
			desc: "no user",
			shift: &Shift{
				StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			desc: "no startDate",
			shift: &Shift{
				User: "foo",
			},
			wantErr: true,
		},
		{
			desc: "stop before start",
			shift: &Shift{
				User:      "foo",
				StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
				StopDate:  time.Date(2020, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			desc: "valid shift",
			shift: &Shift{
				StartDate:    time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
				StopDate:     time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
				User:         "foo",
				UserOverride: "bar",
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.shift.Validate()

			if tc.wantErr && err == nil {
				t.Errorf("err expected and not received.")
				return
			} else if !tc.wantErr && err != nil {
				t.Errorf("got unexpected error: %v:", err)
				return
			} else if tc.wantErr {
				// Successfully invoked error condition
				return
			}
		})
	}
}

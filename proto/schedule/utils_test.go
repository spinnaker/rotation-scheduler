package schedule

import (
	"testing"
	"time"
)

func TestEstimateStopTime(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		sched   *Schedule
		wantErr bool
		want    time.Time
	}{
		{
			desc:    "nil schedule",
			wantErr: true,
		},
		{
			desc:    "nil shifts",
			sched:   &Schedule{},
			wantErr: true,
		},
		{
			desc: "empty shifts",
			sched: &Schedule{
				Shifts: []*Shift{},
			},
			wantErr: true,
		},
		{
			desc: "single shift",
			sched: &Schedule{
				Shifts: []*Shift{
					{
						StartDate: "Wed 01 Jan 2020",
					},
				},
			},
			want: time.Date(2020, 1, 1+defaultStopTimeDays, 0, 0, 0, 0, time.UTC),
		},
		{
			desc: "2+ shifts",
			sched: &Schedule{
				Shifts: []*Shift{
					{
						StartDate: "Wed 01 Jan 2020",
					},
					{
						StartDate: "Wed 03 Jan 2020", // 2 day diff
					},
				},
			},
			want: time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.sched.EstimateStopTime()
			if tc.wantErr && err == nil {
				t.Errorf("err expected and not received.")
			} else if !tc.wantErr && err != nil {
				t.Errorf("got error from EstimateStopTime: %v:", err)
			} else if tc.wantErr {
				// Successfully invoked error condition
				return
			}

			if got != tc.want {
				t.Errorf("wanted stop time %v, got %v", tc.want, got)
			}
		})
	}
}

func TestValidatePreviousShifts(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		sched   *Schedule
		wantErr bool
	}{
		{
			desc:    "nil s",
			wantErr: true,
		},
		{
			desc:  "nil shifts",
			sched: &Schedule{},
		},
		{
			desc: "empty shifts",
			sched: &Schedule{
				Shifts: []*Shift{},
			},
		},
		{
			desc: "bogus Start",
			sched: &Schedule{
				Shifts: []*Shift{
					{
						StartDate: "foobar",
					},
				},
			},
			wantErr: true,
		},
		{
			desc: "missing user",
			sched: &Schedule{
				Shifts: []*Shift{
					{
						StartDate: "Wed 01 Jan 2020",
					},
				},
			},
			wantErr: true,
		},
		{
			desc: "out of order",
			sched: &Schedule{
				Shifts: []*Shift{
					{
						StartDate: "Thu 02 Jan 2020",
					},
					{
						StartDate: "Wed 01 Jan 2020",
					},
				},
			},
			wantErr: true,
		},
		{
			desc: "all good",
			sched: &Schedule{
				Shifts: []*Shift{
					{
						StartDate: "Wed 01 Jan 2020",
						User:      "Foobar",
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.sched.Validate()
			if tc.wantErr && err == nil {
				t.Errorf("err expected and not received.")
			} else if !tc.wantErr && err != nil {
				t.Errorf("got error from validate: %v:", err)
			}
		})
	}
}

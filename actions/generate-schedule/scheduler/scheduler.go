package scheduler

import (
	"fmt"
	"github.com/spinnaker/rotation-scheduler/actions/generate-schedule/users"
	"github.com/spinnaker/rotation-scheduler/proto/schedule"
	"time"
)

// Scheduler creates or extends a new rotation schedule.
type Scheduler struct {
	userSource        users.Source
	shiftDurationDays int
}

// NewScheduler creates a new Scheduler. All args are required.
func NewScheduler(userSource users.Source, shiftDurationDays int) (*Scheduler, error) {
	if userSource == nil {
		return nil, fmt.Errorf("no user source specificed")
	}

	if shiftDurationDays <= 0 {
		return nil, fmt.Errorf("shift duration invalid. Must be greater than 0, was: %v", shiftDurationDays)
	}

	return &Scheduler{
		userSource:        userSource,
		shiftDurationDays: shiftDurationDays,
	}, nil
}

// Schedule creates a new Schedule from start (inclusive) to stop (exclusive). Will return a 1-entry schedule if
// stop is before start.
func (s *Scheduler) Schedule(start, stop time.Time) (*schedule.Schedule, error) {
	sched := &schedule.Schedule{
		Shifts: []*schedule.Shift{
			{
				StartDate: start.Format(schedule.DateFormat),
				User:      s.userSource.NextUser(),
			},
		},
	}
	if err := s.ExtendSchedule(sched, stop); err != nil {
		return nil, err
	}
	return sched, nil
}

// ExtendSchedule takes a previously generated schedule and extends it. The user rotation continues as normal from the
// last shift in the schedule.
func (s *Scheduler) ExtendSchedule(sched *schedule.Schedule, stop time.Time) error {
	mostRecentShift := sched.Shifts[len(sched.Shifts)-1]
	if err := s.userSource.StartAfter(mostRecentShift.User); err != nil {
		return fmt.Errorf("error finding input shift owner (%v) in user source: %v", mostRecentShift.User, err)
	}

	for start := s.startTime(mostRecentShift); start.Before(stop); start = s.nextShiftTime(start) {
		sched.Shifts = append(sched.Shifts, &schedule.Shift{
			User:      s.userSource.NextUser(),
			StartDate: start.Format(schedule.DateFormat),
		})
	}

	return nil
}

func (s *Scheduler) startTime(mostRecentShift *schedule.Shift) time.Time {
	start, err := time.Parse(schedule.DateFormat, mostRecentShift.StartDate)
	if err != nil {
		return time.Now() // Should never happen with validation.
	}

	return s.nextShiftTime(start)
}

func (s *Scheduler) nextShiftTime(previous time.Time) time.Time {
	return previous.AddDate(0, 0, s.shiftDurationDays)
}

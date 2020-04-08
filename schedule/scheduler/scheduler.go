package scheduler

import (
	"fmt"
	"time"

	"github.com/spinnaker/rotation-scheduler/schedule"
	"github.com/spinnaker/rotation-scheduler/users"
)

const (
	DateFormat = "Mon 02 Jan 2006"
)

var (
	today = func() time.Time { return time.Now().Truncate(24 * time.Hour) }
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

// Schedule creates a new Schedule that includes whole shifts of `Scheduler.shiftDuration` from start (inclusive) to
// stop (inclusive).  Will return an error if stop is before start, or either start are stop are zero values.
func (s *Scheduler) Schedule(start, stop time.Time) (*schedule.Schedule, error) {
	if start.IsZero() || stop.IsZero() {
		return nil, fmt.Errorf("neither start (%v) nor stop (%v) can be zero values", start, stop)
	}

	if stop.Before(start) {
		return nil, fmt.Errorf("start cannot be before stop")
	}

	sched := &schedule.Schedule{}
	if err := s.extendSchedule(sched, start, stop); err != nil {
		return nil, fmt.Errorf("error extending schedule: %v", err)
	}

	return sched, nil
}

// ExtendSchedule takes a previously generated schedule and extends it. The user rotation continues as normal from the
// last shift in the schedule.
//
// If prune is true:
// * all shifts completed in the past are purged, and all current and future shifts are checked for owners in the
// current rotation.
// * If a shift is assigned to a user that is no longer in the rotation (including if the missing user is in the
// userOverride field), that shift and all future shifts are rescheduled.
// * Rescheduled shifts do not carry over previous userOverride values.
// * Shifts originally assigned to a missing rotation member, but have a userOverride owner that is in the
// current rotation, are not be rescheduled.
func (s *Scheduler) ExtendSchedule(sched *schedule.Schedule, stopInclusive time.Time, prune bool) error {
	if err := sched.Validate(); err != nil {
		return fmt.Errorf("cannot extend invalid schedule: %v", err)
	}

	if prune {
		s.prune(today(), sched)
	}

	lastShift := sched.LastShift()
	if lastShift.StopDateExclusive().After(stopInclusive) {
		return fmt.Errorf("cannot stop before the last shift of the previous schedule is complete")
	}

	s.userSource.StartAfter(lastShift.User)
	firstNewShiftStart := sched.LastShift().StopDateExclusive()
	sched.LastShift().ClearStopDate()

	return s.extendSchedule(sched, firstNewShiftStart, stopInclusive)
}

func (s *Scheduler) prune(start time.Time, sched *schedule.Schedule) {
	s.pruneOldShifts(start, sched)
	s.pruneNotFoundUsers(sched)
}

func (s *Scheduler) pruneOldShifts(start time.Time, sched *schedule.Schedule) {
	for i, shift := range sched.Shifts {
		if i != 0 && start.Before(shift.StartDate) {
			// prune start time happened sometime in between the last shift and this shift.
			sched.Shifts = sched.Shifts[(i - 1):]
			break
		} else if start == shift.StartDate || (shift == sched.LastShift() && start == shift.StopDate) {
			// prune start time landed on a shift start or stop time.
			sched.Shifts = sched.Shifts[i:]
			break
		} else if shift == sched.LastShift() && start.After(shift.StopDate) {
			// entire schedule is in the past.
			s.userSource.StartAfter(sched.LastShift().GetUser())
			sched.Shifts = []*schedule.Shift{
				{
					StartDate: start,
					User:      s.userSource.NextUser(),
				},
			}
			sched.LastShift().SetStopDateExclusive(s.nextShiftTime(sched.LastShift().StartDate))
			break
		}
	}
}

// pruneNotFoundUsers truncates the schedule at the first shift where a user is no longer in the rotation group.
// The intent is to not reschedule too aggressively, so if a removed user shift has been swapped with someone else, that
// shift will not be removed.
func (s *Scheduler) pruneNotFoundUsers(sched *schedule.Schedule) {
	for i, shift := range sched.Shifts {
		if !s.userSource.Contains(shift.GetUser()) {
			sched.Shifts = sched.Shifts[:i]

			if sched.LastShift() == nil {
				s.userSource.StartAfter(shift.GetUser())
				sched.Shifts = append(sched.Shifts, &schedule.Shift{
					StartDate: shift.StartDate,
					User:      s.userSource.NextUser(),
				})
				sched.LastShift().SetStopDateExclusive(s.nextShiftTime(shift.StartDate))
			} else {
				sched.LastShift().SetStopDateExclusive(shift.StartDate)
			}
			return
		}
	}
}

func (s *Scheduler) extendSchedule(sched *schedule.Schedule, start, stopInclusive time.Time) error {
	for ; s.wholeShiftCanFit(start, stopInclusive); start = s.nextShiftTime(start) {
		sched.Shifts = append(sched.Shifts, &schedule.Shift{
			User:      s.userSource.NextUser(),
			StartDate: start,
		})
	}

	if sched.LastShift() == nil {
		return fmt.Errorf("no whole shifts of duration %v days can fit between %v and %v",
			s.shiftDurationDays,
			start.Format(DateFormat),
			stopInclusive.Format(DateFormat))
	}
	// The last value of the loop conveniently contains the exclusive stop date.
	sched.LastShift().SetStopDateExclusive(start)

	if err := sched.Validate(); err != nil {
		return fmt.Errorf("error validating new schedule: %v", err)
	}
	return nil
}

func (s *Scheduler) wholeShiftCanFit(start, stopInclusive time.Time) bool {
	shiftStopIncl := s.nextShiftTime(start).Add(-24 * time.Hour)
	return shiftStopIncl.Before(stopInclusive) || shiftStopIncl == stopInclusive
}

func (s *Scheduler) nextShiftTime(previous time.Time) time.Time {
	return previous.AddDate(0, 0, s.shiftDurationDays)
}

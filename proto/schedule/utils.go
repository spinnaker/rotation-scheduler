package schedule

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"time"
)

const (
	DateFormat                  = "Mon 02 Jan 2006"
	defaultStopTimeDurationDays = 7
)

func (s *Schedule) Validate() error {
	if s == nil {
		return fmt.Errorf("nil Schedule")
	}

	if s.Shifts == nil || len(s.Shifts) == 0 {
		fmt.Printf("No input shifts found. Starting from scratch.")
		return nil
	}

	for i, shift := range s.Shifts {
		if _, err := time.Parse(DateFormat, shift.StartDate); err != nil {
			return fmt.Errorf("error in input shift entry %v, invalid value: %v, err: %v", i, shift.StartDate, err)
		}
		if shift.User == "" {
			return fmt.Errorf("user cannot be empty in input shift entry %v", i)
		}
	}

	return nil // It's all good.
}

func (s *Schedule) FromYAMLFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	s.Reset()
	return yaml.Unmarshal(b, s)
}

// Returns the time difference between the next-to-last and last Shifts in this Schedule as an estimation for when the
// last Shift will end. Defaults to 7 when only 1 shift is in the schedule, and throws an error otherwise.
func (s *Schedule) EstimateStopTime() (time.Time, error) {
	if s == nil {
		return time.Time{}, fmt.Errorf("no last shift on nil schedule")
	}

	if s.Shifts == nil || len(s.Shifts) == 0 {
		return time.Time{}, fmt.Errorf("shifts cannot be nil or empty")
	}

	lastShift := s.Shifts[len(s.Shifts)-1]
	lastShiftTime, err := time.Parse(DateFormat, lastShift.StartDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing last shift start date(%v): %v", lastShift.StartDate, err)
	}

	if len(s.Shifts) == 1 {
		return lastShiftTime.AddDate(0, 0, defaultStopTimeDurationDays), nil
	}

	nextToLastShift := s.Shifts[len(s.Shifts)-2]
	nextToLastShiftTime, err := time.Parse(DateFormat, nextToLastShift.StartDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing next to last shift start date(%v): %v", nextToLastShift.StartDate, err)
	}

	diff := lastShiftTime.Sub(nextToLastShiftTime)
	return lastShiftTime.Add(diff), nil
}

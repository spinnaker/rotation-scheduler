package schedule

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"time"
)

const (
	DateFormat              = "Mon 02 Jan 2006"
	defaultStopTimeDays 	= 7
	dayDuration             = 24 * time.Hour
	defaultStopTimeDuration = defaultStopTimeDays * dayDuration
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

func FromYAMLFile(path string) (*Schedule, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	s := &Schedule{}
	if err := yaml.Unmarshal(b, s); err != nil {
		return nil, err
	}
	return s, nil
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
	if len(s.Shifts) == 1 {
		return lastShift.Add(defaultStopTimeDuration)
	}

	nextToLastShift := s.Shifts[len(s.Shifts)-2]
	between, err := lastShift.DurationBetween(nextToLastShift)
	if err != nil{
		return time.Time{}, fmt.Errorf("error calculating difference: %v", err)
	}

	return lastShift.Add(between)
}

func (a *Shift) Add(duration time.Duration) (time.Time, error) {
	aTime, err := time.Parse(DateFormat, a.StartDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing start date: %v", err)
	}
	return aTime.Add(duration), nil
}

func (a *Shift) DurationBetween(b *Shift) (time.Duration, error) {
	aTime, err1 := time.Parse(DateFormat, a.StartDate)
	bTime, err2 := time.Parse(DateFormat, b.StartDate)

	if err1 != nil || err2 != nil {
		return time.Duration(0), fmt.Errorf("Error parsing shift time difference\nerr1: %v\nerr2: %v", err1, err2)
	}

	return aTime.Sub(bTime), nil
}

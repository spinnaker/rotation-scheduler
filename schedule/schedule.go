package schedule

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ghodss/yaml"
)

const (
	DateFormat = "Mon 02 Jan 2006"
)

// Schedule represents a series of Shifts, in temporal order.
type Schedule struct {
	// Shifts is the list of shifts in order. The last Shift, and only the last Shift, should have a StopTime value.
	Shifts []*Shift `json:"shifts"`
}

func (sch *Schedule) LastShift() *Shift {
	if sch.Shifts == nil || len(sch.Shifts) == 0 {
		return nil
	}
	return sch.Shifts[len(sch.Shifts)-1]
}

// Validate confirms all shifts are valid individually and collectively. Returns a nil error if there are no shifts.
func (sch *Schedule) Validate() error {
	if sch == nil {
		return fmt.Errorf("schedule cannot be nil")
	}

	if sch.LastShift() == nil {
		return nil
	}

	var previousStartDate time.Time
	for i, shift := range sch.Shifts {
		if err := shift.Validate(); err != nil {
			return fmt.Errorf("schedule is invalid because shift %v is invalid: %v", i, err)
		}

		if !previousStartDate.IsZero() && shift.StartDate.Before(previousStartDate) {
			return fmt.Errorf("shifts out of order: shift time %v found after %v", previousStartDate, shift.StartDate)
		}
		previousStartDate = shift.StartDate

		if shift == sch.LastShift() {
			if shift.StopDate.IsZero() {
				return fmt.Errorf("missing stop date on last shift")
			}
		} else {
			// All other shifts should not have a stop date.
			if !shift.StopDate.IsZero() {
				return fmt.Errorf("stop date is only valid on the last shift. Found stop date %v on shift at index %v", shift.StopDate, i)
			}
		}
	}

	return nil // It's all good.
}

func (sch *Schedule) String() string {
	if sch == nil {
		return ""
	}
	b, err := yaml.Marshal(sch)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// A shift represents a span of time someone is on duty.
type Shift struct {
	// The start date, inclusive, of when the User goes on duty. Only year, month, and day field are relevant.
	StartDate time.Time `json:"startDate,omitempty"`
	User      string    `json:"user,omitempty"`

	// UserOverride should be used when Users want to swap or alter shift assignments without disturbing the normal
	// rotation cycle.
	UserOverride string `json:"userOverride,omitempty"`

	// StopDate is inclusive, and must only be used on the last Shift of a Schedule. For all other Shifts, the stop date
	// is implied by the next shift's StartDate, and this value should remain the zero `time.Time` value.
	StopDate time.Time `json:"stopDate,omitempty"`
}

// StopDateExclusive is used when an API uses (start, stop) date ranges as (inclusive, exclusive). May returns a zero
// `time.Time` when stop is not set.
func (sh *Shift) StopDateExclusive() time.Time {
	if sh.StopDate.IsZero() {
		return sh.StopDate // zero value is both inclusive and exclusive.
	}
	return sh.StopDate.Add(24 * time.Hour)
}

// StopDateExclusive can be used when an API uses (start, stop) date ranges as (inclusive, exclusive).
func (sh *Shift) SetStopDateExclusive(excl time.Time) {
	if excl.IsZero() {
		sh.StopDate = excl // zero value is both inclusive and exclusive.
	}
	sh.StopDate = excl.Add(-24 * time.Hour)
}

func (sh *Shift) ClearStopDate() {
	sh.StopDate = time.Time{}
}

func (sh *Shift) Validate() error {
	if sh == nil {
		return fmt.Errorf("shift cannot be nil")
	}

	if sh.User == "" {
		return fmt.Errorf("user cannot be empty")
	}

	if sh.StartDate.IsZero() {
		return fmt.Errorf("start date cannot be zero value")
	}

	if !sh.StopDate.IsZero() && sh.StartDate.After(sh.StopDate) {
		return fmt.Errorf("start date must be before stop date")
	}

	return nil
}

// MarshalJSON returns timestamps in the `DateFormat` format.
func (sh *Shift) MarshalJSON() ([]byte, error) {
	// Technique borrowed from http://choly.ca/post/go-json-marshalling/
	type Alias Shift

	aux := &struct {
		*Alias
		StartDate string `json:"startDate"`
		StopDate  string `json:"stopDate,omitempty"`
	}{
		Alias:     (*Alias)(sh),
		StartDate: sh.StartDate.Format(DateFormat),
	}

	if !sh.StopDate.IsZero() {
		aux.StopDate = sh.StopDate.Format(DateFormat)
	}

	return json.Marshal(aux)
}

// UnmarshalJSON reads timestamps in the `DateFormat` format, and will throw parsing error otherwise.
func (sh *Shift) UnmarshalJSON(data []byte) error {
	// Technique borrowed from http://choly.ca/post/go-json-marshalling/
	type Alias Shift
	aux := &struct {
		*Alias
		StartDate string `json:"startDate"`
		StopDate  string `json:"stopDate,omitempty"`
	}{
		Alias: (*Alias)(sh),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	if sh.StartDate, err = time.Parse(DateFormat, aux.StartDate); err != nil {
		return fmt.Errorf("erroring parsing start date: %v", err)
	}
	if aux.StopDate != "" {
		if sh.StopDate, err = time.Parse(DateFormat, aux.StopDate); err != nil {
			return fmt.Errorf("erroring parsing stop date: %v", err)
		}
	}

	return nil
}

func (sh *Shift) String() string {
	if sh == nil {
		return ""
	}
	b, err := yaml.Marshal(sh)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

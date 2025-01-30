package cmd

import (
	"fmt"
	"regexp"
	"time"
)

// HhMm is a string that represents a 24 hour clock time as hours and minutes in
// HH:MM format.
type HhMm string

// Set sets a clock time and validates that the string argument is in
// the expected 24 hour clock format.
func (hhmm *HhMm) Set(s string) error {
	if s == "" {
		return nil
	}
	ok, err := regexp.Match(`^\d\d:\d\d$`, []byte(s))
	if err != nil {
		return fmt.Errorf("%s is not in the form HH:MM, %w", s, err)
	}
	if !ok {
		return fmt.Errorf("%s is not in the form HH:MM", s)
	}
	_, err = time.Parse("2006-01-02 15:04:05", fmt.Sprintf("2006-01-02 %s:00", s))
	if err != nil {
		return fmt.Errorf("%s is not in the form HH:MM", s)
	}
	*hhmm = HhMm(s)
	return nil
}

// Type returns a string showing how a CLI should display the type.
func (hhmm *HhMm) Type() string {
	return "HH:MM"
}

func (hhmm *HhMm) String() string {
	return string(*hhmm)
}

// Until returns the time.Duration until the specified clock time.
func (hhmm *HhMm) Until() (time.Duration, error) {
	// An unspecified time is equivalent to now.
	if *hhmm == HhMm("") {
		return time.Nanosecond, nil
	}

	// This is some hackery to use Go's time functions to get how long we have to
	// wait by prepending a date to the time and then subtracting the time that
	// has elapsed since midnight on that date.
	now := time.Now()
	tm, _ := time.Parse("2006-01-02 15:04:05", fmt.Sprintf("2006-01-02 %s:00", hhmm))
	tm = tm.AddDate(now.Year()-2006, int(now.Month())-1, now.Day()-2)
	_, offset := now.Zone()
	tm = tm.Add(time.Duration(-offset) * time.Second)
	d := time.Until(tm)
	for {
		if d > 0 {
			break
		}
		// If d < 0, then wait until the next day.
		d += 24 * time.Hour
	}
	return d, nil
}

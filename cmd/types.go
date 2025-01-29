package cmd

import (
	"fmt"
	"regexp"
	"time"
)

type HhMm string

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

func (hhmm *HhMm) Type() string {
	return "HH:MM"
}

func (hhmm *HhMm) String() string {
	return string(*hhmm)
}

func (hhmm *HhMm) Until() (time.Duration, error) {
	if *hhmm == HhMm("") {
		return time.Nanosecond, nil
	}
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
		d += 24 * time.Hour
	}
	return d, nil
}

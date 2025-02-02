/*
Copyright 2025 Carl Meijer.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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
	if len(s) == 4 {
		s = "0" + s
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

	now := time.Now()
	e := strings.Split(string(*hhmm), ":")
	h, err := strconv.Atoi(e[0])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(e[1])
	if err != nil {
		return 0, err
	}
	tm := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
	d := time.Until(tm)
	if d < 0 {
		d += 24 * time.Hour
	}
	return d, nil
}

// SoC represent a battery's state of charge
type SoC int

// Set sets a battery state of charge.
func (soc *SoC) Set(s string) error {
	socInt, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if socInt < 0 || socInt > 100 {
		return fmt.Errorf("battery SoC must be in the range 0-100, not %d", socInt)
	}
	*soc = SoC(socInt)
	return nil
}

// Type returns "SoC" as the type.
func (soc *SoC) Type() string {
	return "SoC"
}

func (soc *SoC) String() string {
	if *soc < 0 {
		return ""
	}
	return fmt.Sprintf("%d", int(*soc))
}

// Int converts a SoC to an int.
func (soc *SoC) Int() int {
	return int(*soc)
}

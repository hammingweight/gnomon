package handlers

import (
	"fmt"
	"log"
	"regexp"
	"time"
)

func getDelayUntil(t string) (time.Duration, error) {
	ok, err := regexp.Match(`^\d\d:\d\d$`, []byte(t))
	if err != nil {
		return 0, fmt.Errorf("%s is not in the form HH:MM, %w", t, err)
	}
	if !ok {
		return 0, fmt.Errorf("%s is not in the form HH:MM", t)
	}
	now := time.Now()
	tm, err := time.Parse("2006-01-02 15:04:05", "2006-01-02 "+t+":00")
	fmt.Println(tm)
	if err != nil {
		return 0, fmt.Errorf("%s is not in the form HH:MM", t)
	}
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

func Execute(start string, end string) error {
	startDelay := time.Second
	var err error
	if start != "" {
		startDelay, err = getDelayUntil(start)
		if err != nil {
			return err
		}
	}
	endDelay, err := getDelayUntil(end)
	if err != nil {
		return err
	}
	if endDelay <= startDelay {
		endDelay += 24 * time.Hour
	}
	log.Printf("Waiting for %s to start\n", startDelay)
	<-time.Tick(startDelay)
	log.Println("starting")
	log.Printf("Waiting for %s to end\n", endDelay)
	<-time.Tick(endDelay)
	return nil
}

package handlers

import (
	"log"
	"time"
)

func Execute(start time.Duration, runTime time.Duration) error {
	log.Printf("Waiting for %s to start\n", start)
	<-time.Tick(start)
	log.Println("starting")
	log.Printf("Waiting for %s to end\n", runTime)
	<-time.Tick(runTime)
	log.Println("done")
	return nil
}

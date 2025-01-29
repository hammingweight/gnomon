package handlers

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/hammingweight/gnomon/rest"
)

func Execute(start time.Duration, runTime time.Duration, configFile string) error {
	log.Printf("Waiting for %s to start...\n", start)
	<-time.Tick(start)
	log.Println("Starting inverter management")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, runTime)
	defer cancel()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	displayChan := make(chan rest.State)
	go DisplayHandler(ctx, wg, displayChan)

	wg.Add(1)
	socChan := make(chan rest.State)
	go SocHandler(ctx, wg, socChan)

	fanout := Fanout(displayChan, socChan)

	rest.Poll(ctx, configFile, fanout)

	wg.Wait()
	return nil
}

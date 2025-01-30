package handlers

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/hammingweight/gnomon/rest"
)

func Execute(start time.Duration, runTime time.Duration, configFile string, ct bool) error {
	if start >= time.Minute {
		log.Printf("Waiting for %s to start...\n", start)
	}
	<-time.Tick(start)
	log.Println("Started")
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

	chans := []chan rest.State{displayChan, socChan}

	if ct {
		wg.Add(1)
		ctChan := make(chan rest.State)
		go CtCoilHandler(ctx, wg, ctChan)
		chans = append(chans, ctChan)
	}

	fanout := Fanout(chans...)

	go rest.Poll(ctx, configFile, fanout)

	wg.Wait()
	return nil
}

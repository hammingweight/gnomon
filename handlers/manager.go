package handlers

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hammingweight/gnomon/rest"
)

func setupLogging(logfile string) (*os.File, error) {
	if logfile == "" {
		return nil, nil
	}

	f, err := os.OpenFile(logfile, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	log.SetOutput(f)
	return f, nil
}

func ManageInverter(logfile string, delay time.Duration, runTime time.Duration, configFile string, ct bool) error {
	f, err := setupLogging(logfile)
	if err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	if delay >= 5*time.Second {
		log.Printf("Waiting for %s to start...\n", delay)
	}
	time.Sleep(delay)
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

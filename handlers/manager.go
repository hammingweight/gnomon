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

// Package handlers includes functions that respond to changes reported by the
// inverter and can update the inverter's settings.
package handlers

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hammingweight/gnomon/api"
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

// ManageInverter spawns handlers to respond to changes in the inverter's state.
func ManageInverter(logfile string, delay time.Duration, runTime time.Duration, configFile string, minSoc int, ct bool) error {
	// Set up logging
	f, err := setupLogging(logfile)
	if err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	// Wait...
	if delay >= 5*time.Second {
		log.Printf("Waiting for %s to start...\n", delay)
	}
	time.Sleep(delay)
	log.Println("Starting management of the inverter")

	// Set up a context that will expire after the specified timeout, at which point this code
	// will stop managing the inverter.
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, runTime)
	defer cancel()

	// Add a handler to display the inverter's statistics.
	displayChan := make(chan api.State)
	go DisplayHandler(ctx, displayChan)

	// Add a handler to manage the battery's depth of discharge.
	wg := &sync.WaitGroup{}
	wg.Add(1)
	socChan := make(chan api.State)
	go SocHandler(ctx, wg, minSoc, socChan)

	// A slice of channels with handlers to respond to state changes.
	chans := []chan api.State{displayChan, socChan}

	// If the user wants gnomon to enable/disable power to the non-essential circuits,
	// add a handler.
	if ct {
		wg.Add(1)
		ctChan := make(chan api.State)
		go CtCoilHandler(ctx, wg, ctChan)
		chans = append(chans, ctChan)
	}

	// Create a fanout channel that will relay messages to the handler channels.
	fanout := Fanout(chans...)

	// Start polling and sending messages to the handlers when there are changes in state.
	go api.Poll(ctx, configFile, fanout)

	wg.Wait()
	if ctx.Err() != nil {
		log.Println("Deadline has expired; exiting")
	} else {
		log.Println("Handlers have finished managing the inverter; exiting early")
	}
	return nil
}

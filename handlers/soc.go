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

package handlers

import (
	"context"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/hammingweight/gnomon/api"
)

// SocHandler watches the battery's SoC and determines how to adjust the depth of
// discharge of the battery.
func SocHandler(ctx context.Context, wg *sync.WaitGroup, minSoc int, ch chan api.State) {
	log.Println("Starting management of the battery SOC")
	defer wg.Done()
	defer log.Println("Finished management of the battery SOC")

	var threshold int
	var err error
	for {
		select {
		case <-ch:
			threshold, err = api.BatteryDischargeThreshold(ctx)
			if err != nil {
				log.Println("Failed to read discharge threshold: ", err)
				continue
			}
			var lowBatteryCap int
			lowBatteryCap, err = api.LowBatteryCapacity(ctx)
			if err != nil {
				log.Println("Failed to read low battery capacity: ", err)
				continue
			}
			if minSoc < 0 {
				minSoc = lowBatteryCap + 20
			} else if minSoc < lowBatteryCap {
				log.Printf("Specified minimum battery SOC = %d%% is too low\n", minSoc)
				minSoc = lowBatteryCap
			}
			log.Printf("Minimum battery SOC = %d%%\n", minSoc)
		case <-ctx.Done():
			return
		}
		break
	}

	var maxSoc int
L:
	for {
		select {
		case <-ctx.Done():
			break L
		case s := <-ch:
			if s.Soc > maxSoc {
				maxSoc = s.Soc
			}
			if maxSoc >= 100 {
				break L
			}
		}
	}

	if maxSoc == 100 {
		threshold = 9 * threshold / 10
	} else if maxSoc == 99 {
		r := rand.Intn(2)
		threshold += r
	} else {
		r := math.Pow(100.0/float64(maxSoc), 0.5)
		newThreshold := int(r * float64(threshold))
		if newThreshold-threshold < 1 {
			newThreshold = threshold + 1
		}
		threshold = newThreshold
	}

	// Sanity checks
	if threshold < minSoc {
		threshold = minSoc
	}
	if threshold > 100 {
		threshold = 100
	}

	log.Printf("Setting battery's minimum SOC to %d%%\n", threshold)
	for i := 0; i < 10; i++ {
		if err = api.UpdateBatteryCapacity(threshold); err == nil {
			return
		}
		log.Println("Updating battery capacity failed: ", err)
		time.Sleep(5 * time.Second)
	}
	log.Println("Couldn't update battery capacity after 10 attempts, giving up")
}

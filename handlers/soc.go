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
	"sync"

	"github.com/hammingweight/gnomon/api"
)

// SocHandler watches the battery's SoC and determines how to adjust the depth of
// discharge of the battery.
func SocHandler(ctx context.Context, wg *sync.WaitGroup, minSoc int, ch chan api.State) {
	log.Println("Managing battery depth of discharge")
	defer wg.Done()

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
				log.Printf("Specified minimum battery SoC = %d%% is too low\n", minSoc)
				minSoc = lowBatteryCap
			}
			log.Printf("Minimum battery SoC = %d%%\n", minSoc)
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

	if maxSoc == 99 {
		log.Println("Leaving battery's minimum SoC unchanged")
		return
	} else if maxSoc == 100 {
		threshold -= 10
	} else if maxSoc < 80 {
		delta := 100 - maxSoc
		threshold += delta - 2
	} else {
		threshold += 2
	}

	// Sanity checks
	if threshold < minSoc {
		threshold = minSoc
	}
	if threshold > 100 {
		threshold = 100
	}

	log.Printf("Setting battery's minimum SoC to %d%%\n", threshold)
	err = api.UpdateBatteryCapacity(threshold)
	if err != nil {
		log.Println("updating battery capacity failed: ", err)
	}
}

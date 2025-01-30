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

func average(l []int) int {
	s := 0
	for _, v := range l {
		s += v
	}
	return s / len(l)
}

func switchOn(averagePower int, inverterPower int, soc int, thresholdSoc int) bool {
	if soc >= thresholdSoc+40 {
		return true
	}

	if soc < thresholdSoc+20 {
		return false
	}

	return averagePower > inverterPower/4
}

func switchOff(averagePower int, inverterPower int, soc int, thresholdSoc int) bool {
	if soc <= thresholdSoc+10 {
		return true
	}

	if soc >= 95 {
		return false
	}

	return averagePower < inverterPower/8 && soc < thresholdSoc+40
}

func handleEssentialOnly(averagePower int, inverterPower int, soc int, threshold int) (bool, error) {
	if switchOn(averagePower, inverterPower, soc, threshold) {
		log.Println("Configuring inverter to power all loads")
		return false, api.UpdateEssentialOnly(false)
	}
	return true, nil
}

func handleAllLoads(averagePower int, inverterPower int, soc int, threshold int) (bool, error) {
	if switchOff(averagePower, inverterPower, soc, threshold) {
		log.Println("Configuring inverter to power only the essential loads")
		return true, api.UpdateEssentialOnly(true)
	}
	return false, nil
}

// CtCoilHandler enables or disables power flowing from the inverter to non-essential
// circuits depending on the battery's SoC and the input power.
func CtCoilHandler(ctx context.Context, wg *sync.WaitGroup, ch chan api.State) {
	log.Println("Managing power to the CT")
	defer wg.Done()
	defer func() {
		log.Println("Shutting down; configuring inverter to power only the essential loads")
		err := api.UpdateEssentialOnly(true)
		if err != nil {
			log.Println("Failed to update inverter's settings: ", err)
		}
	}()

	powerReadings := []int{}
	var inverterPower int
	var threshold int
	var essentialOnly bool
	var err error
	for {
		select {
		case <-ch:
			inverterPower, err = api.InverterRatedPower()
			if err != nil {
				log.Println("Failed to read inverter's rated power: ", err)
				continue
			}
			essentialOnly = api.EssentialOnly()
			threshold, err = api.BatteryDischargeThreshold()
			if err != nil {
				log.Println("Failed to read discharge threshold: ", err)
				continue
			}
		case <-ctx.Done():
			return
		}
		break
	}

	for {
		select {
		case <-ctx.Done():
			return
		case s := <-ch:
			powerReadings = append(powerReadings, s.Power)
			if len(powerReadings) > 4 {
				powerReadings = powerReadings[len(powerReadings)-4:]
			}
			averagePower := average(powerReadings)
			var err error
			if essentialOnly {
				essentialOnly, err = handleEssentialOnly(averagePower, inverterPower, s.Soc, threshold)
			} else {
				essentialOnly, err = handleAllLoads(averagePower, inverterPower, s.Soc, threshold)
			}
			if err != nil {
				log.Println("Failed to reconfigure inverter: ", err)
			}
		}
	}
}

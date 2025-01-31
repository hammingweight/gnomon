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
	if len(l) == 0 {
		return 0
	}
	s := 0
	for _, v := range l {
		s += v
	}
	return s / len(l)
}

// upperTriggerOnSoc is the SOC at which the inverter should power
// non-essential loads irrespective of the input power.
func upperTriggerOnSoc(threshold int) int {
	if threshold+40 > 80 {
		return threshold + 40
	}
	return 80
}

// lowerTriggerOnSoc is the lowest SOC at which the inverter should power
// non-essential loads (but only if the input power exceeds some power threshold).
func lowerTriggerOnSoc(threshold int) int {
	return threshold + 20
}

// lowerTriggerSoc is the SOC at which the inverter should power only
// the essential loads irrespective of the input power
func lowerTriggerOffSoc(threshold int) int {
	return threshold + 10
}

func triggerOnPower(ratedPower int, threshold int, soc int) int {
	pu := ratedPower / 8
	pl := ratedPower / 4
	su := upperTriggerOnSoc(threshold)
	sl := lowerTriggerOnSoc(threshold)
	p := pl + (pu-pl)*(sl-soc)/(sl-su)
	return p
}

func shouldSwitchOn(averagePower int, inverterPower int, soc int, thresholdSoc int) bool {
	triggerSoc := upperTriggerOnSoc(thresholdSoc)
	if soc >= triggerSoc {
		log.Printf("Battery SOC, %d%%, is above %d%%", soc, triggerSoc)
		return true
	}

	if soc < lowerTriggerOnSoc(thresholdSoc) {
		return false
	}

	turnOnPower := triggerOnPower(inverterPower, thresholdSoc, soc)
	if averagePower > turnOnPower {
		log.Printf("Average input power, %dW, is above %dW", averagePower, turnOnPower)
		return true
	}
	return false
}

func shouldSwitchOff(averagePower int, inverterPower int, soc int, thresholdSoc int) bool {
	triggerSoc := lowerTriggerOffSoc(thresholdSoc)
	if soc <= triggerSoc {
		log.Printf("Battery SOC, %d%%, is below %d%%", soc, triggerSoc)
		return true
	}

	if soc >= 95 {
		return false
	}

	turnoffPower := inverterPower / 8
	if averagePower < turnoffPower && soc < upperTriggerOnSoc(thresholdSoc) {
		log.Printf("Average input power, %dW, is below %dW", averagePower, turnoffPower)
		return true
	}
	return false
}

func handleEssentialOnly(averagePower int, inverterPower int, soc int, threshold int) {
	if shouldSwitchOn(averagePower, inverterPower, soc, threshold) {
		log.Println("Configuring inverter to power all loads")
		err := api.UpdateEssentialOnly(false)
		if err != nil {
			log.Println("Failed to enable CT coil: ", err)
		}
	}
}

func handleAllLoads(averagePower int, inverterPower int, soc int, threshold int) {
	if shouldSwitchOff(averagePower, inverterPower, soc, threshold) {
		log.Println("Configuring inverter to power only the essential loads")
		err := api.UpdateEssentialOnly(true)
		if err != nil {
			log.Println("Failed to disable CT coil: ", err)
		}
	}
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
	var err error
	for {
		select {
		case <-ch:
			inverterPower, err = api.InverterRatedPower()
			if err != nil {
				log.Println("Failed to read inverter's rated power: ", err)
				continue
			}
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
			essentialOnly := api.EssentialOnly()
			averagePower := average(powerReadings)
			if essentialOnly {
				handleEssentialOnly(averagePower, inverterPower, s.Soc, threshold)
			} else {
				handleAllLoads(averagePower, inverterPower, s.Soc, threshold)
			}
		}
	}
}

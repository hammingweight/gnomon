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
	"time"

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
	u := min(threshold+45, 99)
	if u <= 85 {
		return 85
	}
	l := lowerTriggerOnSoc(threshold)
	return max(u, l)
}

// lowerTriggerOnSoc is the lowest SOC at which the inverter should power
// non-essential loads (but only if the input power exceeds some power threshold).
func lowerTriggerOnSoc(threshold int) int {
	if threshold+25 >= 95 {
		return 101
	}
	return threshold + 25
}

// lowerTriggerSoc is the SOC at which the inverter should power only
// the essential loads irrespective of the input power
func lowerTriggerOffSoc(threshold int) int {
	return threshold + 20
}

// triggerOnPower returns the minimum input power that is needed to allow
// the inverter to power non-essential loads. The higher the battery SoC,
// the lower the input power needed.
func triggerOnPower(ratedPower int, threshold int, soc int) int {
	pu := ratedPower / 8
	pl := ratedPower / 4
	su := upperTriggerOnSoc(threshold)
	sl := lowerTriggerOnSoc(threshold)
	if sl < su {
		return pl + (pu-pl)*(sl-soc)/(sl-su)
	}
	return pl
}

// shouldSwitchOn returns true if the SoC and input power are high enough to
// justify powering the non-essential loads from the inverter.
func shouldSwitchOn(averagePower int, inverterPower int, soc int, thresholdSoc int) bool {
	triggerSoc := upperTriggerOnSoc(thresholdSoc)
	if soc >= triggerSoc {
		return true
	}

	if soc < lowerTriggerOnSoc(thresholdSoc) {
		return false
	}

	turnOnPower := triggerOnPower(inverterPower, thresholdSoc, soc)
	return averagePower > turnOnPower
}

// shouldSwitchOff returns true if the SoC or power are low and the inverter should
// not power non-essential circuits.
func shouldSwitchOff(averagePower int, inverterPower int, soc int, thresholdSoc int) bool {
	return !shouldSwitchOn(averagePower, inverterPower, soc, thresholdSoc)
}

func handleEssentialOnly(ctx context.Context, averagePower int, inverterPower int, soc int, threshold int) {
	if shouldSwitchOn(averagePower, inverterPower, soc, threshold) {
		log.Println("Configuring inverter to power all loads")
		if err := api.UpdateEssentialOnly(false); err != nil {
			log.Println("Failed to enable CT coil: ", err)
		}
		for i := 0; i < 10; i++ {
			if !api.EssentialOnly(ctx) {
				log.Println("Successfully updated inverter")
				return
			}
			time.Sleep(10 * time.Second)
		}
		log.Println("Failed to update inverter")
	}
}

func handleAllLoads(ctx context.Context, averagePower int, inverterPower int, soc int, threshold int) {
	if shouldSwitchOff(averagePower, inverterPower, soc, threshold) {
		log.Println("Configuring inverter to power only essential loads")
		if err := api.UpdateEssentialOnly(true); err != nil {
			log.Println("Failed to disable CT coil: ", err)
		}
		for i := 0; i < 10; i++ {
			if api.EssentialOnly(ctx) {
				log.Println("Successfully updated inverter")
				return
			}
			time.Sleep(10 * time.Second)
		}
		log.Println("Failed to update inverter")
	}
}

func manageCoil(ctx context.Context, averagePower int, inverterPower int, soc int, threshold int) {
	essentialOnly := api.EssentialOnly(ctx)
	if essentialOnly {
		handleEssentialOnly(ctx, averagePower, inverterPower, soc, threshold)
	} else {
		handleAllLoads(ctx, averagePower, inverterPower, soc, threshold)
	}
}

type powerTime struct {
	power int
	t     time.Time
}

func newPowerTime(power int) powerTime {
	return powerTime{power, time.Now()}
}

func getRecentPowerReadings(powerTimes *[]powerTime) []int {
	for {
		if (*powerTimes)[0].t.After(time.Now().Add(-20 * time.Minute)) {
			break
		}
		*powerTimes = (*powerTimes)[1:]
	}
	powers := []int{}
	for _, v := range *powerTimes {
		powers = append(powers, v.power)
	}
	return powers
}

// CtCoilHandler enables or disables power flowing from the inverter to non-essential
// circuits depending on the battery's SoC and the input power.
func CtCoilHandler(ctx context.Context, wg *sync.WaitGroup, ch chan api.State) {
	log.Println("Starting power management to the CT")
	defer wg.Done()
	defer func() {
		log.Println("Configuring inverter to power only the essential loads")
		for i := 0; i < 10; i++ {
			err := api.UpdateEssentialOnly(true)
			if err != nil {
				log.Println("Failed to update inverter's settings: ", err)
			}
			time.Sleep(30 * time.Second)
			if api.EssentialOnly(context.Background()) {
				break
			}
		}
		log.Println("Finished power management to the CT")
	}()

	powerReadings := []powerTime{}
	var inverterPower int
	var threshold int
	var err error
	for {
		select {
		case <-ch:
			inverterPower, err = api.InverterRatedPower(ctx)
			if err != nil {
				log.Println("Failed to read inverter's rated power: ", err)
				continue
			}
			threshold, err = api.BatteryDischargeThreshold(ctx)
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
			powerReadings = append(powerReadings, newPowerTime(s.Power))
			averagePower := average(getRecentPowerReadings(&powerReadings))
			manageCoil(ctx, averagePower, inverterPower, s.Soc, threshold)
		}
	}
}

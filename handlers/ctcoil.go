package handlers

import (
	"context"
	"log"
	"sync"

	"github.com/hammingweight/gnomon/rest"
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

func handleEssentialOnly(averagePower int, inverterPower int, soc int, threshold int) error {
	if switchOn(averagePower, inverterPower, soc, threshold) {
		log.Println("Configuring inverter to power all loads")
		return rest.UpdateEssentialOnly(false)
	}
	return nil
}

func handleAllLoads(averagePower int, inverterPower int, soc int, threshold int) error {
	if switchOff(averagePower, inverterPower, soc, threshold) {
		log.Println("Configuring inverter to power only the essential loads")
		return rest.UpdateEssentialOnly(true)
	}
	return nil
}

func CtCoilHandler(ctx context.Context, wg *sync.WaitGroup, ch chan rest.State) {
	defer wg.Done()
	defer func() {
		log.Println("Shutting down; configuring inverter to power only the essential loads")
		err := rest.UpdateEssentialOnly(true)
		if err != nil {
			log.Println("Failed to update inverter's settings: ", err)
		}
	}()

	powerReadings := make([]int, 4)
	var inverterPower int
	var err error
	for {
		select {
		case <-ch:
		case <-ctx.Done():
			return
		}
		inverterPower, err = rest.GetInverterPower()
		if err == nil {
			break
		}
		log.Println("Failed to read inverter's rated power: ", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case s := <-ch:
			powerReadings = powerReadings[1:]
			powerReadings = append(powerReadings, s.Power)
			averagePower := average(powerReadings)
			var err error
			if s.EssentialOnly {
				err = handleEssentialOnly(averagePower, inverterPower, s.Soc, s.Threshold)
			} else {
				err = handleAllLoads(averagePower, inverterPower, s.Soc, s.Threshold)
			}
			if err != nil {
				log.Println("Failed to reconfigure inverter: ", err)
			}
		}
	}
}

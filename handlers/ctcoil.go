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

func handleEssentialOnly(averagePower int, inverterPower int, soc int, threshold int) (bool, error) {
	if switchOn(averagePower, inverterPower, soc, threshold) {
		log.Println("Configuring inverter to power all loads")
		return false, rest.UpdateEssentialOnly(false)
	}
	return true, nil
}

func handleAllLoads(averagePower int, inverterPower int, soc int, threshold int) (bool, error) {
	if switchOff(averagePower, inverterPower, soc, threshold) {
		log.Println("Configuring inverter to power only the essential loads")
		return true, rest.UpdateEssentialOnly(true)
	}
	return false, nil
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
	var threshold int
	var essentialOnly bool
	var err error
	for {
		select {
		case <-ch:
			inverterPower, err = rest.GetInverterPower()
			if err != nil {
				log.Println("Failed to read inverter's rated power: ", err)
				continue
			}
			essentialOnly = rest.EssentialOnly()
			threshold, err = rest.GetDischargeThreshold()
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
			powerReadings = powerReadings[1:]
			powerReadings = append(powerReadings, s.Power)
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

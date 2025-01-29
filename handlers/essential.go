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

	return averagePower < inverterPower/8
}

func handleEssentialOnly(averagePower int, inverterPower int, soc int, threshold int) {
	if switchOn(averagePower, inverterPower, soc, threshold) {
		log.Println("powering all loads")
		rest.UpdateEssentialOnly(false)
	}
}

func handleAllLoads(averagePower int, inverterPower int, soc int, threshold int) {
	if switchOff(averagePower, inverterPower, soc, threshold) {
		log.Println("powering essential loads only")
		rest.UpdateEssentialOnly(true)
	}
}

func EssentialOnlyHandler(ctx context.Context, wg *sync.WaitGroup, ch chan rest.State) {
	defer wg.Done()
	powerReadings := make([]int, 4)

	inverterPower := rest.GetInverterPower()

	for {
		select {
		case <-ctx.Done():
			rest.UpdateEssentialOnly(true)
			return
		case s := <-ch:
			powerReadings = powerReadings[1:]
			powerReadings = append(powerReadings, s.Power)
			averagePower := average(powerReadings)
			log.Println("Average power: ", averagePower)
			if s.EssentialOnly {
				handleEssentialOnly(averagePower, inverterPower, s.Soc, s.Threshold)
			} else {
				handleAllLoads(averagePower, inverterPower, s.Soc, s.Threshold)
			}
		}
	}
}

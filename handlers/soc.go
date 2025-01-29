package handlers

import (
	"context"
	"log"
	"sync"

	"github.com/hammingweight/gnomon/rest"
)

func SocHandler(ctx context.Context, wg *sync.WaitGroup, ch chan rest.State) {
	defer wg.Done()

	threshold := -1
	maxSoc := -1
L:
	for {
		select {
		case <-ctx.Done():
			break L
		case s := <-ch:
			threshold = s.Threshold
			if s.Soc > maxSoc {
				maxSoc = s.Soc
			}
			if maxSoc >= 100 {
				break L
			}
		}
	}

	if maxSoc == 99 || threshold == -1 {
		return
	}
	if maxSoc == 100 {
		threshold -= 10
		if threshold < 30 {
			threshold = 30
		}
	} else {
		threshold += 2
		if threshold > 100 {
			threshold = 100
		}
	}

	log.Printf("Setting battery discharge threshold to %d%%\n", threshold)
	err := rest.UpdateBatteryCapacity(threshold)
	if err != nil {
		log.Println("updating battery capacity failed: ", err)
	}
}

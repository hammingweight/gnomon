package handlers

import (
	"context"
	"log"
	"sync"

	"github.com/hammingweight/gnomon/rest"
)

func SocHandler(ctx context.Context, wg *sync.WaitGroup, ch chan rest.State) {
	defer wg.Done()

	var threshold int
	var maxSoc int
	var err error
	for {
		select {
		case <-ch:
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
	err = rest.UpdateBatteryCapacity(threshold)
	if err != nil {
		log.Println("updating battery capacity failed: ", err)
	}
}

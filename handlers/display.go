package handlers

import (
	"context"
	"log"
	"sync"

	"github.com/hammingweight/gnomon/rest"
)

func DisplayHandler(ctx context.Context, wg *sync.WaitGroup, ch chan rest.State) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("Done")
			return

		case state := <-ch:
			log.Println(state)
		}
	}
}

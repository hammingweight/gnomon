package handlers

import (
	"github.com/hammingweight/gnomon/rest"
)

func Fanout(chans ...chan rest.State) chan rest.State {
	ch := make(chan rest.State)
	go func() {
		for {
			v := <-ch
			for _, c := range chans {
				select {
				case c <- v:
					continue
				default:
					continue
				}
			}
		}
	}()

	return ch
}

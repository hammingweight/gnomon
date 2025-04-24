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
	"github.com/emicklei/go-restful/v3/log"
	"github.com/hammingweight/gnomon/api"
)

// Fanout creates channel that receives messages and relays them to handler channels.
func Fanout(chans ...chan api.State) chan api.State {
	ch := make(chan api.State)
	go func() {
		defer log.Println("Finished fanout")
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

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

	"github.com/hammingweight/gnomon/api"
)

// DisplayHandler displays the state of the inverter whenever it changes.
func DisplayHandler(ctx context.Context, ch chan api.State) {
	defer log.Println("Finished displaying inverter state")
	for {
		select {
		case <-ctx.Done():
			return
		case state := <-ch:
			log.Println(state)
		}
	}
}

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

package api

import "fmt"

// State represents the state of the inverter: input power, battery SoC, load power and the
// time at which the measurements were taken.
type State struct {
	Power int
	Soc   int
	Load  int
	Time  string
}

func (s State) String() string {
	return fmt.Sprintf("Input power = %dW, Battery SOC = %d%%, Load = %dW.", s.Power, s.Soc, s.Load)
}

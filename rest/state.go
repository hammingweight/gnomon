package rest

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

package rest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hammingweight/synkctl/configuration"
	"github.com/hammingweight/synkctl/rest"
)

type client struct {
	mutex      sync.Mutex
	client     *rest.SynkClient
	configFile string
}

var c client

type State struct {
	Power         int
	MaxPower      int
	Soc           int
	MaxSoc        int
	Load          int
	EssentialOnly bool
	Threshold     int
	Time          string
}

func (s State) String() string {
	whichLoads := "The inverter is powering all loads"
	if s.EssentialOnly {
		whichLoads = "The inverter is powering essential loads only"
	}
	return fmt.Sprintf("%s: Input power = %dW, Battery SOC = %d%%, Load = %dW", whichLoads, s.Power, s.Soc, s.Load)
}

func Authenticate(ctx context.Context) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cfg, err := configuration.ReadConfigurationFromFile(c.configFile)
	if err != nil {
		log.Println("error authenticating: ", err)
		os.Exit(1)
	}
	for {
		if ctx.Err() != nil {
			return
		}
		client, err := rest.Authenticate(ctx, cfg)
		if err == nil {
			c.client = client
			return
		}
		log.Println("failed to authenticate: ", err)
		time.Sleep(30 * time.Second)
	}
}

var s State

func readState(ctx context.Context, lastUpdateTime string) (*State, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	input, err := c.client.Input(ctx)
	if err != nil {
		return nil, err
	}
	s.Power = input.Power()
	if s.Power > s.MaxPower {
		s.MaxPower = s.Power
	}
	pv, ok := input.PV(0)
	if !ok {
		return nil, errors.New("can't read time")
	}
	t, ok := pv["time"]
	if !ok {
		return nil, errors.New("can't read time")
	}
	s.Time = t.(string)
	if s.Time == lastUpdateTime {
		return nil, nil
	}

	inv, err := c.client.Inverter(ctx)
	if err != nil {
		return nil, err
	}
	s.Threshold = inv.BatteryCapacity()
	s.EssentialOnly = inv.EssentialOnly()

	bat, err := c.client.Battery(ctx)
	if err != nil {
		return nil, err
	}
	s.Soc = bat.SOC()
	if s.Soc > s.MaxSoc {
		s.MaxSoc = s.Soc
	}

	load, err := c.client.Load(ctx)
	if err != nil {
		return nil, err
	}
	s.Load = load.Power()

	return &s, nil
}

func Poll(ctx context.Context, configFile string, ch chan State) {
	reauthFlag := true
	lastUpdateTime := ""
	c.configFile = configFile
	for {
		if reauthFlag {
			Authenticate(ctx)
		} else {
			select {
			case <-time.Tick(time.Minute):
			case <-ctx.Done():
				return
			}
		}
		reauthFlag = false
		state, err := readState(ctx, lastUpdateTime)
		if err != nil {
			reauthFlag = true
			log.Println("error during poll: ", err)
			time.Sleep(30 * time.Second)
			continue
		}
		if state == nil {
			continue
		}
		lastUpdateTime = state.Time
		ch <- *state
	}
}

func UpdateBatteryCapacity(cap int) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ctx := context.Background()

	inv, err := c.client.Inverter(ctx)
	if err != nil {
		return err
	}
	inv.SetBatteryCapacity(cap)
	return c.client.UpdateInverter(ctx, inv)
}

func UpdateEssentialOnly(eo bool) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ctx := context.Background()

	inv, err := c.client.Inverter(ctx)
	if err != nil {
		return err
	}
	inv.SetEssentialOnly(eo)
	return c.client.UpdateInverter(ctx, inv)
}

func GetInverterPower() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ctx := context.Background()

	for {
		details, err := c.client.Details(ctx)
		if err != nil {
			Authenticate(ctx)
			continue
		}
		return details.RatedPower()
	}
}

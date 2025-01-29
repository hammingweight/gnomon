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
	Soc           int
	Load          int
	EssentialOnly bool
	Threshold     int
	time          string
}

func (s State) String() string {
	whichLoads := "all loads"
	if s.EssentialOnly {
		whichLoads = "only essential loads"
	}
	return fmt.Sprintf("The inverter is powering %s. Input power = %dW, Battery SOC = %d%%, Load = %dW.", whichLoads, s.Power, s.Soc, s.Load)
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

func readState(ctx context.Context, s *State) (bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	input, err := c.client.Input(ctx)
	if err != nil {
		return false, err
	}
	s.Power, err = input.Power()
	if err != nil {
		return false, err
	}
	pv, ok := input.PV(0)
	if !ok {
		return false, errors.New("can't read MPPT values")
	}
	t, ok := pv["time"]
	if !ok {
		return false, errors.New("can't read update time")
	}
	updateTime := t.(string)
	if s.time == updateTime {
		return false, nil
	}
	s.time = updateTime

	inv, err := c.client.Inverter(ctx)
	if err != nil {
		return false, err
	}
	bc, err := inv.BatteryCapacity()
	if err != nil {
		return false, err
	}
	s.Threshold = bc
	s.EssentialOnly = inv.EssentialOnly()

	bat, err := c.client.Battery(ctx)
	if err != nil {
		return false, err
	}
	s.Soc, err = bat.SOC()
	if err != nil {
		return false, err
	}

	load, err := c.client.Load(ctx)
	if err != nil {
		return false, err
	}
	s.Load, err = load.Power()
	if err != nil {
		return false, err
	}

	return true, nil
}

func Poll(ctx context.Context, configFile string, ch chan State) {
	reauthFlag := true
	c.configFile = configFile
	s := &State{}
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
		changed, err := readState(ctx, s)
		if err != nil {
			reauthFlag = true
			log.Println("error during poll: ", err)
			time.Sleep(30 * time.Second)
			continue
		}
		if changed {
			ch <- *s
		}
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

func GetInverterPower() (int, error) {
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

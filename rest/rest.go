package rest

import (
	"context"
	"errors"
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

func authenticate(ctx context.Context) {
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

// ReadState reads the current state of the inverter. The state must
// be passed as a pointer; the reference state will be updated if the
// SunSynk API returns fresh data. This function returns false if the
// state is unchanged.
func ReadState(ctx context.Context, s *State) (bool, error) {
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
	if s.Time == updateTime {
		return false, nil
	}
	s.Time = updateTime

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

// Poll polls the SunSynk API and sends changes to the channel passed
// as an argument.
func Poll(ctx context.Context, configFile string, ch chan State) {
	reauthFlag := true
	c.configFile = configFile
	s := &State{}
	for {
		if reauthFlag {
			authenticate(ctx)
		} else {
			select {
			case <-time.Tick(time.Minute):
			case <-ctx.Done():
				return
			}
		}
		reauthFlag = false
		changed, err := ReadState(ctx, s)
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
			authenticate(ctx)
			continue
		}
		return details.RatedPower()
	}
}

func GetDischargeThreshold() (int, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ctx := context.Background()

	for {
		inv, err := c.client.Inverter(ctx)
		if err != nil {
			authenticate(ctx)
			continue
		}
		return inv.BatteryCapacity()
	}
}

func EssentialOnly() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ctx := context.Background()

	for {
		inv, err := c.client.Inverter(ctx)
		if err != nil {
			authenticate(ctx)
			continue
		}
		return inv.EssentialOnly()
	}
}

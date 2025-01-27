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

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hammingweight/synkctl/configuration"
	"github.com/hammingweight/synkctl/rest"
	"github.com/spf13/cobra"
)

var Version string = "1.0.0"

func getConfigurationFromFlags(cmd *cobra.Command) (*configuration.Configuration, error) {
	username, _ := cmd.Flags().GetString("user")
	password, _ := cmd.Flags().GetString("password")
	if password == "" {
		return nil, errors.New("no password (--password) specified")
	}
	endpoint, _ := cmd.Flags().GetString("endpoint")
	return &configuration.Configuration{User: username, Password: password, Endpoint: endpoint}, nil

}

func run(cmd *cobra.Command) error {
	t, err := cmd.Flags().GetInt("time")
	if err != nil {
		return err
	}
	if t <= 0 || t >= 24 {
		return errors.New("you must specify a running time (--time) in the range 1-23 hours")
	}
	username, _ := cmd.Flags().GetString("user")
	var config *configuration.Configuration
	if username != "" {
		config, err = getConfigurationFromFlags(cmd)
	} else {
		configFile, _ := cmd.Flags().GetString("config")
		config, err = configuration.ReadConfigurationFromFile(configFile)
	}
	if err != nil {
		return err
	}
	if config.DefaultInverterSN == "" {
		sn, _ := cmd.Flags().GetString("inverter")
		if sn == "" {
			return errors.New("no inverter serial number (--inverter) specified")
		}
		config.DefaultInverterSN = sn
	}

	client, err := rest.Authenticate(context.Background(), config)
	if err != nil {
		return err
	}
	inverters, _ := client.ListInverters(context.Background())
	invOk := false
	for _, i := range inverters {
		if i == config.DefaultInverterSN {
			invOk = true
		}
	}
	if !invOk {
		return fmt.Errorf("'%s' is not a valid serial number", config.DefaultInverterSN)
	}

	return nil
}

var gnomonCmd = &cobra.Command{
	Use:     "gnomon",
	Short:   "Manages a SunSynk inverter's settings daily",
	Args:    cobra.ExactArgs(0),
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(cmd)
	},
}

// Execute is called by main.main() and executes the gnomon command.
func Execute() {
	err := gnomonCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	configFile, err := configuration.DefaultConfigurationFile()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	gnomonCmd.Flags().StringP("config", "c", configFile, "synkctl config file path")
	gnomonCmd.Flags().StringP("user", "u", "", "SunSynk user")
	gnomonCmd.Flags().StringP("password", "p", "", "SunSynk user's password")
	gnomonCmd.Flags().StringP("inverter", "i", "", "inverter serial number")
	gnomonCmd.Flags().StringP("endpoint", "e", "https://api.sunsynk.net", "SunSynk API endpoint")
	gnomonCmd.Flags().IntP("time", "t", 0, "time (in hours) for which gnomom must run (0<t<24)")
}

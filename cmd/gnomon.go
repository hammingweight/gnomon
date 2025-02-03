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
	"fmt"
	"os"
	"time"

	"github.com/hammingweight/gnomon/handlers"
	"github.com/hammingweight/synkctl/configuration"
	"github.com/spf13/cobra"
)

// Version is injected by the build.
var Version string = ""

func getDelayAndRunningTime() (time.Duration, time.Duration, error) {
	delay, err := startTime.Until()
	if err != nil {
		return 0, 0, err
	}
	var end time.Duration
	if endTime == "" {
		end = time.Hour * 12
	} else {
		end, err = endTime.Until()
		if err != nil {
			return 0, 0, err
		}
	}
	if end < delay {
		end += 24 * time.Hour
	}
	runTime := end - delay
	return delay, runTime, nil
}

var startTime HhMm
var endTime HhMm
var minSoc SoC = SoC(-1)

func run(cmd *cobra.Command) error {
	// Set up logging
	logfile, err := cmd.Flags().GetString("logfile")
	if err != nil {
		return err
	}

	// Find when to start running and for how long.
	delay, runTime, err := getDelayAndRunningTime()
	if err != nil {
		return err
	}

	// Find the config file.
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	// Geyser1
	g1Start := HhMm("08:59")
	g1Delay, _ := g1Start.Until()

	// Geyser2
	g2Start := HhMm("10:29")
	g2Delay, _ := g2Start.Until()

	// Start managing.
	return handlers.ManageInverter(logfile, delay, runTime, configFile, minSoc.Int(), g1Delay, g2Delay)
}

var gnomonCmd = &cobra.Command{
	Use:   "gnomon",
	Short: "Manages a SunSynk inverter's settings",
	Long: `gnomon is a tool for automatically managing a SunSynk inverter's settings.
It adjusts the depth of discharge of the battery and, optionally, can decide when to
allow the inverter to power non-essential loads.`,
	Args:    cobra.ExactArgs(0),
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(cmd)
	},
}

// Execute is called by main.main() and executes the gnomon command.
func Execute() {
	if err := gnomonCmd.Execute(); err != nil {
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
	gnomonCmd.Flags().VarP(&startTime, "start", "s", "start time in 24 hour HH:MM format, e.g. 06:00")
	gnomonCmd.Flags().VarP(&endTime, "end", "e", "end time in 24 hour HH:MM format, e.g. 19:30")
	gnomonCmd.Flags().StringP("logfile", "l", "", "log file path")
	gnomonCmd.Flags().VarP(&minSoc, "min-soc", "m", "minimum battery state of charge")
}

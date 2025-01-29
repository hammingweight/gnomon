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

var Version string = "1.0.0"

func run(_ *cobra.Command) error {
	start, err := startTime.Until()
	if err != nil {
		return err
	}
	var end time.Duration
	end, err = endTime.Until()
	if err != nil {
		return err
	}
	if end < start {
		end += 24 * time.Hour
	}
	if endTime == "" {
		end = time.Hour * 12
	}
	runTime := end - start
	return handlers.Execute(start, runTime)
}

var startTime HhMm
var endTime HhMm

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
	gnomonCmd.Flags().VarP(&startTime, "start", "s", "start time in 24 hour HH:MM format, e.g. 06:00")
	gnomonCmd.Flags().VarP(&endTime, "end", "e", "end time in 24 hour HH:MM format, e.g. 19:30")
}

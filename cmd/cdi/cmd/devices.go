/*
   Copyright © 2021 The CDI Authors

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
	"path/filepath"

	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	"github.com/spf13/cobra"
)

type devicesFlags struct {
	verbose bool
	output  string
}

// devicesCmd is our command for listing devices found in the CDI registry.
var devicesCmd = &cobra.Command{
	Aliases: []string{"devs", "dev"},
	Use:     "devices",
	Short:   "List devices in the CDI registry",
	Long: `
The 'devices' command lists devices found in the CDI registry.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listDevices(devicesCfg.verbose, devicesCfg.output)
	},
}

func listDevices(verbose bool, output string) error {
	devices := cdi.ListDevices()
	if len(devices) == 0 {
		fmt.Printf("No CDI Devices found.\n")
		return nil
	}

	fmt.Printf("CDI Devices:\n")
	for _, device := range devices {
		printDevice(device, verbose, output, 2)
	}

	return nil
}

func printDevice(device string, verbose bool, output string, level int) error {
	if !verbose {
		fmt.Printf("  %s\n", device)
		return nil
	}

	dev := cdi.GetRegistry().GetDevice(device)
	spec := dev.GetSpec()

	if output == "" {
		if filepath.Ext(spec.GetPath()) == ".json" {
			output = "json"
		} else {
			output = "yaml"
		}
	}

	fmt.Printf("  %s (%s)\n", device, spec.GetPath())
	fmt.Printf("%s", marshalObject(level+2, dev.Device, output))
	specEdits := spec.ContainerEdits
	if len(specEdits.Env) > 0 || len(specEdits.DeviceNodes) > 0 ||
		len(specEdits.Hooks) > 0 || len(specEdits.Mounts) > 0 {
		fmt.Printf("%s(spec-scope-)containerEdits:\n", indent(level+2))
		fmt.Printf("%s", marshalObject(level+4, spec.ContainerEdits, output))
	}

	return nil
}

var (
	devicesCfg devicesFlags
)

func init() {
	rootCmd.AddCommand(devicesCmd)
	devicesCmd.Flags().BoolVarP(&devicesCfg.verbose,
		"verbose", "v", false, "list CDI Spec details")
	devicesCmd.Flags().StringVarP(&devicesCfg.output,
		"output", "o", "", "output format for details (json|yaml)")
}

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
	"sort"
	"strings"

	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	"github.com/spf13/cobra"
)

// classesCmd is our command for listing device classes in the registry.
var classesCmd = &cobra.Command{
	Use:   "classes",
	Short: "List CDI device classes",
	Long:  `List CDI device classes found in the registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		listDeviceClasses()
	},
}

func listDeviceClasses() {
	vendors := map[string][]string{}
	for _, vendor := range cdi.GetRegistry().ListVendors() {
		for _, spec := range cdi.GetRegistry().GetVendorSpecs(vendor) {
			class := spec.GetClass()
			vendors[class] = append(vendors[class], vendor)
		}
	}

	if len(vendors) == 0 {
		fmt.Printf("No CDI device classes found.\n")
		return
	}

	classes := []string{}
	for class := range vendors {
		classes = append(classes, class)
		sort.Strings(vendors[class])
	}
	sort.Strings(classes)

	fmt.Printf("CDI device classes:\n")
	for _, class := range classes {
		var (
			uniq []string
			prev string
		)
		for _, vendor := range vendors[class] {
			if vendor != prev {
				uniq = append(uniq, vendor)
			}
			prev = vendor
		}

		fmt.Printf("  %s (%s)\n", class, strings.Join(uniq, ", "))
	}
}

func init() {
	rootCmd.AddCommand(classesCmd)
}

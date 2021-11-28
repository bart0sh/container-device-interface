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

	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	"github.com/spf13/cobra"
)

// vendorsCmd is our command for listing vendors.
var vendorsCmd = &cobra.Command{
	Use:   "vendors",
	Short: "List vendors",
	Long:  `List vendors with CDI Specs in the registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		listVendors()
	},
}

func listVendors() {
	vendors := cdi.GetRegistry().ListVendors()
	if len(vendors) == 0 {
		fmt.Printf("No vendors found in CDI registry.\n")
		return
	}

	fmt.Printf("Vendors in CDI Registry:\n")
	for _, vendor := range vendors {
		fmt.Printf("  %s (%d CDI Spec Files)\n", vendor,
			len(cdi.GetRegistry().GetVendorSpecs(vendor)))
	}
}

func init() {
	rootCmd.AddCommand(vendorsCmd)
}

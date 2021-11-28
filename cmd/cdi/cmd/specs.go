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
	"strings"

	"github.com/pkg/errors"

	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	"github.com/spf13/cobra"
)

type specFlags struct {
	verbose bool
	output  string
}

// specsCmd is our command for listing Spec files.
var specsCmd = &cobra.Command{
	Use:   "specs [vendor-list]",
	Short: "List available CDI Specs",
	Long: fmt.Sprintf(`
The 'specs' command lists all CDI Specs present in the registry.
If a vendor list is given, only CDI Specs by the given vendors are
listed. The CDI Specs are discovered and loaded to the registry
from CDI Spec directories. The default CDI Spec directories are:
    %s.`, strings.Join(cdi.DefaultSpecDirs, ", ")),
	RunE: func(cmd *cobra.Command, vendors []string) error {
		return listSpecFiles(specCfg.verbose, specCfg.output, vendors...)
	},
}

func listSpecFiles(verbose bool, output string, vendors ...string) error {
	switch output {
	case "yaml", "json", "":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	if len(vendors) == 0 {
		vendors = cdi.GetRegistry().ListVendors()
	}

	fmt.Printf("Spec Files in CDI Registry:\n")
	for _, vendor := range vendors {
		specs := cdi.GetRegistry().GetVendorSpecs(vendor)
		if len(specs) == 0 {
			fmt.Printf("Vendor %s has no CDI Specs\n", vendor)
			continue
		}

		fmt.Printf("Vendor %s\n", vendor)
		for _, spec := range specs {
			printSpec(spec, verbose, output, 2)
		}
	}

	if errors := cdi.Errors(); len(errors) != 0 {
		fmt.Printf("*** CDI Registry has errors:\n")
		for _, err := range errors {
			fmt.Printf("  %v\n", err)
		}
	}

	return nil
}

func printSpec(spec *cdi.Spec, verbose bool, output string, level int) {
	fmt.Printf("%sSpec File %s\n", indent(level), spec.GetPath())

	if !verbose {
		return
	}

	if output == "" {
		if filepath.Ext(spec.GetPath()) == ".json" {
			output = "json"
		} else {
			output = "yaml"
		}
	}

	fmt.Printf("%s", marshalObject(level+2, spec.Spec, output))
}

var (
	specCfg specFlags
)

func init() {
	rootCmd.AddCommand(specsCmd)
	specsCmd.Flags().BoolVarP(&specCfg.verbose,
		"verbose", "v", false, "list CDI Spec details")
	specsCmd.Flags().StringVarP(&specCfg.output,
		"output", "o", "", "output format for details (json|yaml)")
}

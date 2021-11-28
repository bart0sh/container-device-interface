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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"sigs.k8s.io/yaml"

	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	"github.com/spf13/cobra"
)

type injectFlags struct {
	output string
}

// injectCmd is our command for injecting CDI devices into an OCI Spec.
var injectCmd = &cobra.Command{
	Aliases: []string{"inj", "in", "oci"},
	Use:     "inject <OCI Spec File> <CDI-device-list>",
	Short:   "Inject CDI devices into an OCI Spec",
	Long: `
The 'inject' command reads an OCI Spec from a file (use "-" for stdin),
injects a requested set of CDI devices into it and dumps the resulting
updated OCI Spec.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("OCI Spec and CDI device(s) arguments expected")
		}
		if err := injectDevices(args[0], args[1:]...); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		return nil
	},
}

func injectDevices(ociSpecFile string, patterns ...string) error {
	ociSpec, err := readOCISpec(ociSpecFile)
	if err != nil {
		return err
	}

	devices := []string{}
	for _, p := range patterns {
		matches, err := cdi.MatchDevices(p)
		if err != nil {
			return errors.Wrapf(err, "device match failed for pattern %q", p)
		}
		devices = append(devices, matches...)
	}

	unresolved, err := cdi.InjectDevices(ociSpec, devices)
	if err != nil {
		return errors.Wrapf(err, "OCI device injection failed")
	}
	if len(unresolved) != 0 {
		return errors.Errorf("unknown OCI devices: %s",
			strings.Join(devices, ", "))
	}

	output := injectCfg.output
	if output == "" {
		if filepath.Ext(ociSpecFile) == ".json" {
			output = "json"
		} else {
			output = "yaml"
		}
	}

	fmt.Printf("Updated OCI Spec:\n")
	fmt.Printf("%s", marshalObject(2, ociSpec, output))

	return nil
}

func readOCISpec(path string) (*oci.Spec, error) {
	var (
		spec *oci.Spec
		data []byte
		err  error
	)

	if path == "-" {
		data, err = ioutil.ReadAll(os.Stdin)
	} else {
		data, err = ioutil.ReadFile(path)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to read OCI Spec (%q)", path)
	}

	spec = &oci.Spec{}
	if err = yaml.Unmarshal(data, spec); err != nil {
		return nil, errors.Wrapf(err, "failed to parse OCI Spec (%q)", path)
	}

	return spec, nil
}

var (
	injectCfg injectFlags
)

func init() {
	rootCmd.AddCommand(injectCmd)
	injectCmd.Flags().StringVarP(&injectCfg.output,
		"output", "o", "", "output format for OCI Spec (json|yaml)")
}

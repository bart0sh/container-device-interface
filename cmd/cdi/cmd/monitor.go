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
	"os"
	"path/filepath"

	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

type monitorFlags struct {
	verbose bool
	output  string
}

// monitorCmd is our command for monitoring CDI Spec refreshes.
var monitorCmd = &cobra.Command{
	Use:   "monitor [specs] [vendors] [classes] [devices] [all]",
	Short: "Monitor CDI Spec directories and refresh on changes",
	Long: `
The 'monitor' command monitors the CDI Spec directories and refreshes
the cache upon changes. The arguments passed to monitor control what
information to show upon each refresh.`,
	Run: func(cmd *cobra.Command, args []string) {
		monitorSpecDirs(args...)
	},
}

func monitorSpecDirs(args ...string) {
	var (
		w    *fsnotify.Watcher
		err  error
		done chan error
	)

	if w, err = fsnotify.NewWatcher(); err != nil {
		fmt.Printf("failed to create fsnotify watch: %v", err)
		os.Exit(1)
	}

	for _, dir := range cdi.GetRegistry().GetSpecDirectories() {
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("directory %s does not exist, ignoring it...\n",
					dir)
				continue
			}
			fmt.Printf("failed to stat directory %s: %v\n", dir, err)
			os.Exit(1)
		}

		if err = w.Add(dir); err != nil {
			fmt.Printf("failed to watch CDI directory %q: %v\n", dir, err)
			os.Exit(1)
		}
	}

	done = make(chan error, 1)

	go func() {
		if len(args) == 0 {
			args = []string{"all"}
		}

		showRegistry(args...)

		for {
			select {
			case evt, ok := <-w.Events:
				if !ok {
					close(done)
					return
				}

				fmt.Printf("* watch event: %q %s\n", evt.Name, evt.Op)

				if evt.Op != fsnotify.Write && evt.Op != fsnotify.Remove {
					fmt.Printf(" => ignored\n")
					continue
				}
				name := filepath.Base(evt.Name)
				if name != "" && (name[0] == '.' || name[0] == '#') {
					fmt.Printf(" => ignored\n")
					continue
				}

				if err = cdi.Refresh(); err != nil {
					fmt.Printf("CDI Registry refresh failed: %v\n", err)
				} else {
					fmt.Printf(" => CDI Registry refreshed\n")
					showRegistry(args...)
				}

			case err, ok := <-w.Errors:
				if ok {
					done <- err
				}
				return
			}
		}
	}()

	err = <-done
	if err != nil {
		fmt.Printf("CDI Spec watch failed: %v\n", err)
		os.Exit(1)
	}
}

func showRegistry(args ...string) {
	for _, what := range args {
		switch what {
		case "vendors", "vendor":
			listVendors()
		case "classes", "class":
			listDeviceClasses()
		case "specs", "spec":
			listSpecFiles(monitorCfg.verbose, monitorCfg.output)
		case "devices", "device":
			listDevices(monitorCfg.verbose, monitorCfg.output)
		case "all":
			listSpecFiles(monitorCfg.verbose, monitorCfg.output)
			listVendors()
			listDeviceClasses()
			listDevices(monitorCfg.verbose, monitorCfg.output)
		default:
			fmt.Printf("*** don't know how to list %q...\n", what)
		}
	}
}

var (
	monitorCfg monitorFlags
)

func init() {
	rootCmd.AddCommand(monitorCmd)
	monitorCmd.Flags().BoolVarP(&monitorCfg.verbose,
		"verbose", "v", false, "print details")
	monitorCmd.Flags().StringVarP(&monitorCfg.output,
		"output", "o", "", "output format for details (json|yaml)")
}

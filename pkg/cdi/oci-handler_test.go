/*
   Copyright © 2022 The CDI Authors

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

package cdi

import (
	"fmt"
	"testing"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/require"
	_ "github.com/stretchr/testify/require"
)

func TestLinuxDevice(t *testing.T) {
	var (
		devError = fmt.Errorf("device error")
		mntError = fmt.Errorf("mount error")
		device   bool
		mount    bool
		devOk    = func(*oci.LinuxDevice, *oci.LinuxDeviceCgroup) error {
			device = true
			return nil
		}
		devFail = func(*oci.LinuxDevice, *oci.LinuxDeviceCgroup) error {
			device = false
			return devError
		}
		mntOk = func(*oci.Mount) error {
			mount = true
			return nil
		}
		mntFail = func(*oci.Mount) error {
			mount = false
			return mntError
		}
	)
	type testCase struct {
		name   string
		h      *OciHandler
		device bool
		mount  bool
		err    error
	}
	for _, tc := range []*testCase{
		{
			name: "nil handler",
		},
		{
			name: "device only, OK",
			h: &OciHandler{
				LinuxDevice: devOk,
			},
			device: true,
		},
		{
			name: "mount only, OK",
			h: &OciHandler{
				Mount: mntOk,
			},
			mount: true,
		},
		{
			name: "device and mount, OK",
			h: &OciHandler{
				LinuxDevice: devOk,
				Mount:       mntOk,
			},
			device: true,
			mount:  true,
		},
		{
			name: "device and mount, device fails",
			h: &OciHandler{
				LinuxDevice: devFail,
				Mount:       mntOk,
			},
			err: devError,
		},
		{
			name: "device and mount, mount fails",
			h: &OciHandler{
				LinuxDevice: devOk,
				Mount:       mntFail,
			},
			err: mntError,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			device, mount = false, false
			err := tc.h.linuxDevice(nil, nil)
			if err == nil {
				err = tc.h.mount(nil)
			}
			if tc.err == nil {
				require.NoError(t, err, "device handler")
				require.Equal(t, tc.device, device, "device handler")
				require.Equal(t, tc.mount, mount, "mount handler")
			} else {
				require.Equal(t, tc.err, err)
			}
		})
	}
}

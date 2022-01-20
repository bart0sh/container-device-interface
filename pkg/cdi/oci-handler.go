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
	oci "github.com/opencontainers/runtime-spec/specs-go"
)

// OciHandlers are functions which allow altering OCI data
// prior to injection into an OCI Spec. These handlers are
// used by runtimes to adjust data in any runtime-specific
// way necessary or to carry out any extra processing once
// an OCI Spec has been updated with data for CDI devices.
type OciHandler struct {
	// LinuxDevice can modify devices prior to injection.
	LinuxDevice func(*oci.LinuxDevice, *oci.LinuxDeviceCgroup) error
	// Mount can modify mounts prior to injection.
	Mount func(*oci.Mount) error
}

func (h *OciHandler) linuxDevice(d *oci.LinuxDevice, dc *oci.LinuxDeviceCgroup) error {
	if h == nil || h.LinuxDevice == nil {
		return nil
	}
	return h.LinuxDevice(d, dc)
}

func (h *OciHandler) mount(m *oci.Mount) error {
	if h == nil || h.Mount == nil {
		return nil
	}
	return h.Mount(m)
}

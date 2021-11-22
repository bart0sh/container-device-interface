package cdi

import (
	"testing"

	"github.com/stretchr/testify/require"

	specs "github.com/container-orchestrated-devices/container-device-interface/specs-go"
)

var (
	invalidVersionSpec = `
cdiVersion: "0.0.0"
kind:       "vendor.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
        path: "/dev/dev1"
        type: "b"
        major: 666
        minor: 1
`
	noEditsSpec = `
cdiVersion: "0.2.0"
kind:       "vendor.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
        path: "/dev/dev1"
        type: "b"
        major: 666
        minor: 1
  - name: "dev2"
`
	emptyEditsSpec = `
cdiVersion: "0.2.0"
kind:       "vendor.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
        path: "/dev/dev1"
        type: "b"
        major: 666
        minor: 1
  - name: "dev2"
    containerEdits:
`
	deviceNameCollisionSpec = `
cdiVersion: "0.2.0"
kind:       "vendor.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev1"
        type: "b"
        major: 666
        minor: 1
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev2"
        type: "b"
        major: 666
        minor: 2
  - name: "dev3"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev1"
        type: "b"
        major: 666
        minor: 3
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev2"
        type: "b"
        major: 666
        minor: 2
`
	vendorAFooSpec1234 = `
cdiVersion: "0.2.0"
kind:       "vendorA.com/foo"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev1"
        type: "b"
        major: 666
        minor: 1
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev2"
        type: "b"
        major: 666
        minor: 2
  - name: "dev3"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev3"
        type: "b"
        major: 666
        minor: 3
  - name: "dev4"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev4"
        type: "b"
        major: 666
        minor: 4
`
	vendorBBarSpec3241 = `
cdiVersion: "0.2.0"
kind:       "vendorB.org/bar"
devices:
  - name: "dev3"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev3"
        type: "b"
        major: 666
        minor: 3
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev2"
        type: "b"
        major: 666
        minor: 2
  - name: "dev4"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev4"
        type: "b"
        major: 666
        minor: 4
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/dev1"
        type: "b"
        major: 666
        minor: 1
`
)

func TestLoadSpec(t *testing.T) {
	type testCase struct {
		name       string
		data       string
		unparsable bool
		invalid    bool
		vendor     string
		class      string
		devices    []string
		qualified  []string
	}
	for _, tc := range []*testCase{
		{
			name:       "with invalid version",
			data:       invalidVersionSpec,
			unparsable: true,
		},
		{
			name:       "with device lacking edits",
			data:       noEditsSpec,
			unparsable: true,
		},
		{
			name:       "with device with empty edits",
			data:       emptyEditsSpec,
			unparsable: true,
		},
		{
			name:    "with device name collision",
			data:    deviceNameCollisionSpec,
			invalid: true,
		},
		{
			name:    "with vendor A, devices foo 1, 2, 3, 4",
			data:    vendorAFooSpec1234,
			devices: []string{"dev1", "dev2", "dev3", "dev4"},
			qualified: []string{
				"vendorA.com/foo=dev1",
				"vendorA.com/foo=dev2",
				"vendorA.com/foo=dev3",
				"vendorA.com/foo=dev4",
			},
		},
		{
			name:    "with vendor B, devices bar 3, 2, 4, 1",
			data:    vendorBBarSpec3241,
			devices: []string{"dev1", "dev2", "dev3", "dev4"},
			qualified: []string{
				"vendorB.org/bar=dev1",
				"vendorB.org/bar=dev2",
				"vendorB.org/bar=dev3",
				"vendorB.org/bar=dev4",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				raw     *specs.Spec
				spec    *Spec
				devices []string
				err     error
			)

			raw, err = parseSpecData([]byte(tc.data))
			if tc.unparsable {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			spec, err = NewSpec(raw, tc.name, 1)
			if tc.invalid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			devices, err = spec.ListDeviceNames()
			require.NoError(t, err)
			require.Equal(t, devices, tc.devices)

			devices, err = spec.ListQualifiedDeviceNames()
			require.NoError(t, err)
			require.Equal(t, devices, tc.qualified)
		})
	}
}

func TestGetVendorAndClass(t *testing.T) {
	type testCase struct {
		name   string
		data   string
		vendor string
		class  string
	}
	for _, tc := range []*testCase{
		{
			name:   "for vendor A with class foo",
			data:   vendorAFooSpec1234,
			vendor: "vendorA.com",
			class:  "foo",
		},
		{
			name:   "for vendor B with class bar",
			data:   vendorBBarSpec3241,
			vendor: "vendorB.org",
			class:  "bar",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				raw    *specs.Spec
				spec   *Spec
				vendor string
				class  string
				err    error
			)

			raw, err = parseSpecData([]byte(tc.data))
			require.NoError(t, err)

			spec, err = NewSpec(raw, tc.name, 1)
			require.NoError(t, err)

			vendor, class = spec.GetVendor(), spec.GetClass()
			require.NoError(t, err)
			require.Equal(t, vendor, tc.vendor)
			require.Equal(t, class, tc.class)
		})
	}
}

func TestListDeviceNames(t *testing.T) {
	type testCase struct {
		name      string
		data      string
		devices   []string
		qualified []string
	}
	for _, tc := range []*testCase{
		{
			name:    "for vendor A, devices foo 1, 2, 3, 4",
			data:    vendorAFooSpec1234,
			devices: []string{"dev1", "dev2", "dev3", "dev4"},
			qualified: []string{
				"vendorA.com/foo=dev1",
				"vendorA.com/foo=dev2",
				"vendorA.com/foo=dev3",
				"vendorA.com/foo=dev4",
			},
		},
		{
			name:    "for vendor B, devices bar 3, 2, 4, 1",
			data:    vendorBBarSpec3241,
			devices: []string{"dev1", "dev2", "dev3", "dev4"},
			qualified: []string{
				"vendorB.org/bar=dev1",
				"vendorB.org/bar=dev2",
				"vendorB.org/bar=dev3",
				"vendorB.org/bar=dev4",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				raw     *specs.Spec
				spec    *Spec
				devices []string
				err     error
			)

			raw, err = parseSpecData([]byte(tc.data))
			require.NoError(t, err)

			spec, err = NewSpec(raw, tc.name, 1)
			require.NoError(t, err)

			devices, err = spec.ListDeviceNames()
			require.NoError(t, err)
			require.Equal(t, devices, tc.devices)

			devices, err = spec.ListQualifiedDeviceNames()
			require.NoError(t, err)
			require.Equal(t, devices, tc.qualified)
		})
	}
}

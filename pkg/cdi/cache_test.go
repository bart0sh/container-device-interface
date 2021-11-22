package cdi

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"testing"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var (
	vendorA        = "vendorA.com"
	kindAFoo       = vendorA + "/foo"
	vendorAFoo1234 = `
cdiVersion: "0.2.0"
kind:       "vendorA.com/foo"
devices:
  - name: "dev1"
    containerEdits:
      env:
      - "A_FOO_1=1234"
      deviceNodes:
      - path: "/dev/foo1"
        type: "b"
        major: 666
        minor: 1
  - name: "dev2"
    containerEdits:
      env:
      - "A_FOO_2=1234"
      deviceNodes:
      - path: "/dev/foo2"
        type: "b"
        major: 666
        minor: 2
  - name: "dev3"
    containerEdits:
      env:
      - "A_FOO_3=1234"
      deviceNodes:
      - path: "/dev/foo3"
        type: "b"
        major: 666
        minor: 3
  - name: "dev4"
    containerEdits:
      env:
      - "A_FOO_4=1234"
      deviceNodes:
      - path: "/dev/foo4"
        type: "b"
        major: 666
        minor: 4
`
	vendorAFoo3156 = `
cdiVersion: "0.2.0"
kind:       "vendorA.com/foo"
devices:
  - name: "dev3"
    containerEdits:
      env:
      - "A_FOO_3=3156"
      deviceNodes:
      - path: "/dev/foo3"
        type: "b"
        major: 666
        minor: 3
  - name: "dev1"
    containerEdits:
      env:
      - "A_FOO_1=3156"
      deviceNodes:
      - path: "/dev/foo1"
        type: "b"
        major: 666
        minor: 1
  - name: "dev5"
    containerEdits:
      env:
      - "A_FOO_5=3156"
      deviceNodes:
      - path: "/dev/foo5"
        type: "b"
        major: 666
        minor: 5
  - name: "dev6"
    containerEdits:
      env:
      - "A_FOO_6=3156"
      deviceNodes:
      - path: "/dev/foo6"
        type: "b"
        major: 666
        minor: 6
`
	vendorB        = "vendorB.org"
	kindBBar       = vendorB + "/bar"
	vendorBBar3241 = `
cdiVersion: "0.2.0"
kind:       "vendorB.org/bar"
devices:
  - name: "dev3"
    containerEdits:
      env:
      - "B_BAR_3=3241"
      deviceNodes:
      - path: "/dev/bar3"
        type: "b"
        major: 667
        minor: 3
  - name: "dev2"
    containerEdits:
      env:
      - "B_BAR_2=3241"
      deviceNodes:
      - path: "/dev/bar2"
        type: "b"
        major: 667
        minor: 2
  - name: "dev4"
    containerEdits:
      env:
      - "B_BAR_4=3241"
      deviceNodes:
      - path: "/dev/bar4"
        type: "b"
        major: 667
        minor: 4
  - name: "dev1"
    containerEdits:
      env:
      - "B_BAR_1=3241"
      deviceNodes:
      - path: "/dev/bar1"
        type: "b"
        major: 667
        minor: 1
`
	vendorBBar2571 = `
cdiVersion: "0.2.0"
kind:       "vendorB.org/bar"
devices:
  - name: "dev2"
    containerEdits:
      env:
      - "B_BAR_2=2571"
      deviceNodes:
      - path: "/dev/bar2"
        type: "b"
        major: 667
        minor: 2
  - name: "dev5"
    containerEdits:
      env:
      - "B_BAR_5=2571"
      deviceNodes:
      - path: "/dev/bar5"
        type: "b"
        major: 667
        minor: 5
  - name: "dev7"
    containerEdits:
      env:
      - "B_BAR_7=2571"
      deviceNodes:
      - path: "/dev/bar7"
        type: "b"
        major: 667
        minor: 7
  - name: "dev1"
    containerEdits:
      env:
      - "B_BAR_1=2571"
      deviceNodes:
      - path: "/dev/bar1"
        type: "b"
        major: 667
        minor: 1
`
	vendorC          = "vendorC.net"
	kindCXyzzy       = vendorC + "/xyzzy"
	vendorCXyzzy2741 = `
cdiVersion: "0.2.0"
kind:       "vendorC.net/xyzzy"
devices:
  - name: "dev2"
    containerEdits:
      env:
      - "C_XYZZY_2=2741"
      deviceNodes:
      - path: "/dev/xyzzy3"
        type: "b"
        major: 668
        minor: 2
  - name: "dev7"
    containerEdits:
      env:
      - "C_XYZZY_7=2741"
      deviceNodes:
      - path: "/dev/xyzzy7"
        type: "b"
        major: 668
        minor: 7
  - name: "dev4"
    containerEdits:
      env:
      - "C_XYZZY_4=2741"
      deviceNodes:
      - path: "/dev/xyzzy4"
        type: "b"
        major: 668
        minor: 4
  - name: "dev1"
    containerEdits:
      env:
      - "C_XYZZY_1=2741"
      deviceNodes:
      - path: "/dev/xyzzy1"
        type: "b"
        major: 668
        minor: 1
`
)

type cacheLoadTest struct {
	name   string
	etc    []string
	run    []string
	errors int
	dir    *tmpDir
	cache  *Cache
}

type cacheVendorTest struct {
	cacheLoadTest
	vendors []string
}

type cacheDeviceTest struct {
	cacheLoadTest
	devices []string
	devprio []int
	matches map[string][]string
}

type cacheRefreshTest struct {
	name    string
	updates []specSet
	devices [][]string
	devprio [][]int
	dir     *tmpDir
	cache   *Cache
}

type specSet struct {
	etc []string
	run []string
}

type cacheInjectTest struct {
	cacheLoadTest
	ociSpec     *oci.Spec
	devices     []string
	updated     *oci.Spec
	unresolved  []string
	expectError bool
}

func TestCacheLoadSpecs(t *testing.T) {
	for _, tc := range []*cacheLoadTest{
		{
			name: "etc:A-1234",
			etc:  []string{vendorAFoo1234},
		},
		{
			name:   "etc:A-1234 twice",
			etc:    []string{vendorAFoo1234, vendorAFoo1234},
			errors: 4,
		},
		{
			name:   "etc:A-1234,A-3156",
			etc:    []string{vendorAFoo1234, vendorAFoo3156},
			errors: 2,
		},
		{
			name: "etc:A-1234, run:A-3156",
			etc:  []string{vendorAFoo1234},
			run:  []string{vendorAFoo3156},
		},
		{
			name: "etc:A-1234,B-3241, run:A-3156,B-2571",
			etc:  []string{vendorAFoo1234, vendorBBar3241},
			run:  []string{vendorAFoo3156, vendorBBar2571},
		},
		{
			name:   "etc:A-1234,A-3156, run:B-3241,B-2571",
			etc:    []string{vendorAFoo1234, vendorAFoo3156},
			run:    []string{vendorBBar3241, vendorBBar2571},
			errors: 2 + 2,
		},
	} {
		testLoadCache(t, tc)
	}

	return
}

func TestCacheListVendors(t *testing.T) {
	for _, tc := range []*cacheVendorTest{
		{
			cacheLoadTest: cacheLoadTest{
				name: "etc:A-1234",
				etc:  []string{vendorAFoo1234},
			},
			vendors: []string{vendorA},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name:   "etc:A-1234 twice",
				etc:    []string{vendorAFoo1234, vendorAFoo1234},
				errors: 4,
			},
			vendors: []string{vendorA},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name:   "etc:A-1234,A-3156",
				etc:    []string{vendorAFoo1234, vendorAFoo3156},
				errors: 2,
			},
			vendors: []string{vendorA},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name: "etc:A-1234, run:A-3156",
				etc:  []string{vendorAFoo1234},
				run:  []string{vendorAFoo3156},
			},
			vendors: []string{vendorA},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name: "etc:A-1234,B-3241, run:A-3156,B-2571",
				etc:  []string{vendorAFoo1234, vendorBBar3241},
				run:  []string{vendorAFoo3156, vendorBBar2571},
			},
			vendors: []string{vendorA, vendorB},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name:   "etc:A-1234,A-3156, run:B-3241,B-2571",
				etc:    []string{vendorAFoo1234, vendorAFoo3156},
				run:    []string{vendorBBar3241, vendorBBar2571, vendorCXyzzy2741},
				errors: 2 + 2,
			},
			vendors: []string{vendorA, vendorB, vendorC},
		},
	} {
		testLoadCache(t, &tc.cacheLoadTest)
		vendors := tc.cache.ListVendors()
		require.Equal(t, tc.vendors, vendors, tc.name)
	}
}

func TestCacheListDevices(t *testing.T) {
	for _, tc := range []*cacheDeviceTest{
		{
			cacheLoadTest: cacheLoadTest{
				name: "etc:A-1234",
				etc:  []string{vendorAFoo1234},
			},
			devices: []string{
				kindAFoo + "=dev1",
				kindAFoo + "=dev2",
				kindAFoo + "=dev3",
				kindAFoo + "=dev4",
			},
			devprio: []int{
				0, 0, 0, 0,
			},
			matches: map[string][]string{
				vendorA + "/*=*": {
					kindAFoo + "=dev1",
					kindAFoo + "=dev2",
					kindAFoo + "=dev3",
					kindAFoo + "=dev4",
				},
				"*/*=dev[13]": {
					kindAFoo + "=dev1",
					kindAFoo + "=dev3",
				},
			},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name:   "etc:A-1234 twice",
				etc:    []string{vendorAFoo1234, vendorAFoo1234},
				errors: 4,
			},
			devices: nil,
		},
		{
			cacheLoadTest: cacheLoadTest{
				name:   "etc:A-1234,A-3156",
				etc:    []string{vendorAFoo1234, vendorAFoo3156},
				errors: 2,
			},
			devices: []string{
				kindAFoo + "=dev2",
				kindAFoo + "=dev4",
				kindAFoo + "=dev5",
				kindAFoo + "=dev6",
			},
			devprio: []int{
				0, 0, 0, 0,
			},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name: "etc:A-1234, run:A-3156",
				etc:  []string{vendorAFoo1234},
				run:  []string{vendorAFoo3156},
			},
			devices: []string{
				kindAFoo + "=dev1",
				kindAFoo + "=dev2",
				kindAFoo + "=dev3",
				kindAFoo + "=dev4",
				kindAFoo + "=dev5",
				kindAFoo + "=dev6",
			},
			devprio: []int{
				1, 0, 1, 0, 1, 1,
			},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name: "etc:A-1234,B-3241, run:A-3156,B-2571",
				etc:  []string{vendorAFoo1234, vendorBBar3241},
				run:  []string{vendorAFoo3156, vendorBBar2571},
			},
			devices: []string{
				kindAFoo + "=dev1",
				kindAFoo + "=dev2",
				kindAFoo + "=dev3",
				kindAFoo + "=dev4",
				kindAFoo + "=dev5",
				kindAFoo + "=dev6",
				kindBBar + "=dev1",
				kindBBar + "=dev2",
				kindBBar + "=dev3",
				kindBBar + "=dev4",
				kindBBar + "=dev5",
				kindBBar + "=dev7",
			},
			devprio: []int{
				1, 0, 1, 0, 1, 1,
				1, 1, 0, 0, 1, 1,
			},
			matches: map[string][]string{
				"*/bar=*[357]": {
					kindBBar + "=dev3",
					kindBBar + "=dev5",
					kindBBar + "=dev7",
				},
			},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name:   "etc:A-1234,A-3156, run:B-3241,B-2571,C-2741",
				etc:    []string{vendorAFoo1234, vendorAFoo3156},
				run:    []string{vendorBBar3241, vendorBBar2571, vendorCXyzzy2741},
				errors: 2 + 2,
			},
			devices: []string{
				kindAFoo + "=dev2",
				kindAFoo + "=dev4",
				kindAFoo + "=dev5",
				kindAFoo + "=dev6",
				kindBBar + "=dev3",
				kindBBar + "=dev4",
				kindBBar + "=dev5",
				kindBBar + "=dev7",
				kindCXyzzy + "=dev1",
				kindCXyzzy + "=dev2",
				kindCXyzzy + "=dev4",
				kindCXyzzy + "=dev7",
			},
			devprio: []int{
				0, 0, 0, 0,
				1, 1, 1, 1,
				1, 1, 1, 1,
			},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name: "etc:A-1234,B-3241, run:A-3156,B-2571,C-2741",
				etc:  []string{vendorAFoo1234, vendorBBar3241},
				run:  []string{vendorAFoo3156, vendorBBar2571, vendorCXyzzy2741},
			},
			devices: []string{
				kindAFoo + "=dev1",
				kindAFoo + "=dev2",
				kindAFoo + "=dev3",
				kindAFoo + "=dev4",
				kindAFoo + "=dev5",
				kindAFoo + "=dev6",
				kindBBar + "=dev1",
				kindBBar + "=dev2",
				kindBBar + "=dev3",
				kindBBar + "=dev4",
				kindBBar + "=dev5",
				kindBBar + "=dev7",
				kindCXyzzy + "=dev1",
				kindCXyzzy + "=dev2",
				kindCXyzzy + "=dev4",
				kindCXyzzy + "=dev7",
			},
			devprio: []int{
				1, 0, 1, 0, 1, 1,
				1, 1, 0, 0, 1, 1,
				1, 1, 1, 1,
			},
		},
	} {
		testLoadCache(t, &tc.cacheLoadTest)
		devices := tc.cache.ListDevices()
		require.Equal(t, tc.devices, devices, tc.name)
		for idx, device := range devices {
			dev := tc.cache.GetDevice(device)
			require.NotNil(t, dev, tc.name)
			require.Equal(t, tc.devprio[idx], dev.getPriority(), tc.name)
		}
		for pattern, expected := range tc.matches {
			matches, err := tc.cache.MatchDevices(pattern)
			require.Nil(t, err, tc.name)
			require.Equal(t, expected, matches, tc.name)
		}
	}
}

func TestCacheRefreshDevices(t *testing.T) {
	for _, tc := range []*cacheRefreshTest{
		{
			name: "none - etc:+A-1234 - etc:+B-3241",
			updates: []specSet{
				{},
				{etc: []string{vendorAFoo1234}},
				{etc: []string{vendorBBar3241}},
			},
			devices: [][]string{
				nil,
				{
					kindAFoo + "=dev1",
					kindAFoo + "=dev2",
					kindAFoo + "=dev3",
					kindAFoo + "=dev4",
				},
				{
					kindBBar + "=dev1",
					kindBBar + "=dev2",
					kindBBar + "=dev3",
					kindBBar + "=dev4",
				},
			},
			devprio: [][]int{
				nil,
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
		},
		{
			name: "none - etc:+A-1234 - run:+B-3241",
			updates: []specSet{
				{},
				{etc: []string{vendorAFoo1234}},
				{run: []string{vendorBBar3241}},
			},
			devices: [][]string{
				nil,
				{
					kindAFoo + "=dev1",
					kindAFoo + "=dev2",
					kindAFoo + "=dev3",
					kindAFoo + "=dev4",
				},
				{
					kindAFoo + "=dev1",
					kindAFoo + "=dev2",
					kindAFoo + "=dev3",
					kindAFoo + "=dev4",
					kindBBar + "=dev1",
					kindBBar + "=dev2",
					kindBBar + "=dev3",
					kindBBar + "=dev4",
				},
			},
			devprio: [][]int{
				nil,
				{0, 0, 0, 0},
				{0, 0, 0, 0, 1, 1, 1, 1},
			},
		},
		{
			name: "none - etc:+A-1234 - etc:+A-3156",
			updates: []specSet{
				{},
				{etc: []string{vendorAFoo1234}},
				{etc: []string{vendorAFoo3156}},
			},
			devices: [][]string{
				nil,
				{
					kindAFoo + "=dev1",
					kindAFoo + "=dev2",
					kindAFoo + "=dev3",
					kindAFoo + "=dev4",
				},
				{
					kindAFoo + "=dev1",
					kindAFoo + "=dev3",
					kindAFoo + "=dev5",
					kindAFoo + "=dev6",
				},
			},
			devprio: [][]int{
				nil,
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
		},
		{
			name: "none - etc:+A-1234 - etc:skip,+A-3156",
			updates: []specSet{
				{},
				{etc: []string{vendorAFoo1234}},
				{etc: []string{"skip", vendorAFoo3156}},
			},
			devices: [][]string{
				nil,
				{
					kindAFoo + "=dev1",
					kindAFoo + "=dev2",
					kindAFoo + "=dev3",
					kindAFoo + "=dev4",
				},
				{
					kindAFoo + "=dev2",
					kindAFoo + "=dev4",
					kindAFoo + "=dev5",
					kindAFoo + "=dev6",
				},
			},
			devprio: [][]int{
				nil,
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
		},
		{
			name: "none - etc:+A-1234 - run:+A-3156",
			updates: []specSet{
				{},
				{etc: []string{vendorAFoo1234}},
				{run: []string{vendorAFoo3156}},
			},
			devices: [][]string{
				nil,
				{
					kindAFoo + "=dev1",
					kindAFoo + "=dev2",
					kindAFoo + "=dev3",
					kindAFoo + "=dev4",
				},
				{
					kindAFoo + "=dev1",
					kindAFoo + "=dev2",
					kindAFoo + "=dev3",
					kindAFoo + "=dev4",
					kindAFoo + "=dev5",
					kindAFoo + "=dev6",
				},
			},
			devprio: [][]int{
				nil,
				{0, 0, 0, 0},
				{1, 0, 1, 0, 1, 1},
			},
		},
	} {
		var err error

		t.Run(tc.name, func(t *testing.T) {
			for idx, specs := range tc.updates {
				if idx == 0 {
					tc.dir, tc.cache, err = populateTestCache(t, specs.etc, specs.run)
					if tc.cache == nil && err != nil {
						t.Errorf("failed to load cache: %v", err)
						return
					}
				} else {
					err = tc.dir.refreshCache(tc.cache, specs.etc, specs.run)
				}
				devices := tc.cache.ListDevices()
				require.Equal(t, tc.devices[idx], devices, tc.name)
				for devIdx, device := range devices {
					dev := tc.cache.GetDevice(device)
					require.NotNil(t, dev, tc.name)
					require.Equal(t, tc.devprio[idx][devIdx], dev.getPriority(), tc.name)
				}
			}
		})
	}
}

func TestCacheInjectDevices(t *testing.T) {
	for _, tc := range []*cacheInjectTest{
		{
			cacheLoadTest: cacheLoadTest{
				name: "one device into empty OCI Spec",
				etc:  []string{vendorAFoo1234},
			},
			ociSpec: &oci.Spec{},
			devices: []string{
				kindAFoo + "=dev1",
			},
			updated: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"A_FOO_1=1234",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path:  "/dev/foo1",
							Type:  "b",
							Major: 666,
							Minor: 1,
						},
					},
				},
			},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name: "two devices into empty OCI Spec",
				etc:  []string{vendorAFoo1234},
			},
			ociSpec: &oci.Spec{},
			devices: []string{
				kindAFoo + "=dev1",
				kindAFoo + "=dev2",
			},
			updated: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"A_FOO_1=1234",
						"A_FOO_2=1234",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path:  "/dev/foo1",
							Type:  "b",
							Major: 666,
							Minor: 1,
						},
						{
							Path:  "/dev/foo2",
							Type:  "b",
							Major: 666,
							Minor: 2,
						},
					},
				},
			},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name: "three devices into empty OCI Spec, one unresolvable",
				etc:  []string{vendorAFoo1234},
			},
			ociSpec: &oci.Spec{},
			devices: []string{
				kindAFoo + "=dev1",
				kindAFoo + "=dev2",
				kindAFoo + "=no-such-dev",
			},
			updated: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"A_FOO_1=1234",
						"A_FOO_2=1234",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path:  "/dev/foo1",
							Type:  "b",
							Major: 666,
							Minor: 1,
						},
						{
							Path:  "/dev/foo2",
							Type:  "b",
							Major: 666,
							Minor: 2,
						},
					},
				},
			},
			unresolved: []string{
				kindAFoo + "=no-such-dev",
			},
		},
	} {
		var (
			unresolved []string
			err        error
		)
		testLoadCache(t, &tc.cacheLoadTest)
		unresolved, err = tc.cache.InjectDevices(tc.ociSpec, tc.devices)
		if !tc.expectError {
			require.Nil(t, err, tc.name)
			require.Equal(t, tc.updated, tc.ociSpec)
			require.Equal(t, tc.unresolved, unresolved, tc.name)
		} else {
			require.NotNil(t, err, tc.name)
		}
	}
}

func TestCacheResolveDevices(t *testing.T) {
	for _, tc := range []*cacheInjectTest{
		{
			cacheLoadTest: cacheLoadTest{
				name: "one device",
				etc:  []string{vendorAFoo1234},
			},
			ociSpec: &oci.Spec{
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{Path: kindAFoo + "=dev1"},
					},
				},
			},
			updated: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"A_FOO_1=1234",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path:  "/dev/foo1",
							Type:  "b",
							Major: 666,
							Minor: 1,
						},
					},
				},
			},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name: "two devices",
				etc:  []string{vendorAFoo1234},
			},
			ociSpec: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"FOO=BAR",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{Path: kindAFoo + "=dev1"},
						{Path: kindAFoo + "=dev2"},
					},
				},
			},
			updated: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"FOO=BAR",
						"A_FOO_1=1234",
						"A_FOO_2=1234",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path:  "/dev/foo1",
							Type:  "b",
							Major: 666,
							Minor: 1,
						},
						{
							Path:  "/dev/foo2",
							Type:  "b",
							Major: 666,
							Minor: 2,
						},
					},
				},
			},
		},
		{
			cacheLoadTest: cacheLoadTest{
				name: "three devices, one unresolvable",
				etc:  []string{vendorAFoo1234},
				run:  []string{vendorBBar2571},
			},
			ociSpec: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"BAR=FOO",
						"XYZZY=FOOBAR",
						"FROB=NICATE",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{Path: kindAFoo + "=dev1"},
						{Path: kindBBar + "=dev7"},
						{Path: kindAFoo + "=no-such-dev"},
					},
				},
			},
			updated: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"BAR=FOO",
						"XYZZY=FOOBAR",
						"FROB=NICATE",
						"A_FOO_1=1234",
						"B_BAR_7=2571",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path:  "/dev/foo1",
							Type:  "b",
							Major: 666,
							Minor: 1,
						},
						{
							Path:  "/dev/bar7",
							Type:  "b",
							Major: 667,
							Minor: 7,
						},
						{Path: kindAFoo + "=no-such-dev"},
					},
				},
			},
			unresolved: []string{
				kindAFoo + "=no-such-dev",
			},
		},
	} {
		var (
			unresolved []string
			err        error
		)
		testLoadCache(t, &tc.cacheLoadTest)
		unresolved, err = tc.cache.ResolveDevices(tc.ociSpec)
		if !tc.expectError {
			require.Nil(t, err, tc.name)
			require.Equal(t, tc.updated, tc.ociSpec)
			require.Equal(t, tc.unresolved, unresolved, tc.name)
		} else {
			require.NotNil(t, err, tc.name)
		}
	}
}

func testLoadCache(t *testing.T, tc *cacheLoadTest) {
	t.Run(tc.name, func(t *testing.T) {
		var err error

		tc.dir, tc.cache, err = populateTestCache(t, tc.etc, tc.run)
		if tc.cache == nil && err != nil {
			t.Errorf("failed to load cache: %v", err)
			return
		}

		if tc.errors == 0 {
			require.Nil(t, err, tc.name)
		} else {
			require.NotNil(t, err, tc.name)
		}

		require.NotNil(t, tc.cache)

		if tc.errors != 0 {
			errors := tc.cache.Errors()
			require.NotEqual(t, len(errors), 0, tc.name)
			if len(errors) != tc.errors {
				t.Logf("WARNING: expected %d warnings, got %d\n",
					tc.errors, len(errors))
			}
		} else {
			require.Equal(t, len(tc.cache.Errors()), 0, tc.name)
		}
	})
}

func populateTestCache(t *testing.T, etc, run []string) (*tmpDir, *Cache, error) {
	var (
		dir   *tmpDir
		cache *Cache
		err   error
	)

	dir, err = mkTmpDir(t, etc, run)
	if err != nil {
		return nil, nil, err
	}

	cache, err = NewCache(
		WithSpecDirs(
			filepath.Join(dir.root, "etc"),
			filepath.Join(dir.root, "run"),
		),
	)

	return dir, cache, err
}

type tmpDir struct {
	root string
	etc  string
	run  string
}

func mkTmpDir(t *testing.T, etcSpecs, runSpecs []string) (dir *tmpDir, retErr error) {
	var (
		tmp string
		err error
	)

	if tmp, err = tempDirForTest(t); err != nil {
		return nil, err
	}
	dir = &tmpDir{
		root: tmp,
		etc:  filepath.Join(tmp, "etc"),
		run:  filepath.Join(tmp, "run"),
	}

	if err = os.MkdirAll(dir.etc, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create temporary 'etc' dir")
	}
	if err = os.MkdirAll(dir.run, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create temporary 'run' dir")
	}
	if err := dir.update(etcSpecs, runSpecs); err != nil {
		return nil, err
	}

	return dir, nil
}

func (dir *tmpDir) update(etcSpecs, runSpecs []string) error {
	for idx, data := range etcSpecs {
		path := filepath.Join(dir.etc, fmt.Sprintf("spec-%d.yaml", idx))
		switch {
		case string(data) == "remove":
			os.Remove(path)
		case string(data) == "skip":
		default:
			if err := ioutil.WriteFile(path+".tmp", []byte(data), 0644); err != nil {
				return errors.Wrap(err, "failed to write temporary etc data")
			}
			if err := os.Rename(path+".tmp", path); err != nil {
				return errors.Wrap(err, "failed to create temporary etc file")
			}
		}
	}
	for idx, data := range runSpecs {
		path := filepath.Join(dir.run, fmt.Sprintf("spec-%d.yaml", idx))
		switch {
		case string(data) == "remove":
			os.Remove(path)
		case string(data) == "skip":
		default:
			if err := ioutil.WriteFile(path+".tmp", []byte(data), 0644); err != nil {
				return errors.Wrap(err, "failed to write temporary run data")
			}
			if err := os.Rename(path+".tmp", path); err != nil {
				return errors.Wrap(err, "failed to create temporary etc file")
			}
		}
	}
	return nil
}

func (dir *tmpDir) refreshCache(cache *Cache, etcSpecs, runSpecs []string) error {
	if err := dir.update(etcSpecs, runSpecs); err != nil {
		return err
	}
	return cache.Refresh()
}

func tempDirForTest(t *testing.T) (string, error) {
	tmp, err := ioutil.TempDir("", ".cache-test")
	if err != nil {
		return "", errors.Wrapf(err, "failed to create temporary test directory")
	}

	t.Cleanup(func() {
		os.RemoveAll(tmp)
	})

	return tmp, nil
}

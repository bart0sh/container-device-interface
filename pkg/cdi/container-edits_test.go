package cdi

import (
	"testing"

	cdi "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/require"
)

func TestVlaidateContainerEdits(t *testing.T) {
	testCases := []struct {
		name    string
		config  *oci.Spec
		edits   *cdi.ContainerEdits
		invalid bool
	}{
		{
			name:  "valid, empty edits",
			edits: nil,
		},
		{
			name: "valid, env var",
			edits: &cdi.ContainerEdits{
				Env: []string{"BAR=BARVALUE1"},
			},
		},
		{
			name: "invalid env, empty var",
			edits: &cdi.ContainerEdits{
				Env: []string{""},
			},
			invalid: true,
		},
		{
			name: "invalid env, no var name",
			edits: &cdi.ContainerEdits{
				Env: []string{"=foo"},
			},
			invalid: true,
		},
		{
			name: "invalid env, no assignment",
			edits: &cdi.ContainerEdits{
				Env: []string{"FOOBAR"},
			},
			invalid: true,
		},
		{
			name: "valid device, path only",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/null",
					},
				},
			},
		},
		{
			name: "valid device, path+type",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/null",
						Type: "b",
					},
				},
			},
		},
		{
			name: "valid device, path+type+permissions",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:        "/dev/null",
						Type:        "b",
						Permissions: "rwm",
					},
				},
			},
		},
		{
			name: "invalid device, empty path",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid device, wrong type",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/vendorctl",
						Type: "f",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid device, wrong permissions",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:        "/dev/vendorctl",
						Type:        "b",
						Permissions: "to land",
					},
				},
			},
			invalid: true,
		},
		{
			name: "valid mount",
			edits: &cdi.ContainerEdits{
				Mounts: []*cdi.Mount{
					{
						HostPath:      "/dev/vendorctl",
						ContainerPath: "/dev/vendorctl",
					},
				},
			},
		},
		{
			name: "invalid mount, empty host path",
			edits: &cdi.ContainerEdits{
				Mounts: []*cdi.Mount{
					{
						HostPath:      "",
						ContainerPath: "/dev/vendorctl",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid mount, empty container path",
			edits: &cdi.ContainerEdits{
				Mounts: []*cdi.Mount{
					{
						HostPath:      "/dev/vendorctl",
						ContainerPath: "",
					},
				},
			},
			invalid: true,
		},
		{
			name: "valid hooks",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "prestart",
						Path:     "/usr/local/bin/prestart-vendor-hook",
						Args:     []string{"--verbose"},
						Env:      []string{"VENDOR_ENV1=value1"},
					},
					{
						HookName: "createRuntime",
						Path:     "/usr/local/bin/cr-vendor-hook",
						Args:     []string{"--debug"},
						Env:      []string{"VENDOR_ENV2=value2"},
					},
					{
						HookName: "createContainer",
						Path:     "/usr/local/bin/cc-vendor-hook",
						Args:     []string{"--create"},
						Env:      []string{"VENDOR_ENV3=value3"},
					},
					{
						HookName: "startContainer",
						Path:     "/usr/local/bin/sc-vendor-hook",
						Args:     []string{"--start"},
						Env:      []string{"VENDOR_ENV4=value4"},
					},
					{
						HookName: "poststart",
						Path:     "/usr/local/bin/poststart-vendor-hook",
						Env:      []string{"VENDOR_ENV5=value5"},
					},
					{
						HookName: "poststop",
						Path:     "/usr/local/bin/poststop-vendor-hook",
					},
				},
			},
		},
		{
			name: "invalid hook, empty path",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "prestart",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid hook, wrong hook name",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "misCreateRuntime",
						Path:     "/usr/local/bin/cr-vendor-hook",
						Args:     []string{"--debug"},
						Env:      []string{"VENDOR_ENV2=value2"},
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid hook, wrong env",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "poststart",
						Path:     "/usr/local/bin/cr-vendor-hook",
						Args:     []string{"--debug"},
						Env:      []string{"=value2"},
					},
				},
			},
			invalid: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			edits := ContainerEdits{tc.edits}
			err := edits.Validate()
			if tc.invalid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestApplyContainerEdits(t *testing.T) {
	testCases := []struct {
		name           string
		config         *oci.Spec
		edits          *cdi.ContainerEdits
		expectedResult oci.Spec
		expectedError  bool
	}{
		{
			name:   "nil edits",
			config: &oci.Spec{},
			edits:  nil,
		},
		{
			name:   "add env to the empty spec",
			config: &oci.Spec{},
			edits: &cdi.ContainerEdits{
				Env: []string{"BAR=BARVALUE1"},
			},
			expectedResult: oci.Spec{
				Process: &oci.Process{
					Env: []string{"BAR=BARVALUE1"},
				},
			},
		},
		{
			name:   "add device nodes to the empty spec",
			config: &oci.Spec{},
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/vendorctl",
					},
				},
			},
			expectedResult: oci.Spec{
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{Path: "/dev/vendorctl"},
					},
				},
			},
		},
		{
			name:   "add mounts to the empty spec",
			config: &oci.Spec{},
			edits: &cdi.ContainerEdits{
				Mounts: []*cdi.Mount{
					{
						HostPath:      "/dev/vendorctl",
						ContainerPath: "/dev/vendorctl",
					},
				},
			},
			expectedResult: oci.Spec{
				Mounts: []oci.Mount{
					{
						Source:      "/dev/vendorctl",
						Destination: "/dev/vendorctl",
					},
				},
			},
		},
		{
			name:   "add hooks to the empty spec",
			config: &oci.Spec{},
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "prestart",
						Path:     "/usr/local/bin/prestart-vendor-hook",
						Args:     []string{"--verbose"},
						Env:      []string{"VENDOR_ENV1=value1"},
					},
					{
						HookName: "createRuntime",
						Path:     "/usr/local/bin/cr-vendor-hook",
						Args:     []string{"--debug"},
						Env:      []string{"VENDOR_ENV2=value2"},
					},
					{
						HookName: "createContainer",
						Path:     "/usr/local/bin/cc-vendor-hook",
						Args:     []string{"--create"},
						Env:      []string{"VENDOR_ENV3=value3"},
					},
					{
						HookName: "startContainer",
						Path:     "/usr/local/bin/sc-vendor-hook",
						Args:     []string{"--start"},
						Env:      []string{"VENDOR_ENV4=value4"},
					},
					{
						HookName: "poststart",
						Path:     "/usr/local/bin/poststart-vendor-hook",
						Env:      []string{"VENDOR_ENV5=value5"},
					},
					{
						HookName: "poststop",
						Path:     "/usr/local/bin/poststop-vendor-hook",
					},
				},
			},
			expectedResult: oci.Spec{
				Hooks: &oci.Hooks{
					Prestart: []oci.Hook{
						{
							Path: "/usr/local/bin/prestart-vendor-hook",
							Args: []string{"--verbose"},
							Env:  []string{"VENDOR_ENV1=value1"},
						},
					},
					CreateRuntime: []oci.Hook{
						{
							Path: "/usr/local/bin/cr-vendor-hook",
							Args: []string{"--debug"},
							Env:  []string{"VENDOR_ENV2=value2"},
						},
					},
					CreateContainer: []oci.Hook{
						{
							Path: "/usr/local/bin/cc-vendor-hook",
							Args: []string{"--create"},
							Env:  []string{"VENDOR_ENV3=value3"},
						},
					},
					StartContainer: []oci.Hook{
						{
							Path: "/usr/local/bin/sc-vendor-hook",
							Args: []string{"--start"},
							Env:  []string{"VENDOR_ENV4=value4"},
						},
					},
					Poststart: []oci.Hook{
						{
							Path: "/usr/local/bin/poststart-vendor-hook",
							Env:  []string{"VENDOR_ENV5=value5"},
						},
					},
					Poststop: []oci.Hook{
						{
							Path: "/usr/local/bin/poststop-vendor-hook",
						},
					},
				},
			},
		},
		{
			name:   "unknown hook",
			config: &oci.Spec{},
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "unknown",
						Path:     "/usr/local/bin/prestart-vendor-hook",
						Args:     []string{"--verbose"},
						Env:      []string{"VENDOR_ENV1=value1"},
					},
				},
			},
			expectedResult: oci.Spec{},
			expectedError:  true,
		},
		{
			name: "multiple edits",
			config: &oci.Spec{
				Version: "1.0.2",
				Process: &oci.Process{
					Env: []string{"ENV=value"},
				},
				Root: &oci.Root{
					Path:     "/chroot/root1",
					Readonly: true,
				},
				Hostname: "some.host.com",
				Mounts: []oci.Mount{
					{
						Source:      "/source",
						Destination: "/destination",
						Type:        "tmpfs",
						Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
					},
				},
				Hooks: &oci.Hooks{
					Prestart: []oci.Hook{
						{
							Path: "/bin/hook",
							Args: []string{"--prestart"},
							Env:  []string{"HOOKENV=hookval"},
						},
					},
				},
			},
			edits: &cdi.ContainerEdits{
				Env: []string{"BAR=BARVALUE1"},
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/device1",
					},
				},
				Hooks: []*cdi.Hook{
					{
						HookName: "prestart",
						Path:     "/bin/vendor-hook",
					},
					{
						HookName: "poststart",
						Path:     "/bin/poststart",
						Args:     []string{"--verbose"},
						Env:      []string{"VENDOR_ENV1=value1"},
					},
				},
				Mounts: []*cdi.Mount{
					{
						HostPath:      "/mnt/mount1",
						ContainerPath: "/mnt/mount1",
						Options:       []string{"noexec", "noatime"},
					},
				},
			},
			expectedResult: oci.Spec{
				Version: "1.0.2",
				Process: &oci.Process{
					Env: []string{"ENV=value", "BAR=BARVALUE1"},
				},
				Root: &oci.Root{
					Path:     "/chroot/root1",
					Readonly: true,
				},
				Hostname: "some.host.com",
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{Path: "/dev/device1"},
					},
				},
				Mounts: []oci.Mount{
					{
						Source:      "/source",
						Destination: "/destination",
						Type:        "tmpfs",
						Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
					},
					{
						Source:      "/mnt/mount1",
						Destination: "/mnt/mount1",
						Options:     []string{"noexec", "noatime"},
					},
				},
				Hooks: &oci.Hooks{
					Prestart: []oci.Hook{
						{
							Path: "/bin/hook",
							Args: []string{"--prestart"},
							Env:  []string{"HOOKENV=hookval"},
						},
						{
							Path: "/bin/vendor-hook",
						},
					},
					Poststart: []oci.Hook{
						{
							Path: "/bin/poststart",
							Args: []string{"--verbose"},
							Env:  []string{"VENDOR_ENV1=value1"},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			edits := ContainerEdits{tc.edits}
			err := edits.Apply(tc.config)
			if tc.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.edits != nil {
				require.Equal(t, tc.expectedResult, *tc.config)
			}
		})
	}
}

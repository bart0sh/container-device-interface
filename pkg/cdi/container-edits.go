package cdi

import (
	"strings"

	specs "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/generate"

	"github.com/pkg/errors"
)

const (
	// PrestartHook is the OCI prestart hook name.
	PrestartHook = "prestart"
	// CreateRuntimeHook is the OCI CreateRuntime hook name.
	CreateRuntimeHook = "createRuntime"
	// CreateContainerHook is the OCI CreateContainer hook name.
	CreateContainerHook = "createContainer"
	// StartContainerHook is the OCI StartContainer hook name.
	StartContainerHook = "startContainer"
	// PoststartHook is the OCI Poststart hook name.
	PoststartHook = "poststart"
	// PoststopHook is the OCI Poststop hook name.
	PoststopHook = "poststop"
)

// ContainerEdits represent updates to be applied to an OCI Spec.
type ContainerEdits struct {
	*specs.ContainerEdits
}

// Apply edits to the given OCI Spec.
func (e *ContainerEdits) Apply(spec *oci.Spec) error {
	if spec == nil {
		return errors.New("nil OCI Spec")
	}
	if e == nil || e.ContainerEdits == nil {
		return nil
	}

	g := generate.NewFromSpec(spec)

	if len(e.Env) > 0 {
		g.AddMultipleProcessEnv(e.Env)
	}
	for _, d := range e.DeviceNodes {
		g.AddDevice(d.ToOCI())
	}
	for _, m := range e.Mounts {
		g.AddMount(m.ToOCI())
	}
	for _, h := range e.Hooks {
		switch h.HookName {
		case PrestartHook:
			g.AddPreStartHook(h.ToOCI())
		case PoststartHook:
			g.AddPostStartHook(h.ToOCI())
		case PoststopHook:
			g.AddPostStopHook(h.ToOCI())
			// TODO: Runtime-tools should be updated to support these, too.
		case CreateRuntimeHook:
			ensureSpecHooks(spec)
			spec.Hooks.CreateRuntime = append(spec.Hooks.CreateRuntime, h.ToOCI())
		case CreateContainerHook:
			ensureSpecHooks(spec)
			spec.Hooks.CreateContainer = append(spec.Hooks.CreateContainer, h.ToOCI())
		case StartContainerHook:
			ensureSpecHooks(spec)
			spec.Hooks.StartContainer = append(spec.Hooks.StartContainer, h.ToOCI())
		default:
			return errors.Errorf("unknown hook name %q", h.HookName)
		}
	}

	return nil
}

// Validate basic sanity of the edits.
func (e *ContainerEdits) Validate() error {
	if e == nil || e.ContainerEdits == nil {
		return nil
	}

	if err := ValidateEnv(e.Env); err != nil {
		return errors.Wrap(err, "invalid container edits")
	}
	for _, d := range e.DeviceNodes {
		if err := (&DeviceNode{d}).Validate(); err != nil {
			return err
		}
	}
	for _, h := range e.Hooks {
		if err := (&Hook{h}).Validate(); err != nil {
			return err
		}
	}
	for _, m := range e.Mounts {
		if err := (&Mount{m}).Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateEnv validates the given environment variables.
func ValidateEnv(env []string) error {
	for _, v := range env {
		if strings.IndexByte(v, byte('=')) <= 0 {
			return errors.Errorf("invalid environment variable %q", v)
		}
	}

	return nil
}

// DeviceNode is CDI Spec DeviceNode wrapper, only used to validate devices.
type DeviceNode struct {
	*specs.DeviceNode
}

// Validate basic sanity of a DeviceNode.
func (d *DeviceNode) Validate() error {
	if d.Path == "" {
		return errors.New("invalid (empty) device path")
	}
	if d.Type != "" && d.Type != "b" && d.Type != "c" {
		return errors.Errorf("device %q: invalid type %q", d.Path, d.Type)
	}
	for _, c := range d.Permissions {
		if c != 'r' && c != 'w' && c != 'm' {
			return errors.Errorf("device %q: invalid permissions %q", d.Path, d.Permissions)
		}
	}
	return nil
}

// Hook is a CDI Spec Hook wrapper, only used to validate hooks.
type Hook struct {
	*specs.Hook
}

// Validate basic sanity of a Hook.
func (h *Hook) Validate() error {
	switch h.HookName {
	case PrestartHook:
	case CreateRuntimeHook:
	case CreateContainerHook:
	case StartContainerHook:
	case PoststartHook:
	case PoststopHook:
	default:
		return errors.Errorf("invalid hook name %q", h.HookName)
	}
	if h.Path == "" {
		return errors.Errorf("invalid hook %q, empty path", h.HookName)
	}
	if err := ValidateEnv(h.Env); err != nil {
		return errors.Wrapf(err, "invalid hook %q", h.HookName)
	}

	return nil
}

func ensureSpecHooks(spec *oci.Spec) {
	if spec.Hooks == nil {
		spec.Hooks = &oci.Hooks{}
	}
}

// Mount is a CDI Spec Mount wrapper, only used to validate mounts.
type Mount struct {
	*specs.Mount
}

// Validate basic sanity of the Mount.
func (m *Mount) Validate() error {
	if m.HostPath == "" {
		return errors.New("invalid mount, empty host path")
	}
	if m.ContainerPath == "" {
		return errors.New("invalid mount, empty container path")
	}

	return nil
}

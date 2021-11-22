package cdi

import (
	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"

	specs "github.com/container-orchestrated-devices/container-device-interface/specs-go"
)

// Device represents a CDI device of a Spec in a Cache.
type Device struct {
	// CDI device.
	*specs.Device
	// Source Spec.
	spec *Spec
}

// newDevice creates a new device instance.
func newDevice(spec *Spec, device specs.Device) (*Device, error) {
	dev := &Device{
		Device: &device,
		spec:   spec,
	}

	if err := dev.validate(); err != nil {
		return nil, err
	}

	return dev, nil
}

// QualifiedName returns the qualified name for the device.
func (d *Device) QualifiedName() string {
	return QualifiedDevice(d.spec.GetVendor(), d.spec.GetClass(), d.Name)
}

// GetSpec returns the Spec for the device.
func (d *Device) GetSpec() *Spec {
	return d.spec
}

// GetEdits returns the container edits for the device.
func (d *Device) GetEdits() *specs.ContainerEdits {
	return &d.ContainerEdits
}

// ApplyEdits applies the device's edits to an OCI Spec.
func (d *Device) ApplyEdits(spec *oci.Spec) error {
	e := ContainerEdits{&d.ContainerEdits}
	return e.Apply(spec)
}

// getPriority returns the priority for the device.
func (d *Device) getPriority() int {
	return d.spec.priority
}

// Validate the device with basic sanity checks.
func (d *Device) validate() error {
	if err := ValidateDeviceName(d.Name); err != nil {
		return err
	}
	edits := ContainerEdits{&d.ContainerEdits}
	if err := edits.Validate(); err != nil {
		return errors.Wrapf(err, "invalid device %q", d.Name)
	}

	return nil
}

package cdi

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	specs "github.com/container-orchestrated-devices/container-device-interface/specs-go"
)

var (
	validSpecVersions = map[string]struct{}{
		"0.1.0": {},
		"0.2.0": {},
	}
)

// Spec represents a CDI Spec, usually stored in a Cache. It is
// typically loaded from its associated path in the filesystem.
// The integer priority is usually the index of the Spec path's
// directory within the Spec directories assigned to the Cache.
// This priority is used to resolve conflicting device entries
// in the cache with identical fully qualified device names.
// The devices from the Spec with the highest priority (or IOW
// latest in the Spec directory list) wins.
type Spec struct {
	*specs.Spec
	path     string
	priority int
	vendor   string
	class    string
	devices  map[string]*Device
}

// LoadSpec loads a Spec from the given path and assigns it the
// the given priority.
func LoadSpec(path string, priority int) (*Spec, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to read CDI Spec")
	}

	spec, err := parseSpecData(data)
	if err != nil {
		return nil, err
	}

	return NewSpec(spec, path, priority)
}

// NewSpec creates a new Spec for the given CDI Spec data. The
// newly created Spec is marked as loaded from the given path
// and is assigned the given priority. NewSpec returns an orphan
// Spec which is not (yet) associated with any Cache.
func NewSpec(s *specs.Spec, path string, priority int) (*Spec, error) {
	spec := &Spec{
		Spec:     s,
		path:     filepath.Clean(path),
		priority: priority,
		devices:  make(map[string]*Device),
	}

	spec.vendor, spec.class = ParseQualifier(spec.Spec.Kind)

	if err := spec.validate(); err != nil {
		return nil, errors.Wrap(err, "invalid CDI Spec")
	}

	for _, d := range spec.Devices {
		device, err := newDevice(spec, d)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to add device %q", d.Name)
		}
		if _, ok := spec.devices[d.Name]; ok {
			return nil, errors.Errorf("multiple devices %q defined", d.Name)
		}
		spec.devices[d.Name] = device
	}

	return spec, nil
}

// parseSpecData parses the given data into a CDI Spec.
func parseSpecData(data []byte) (*specs.Spec, error) {
	spec := &specs.Spec{}
	err := yaml.UnmarshalStrict(data, spec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse CDI Spec")
	}

	return spec, nil
}

// GetPath returns the filesystem path for this Spec.
func (s *Spec) GetPath() string {
	return s.path
}

// GetVendor returns the vendor for this Spec.
func (s *Spec) GetVendor() string {
	return s.vendor
}

// GetClass returns the device class for this Spec.
func (s *Spec) GetClass() string {
	return s.class
}

// ListDeviceNames lists all devices of the Spec by unqualified name.
func (s *Spec) ListDeviceNames() ([]string, error) {
	var devices []string
	for _, d := range s.devices {
		devices = append(devices, d.Name)
	}
	sort.Strings(devices)

	return devices, nil
}

// ListQualifiedDeviceNames lists all devices of the Spec by qualified name.
func (s *Spec) ListQualifiedDeviceNames() ([]string, error) {
	var devices []string
	for _, d := range s.devices {
		devices = append(devices, d.QualifiedName())
	}
	sort.Strings(devices)

	return devices, nil
}

// ApplyEdits applies the global Spec edits to an OCI Spec.
func (s *Spec) ApplyEdits(spec *oci.Spec) error {
	e := ContainerEdits{&s.ContainerEdits}
	return e.Apply(spec)
}

// Validate the Spec with basic sanity checks.
func (s *Spec) validate() error {
	if _, ok := validSpecVersions[s.Version]; !ok {
		return errors.Errorf("invalid version %q", s.Version)
	}
	if err := s.validateSchema(); err != nil {
		return errors.Wrap(err, "JSON Schema validation failed")
	}
	if err := ValidateVendorName(s.vendor); err != nil {
		return errors.Wrap(err, "invalid vendor")
	}
	if err := ValidateClassName(s.class); err != nil {
		return err
	}
	edits := &ContainerEdits{&s.ContainerEdits}
	if err := edits.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate CDI Spec data against the CDI Spec JSON Schema.
func (s *Spec) validateSchema() error {
	// TODO
	return nil
}

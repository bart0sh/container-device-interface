package cdi

import (
	"os"
	"path/filepath"
	"sort"
	"sync"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/generate"
	"github.com/pkg/errors"

	"github.com/hashicorp/go-multierror"
)

const (
	// DefaultStaticDir is the default directory for static CDI Specs.
	DefaultStaticDir = "/etc/cdi"
	// DefaultDynamicDir is the default directory for dynamically generated CDI Specs.
	DefaultDynamicDir = "/var/run/cdi"
)

var (
	// DefaultSpecDirs contains the default CDI Spec directories.
	DefaultSpecDirs = []string{DefaultStaticDir, DefaultDynamicDir}
	// ErrStopScan is used as the return value from a SpecScanFunc to stop the scan.
	ErrStopScan = errors.New("stop spec scan")
)

// ScanSpecsFunc is the function type called by ScanSpecFiles for each Spec.
type ScanSpecsFunc func(string, int, *Spec, error) error

// CacheOption is an option for a Cache.
type CacheOption func(*Cache) error

// WithFilter specifies a device filtering glob pattern for a Cache.
func WithFilter(devicePattern string) CacheOption {
	return func(c *Cache) error {
		f, err := newFilter(devicePattern)
		if err != nil {
			return err
		}
		c.filter = f
		return nil
	}
}

// Cache stores CDI Specs loaded from Spec directories.
type Cache struct {
	sync.Mutex
	filter   *filter
	specDirs []string
	specs    map[string][]*Spec
	errors   []error
	devices  map[string]*Device
}

// NewCache creates a new CDI Cache. The cache is populated from a set
// of CDI Spec directories. These can be specified using a WithSpecDirs
// option. The default set of directories is DefaultSpecDirs. If more
// than one directory is used, Spec files in directories later in the
// list have precedence over Spec files in earlier directories.
func NewCache(opts ...CacheOption) (*Cache, error) {
	c := &Cache{
		specDirs: []string{DefaultStaticDir, DefaultDynamicDir},
	}

	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, errors.Wrap(err, "failed create cache, failed option")
		}
	}

	err := c.Refresh()

	return c, err
}

// GetSpecDirectories returns the spec directories used by this cache.
func (c *Cache) GetSpecDirectories() []string {
	if c == nil {
		return nil
	}
	dirs := make([]string, len(c.specDirs))
	copy(dirs, c.specDirs)
	return dirs
}

// ListVendors lists all vendors known to the cache.
func (c *Cache) ListVendors() []string {
	var vendors []string

	if c == nil {
		return nil
	}

	c.Lock()
	defer c.Unlock()

	for vendor := range c.specs {
		vendors = append(vendors, vendor)
	}

	sort.Strings(vendors)

	return vendors
}

// GetVendorSpecs returns all Specs for the given vendor.
func (c *Cache) GetVendorSpecs(vendor string) []*Spec {
	if c == nil {
		return nil
	}

	c.Lock()
	defer c.Unlock()

	return c.specs[vendor]
}

// HasDevice checks if the cache contains the given qualified device.
func (c *Cache) HasDevice(device string) bool {
	if c == nil {
		return false
	}

	c.Lock()
	defer c.Unlock()

	_, found := c.devices[device]
	return found
}

// ListDevices lists the qualified names of all devices in the cache.
func (c *Cache) ListDevices() []string {
	var devices []string

	if c == nil {
		return nil
	}

	c.Lock()
	defer c.Unlock()

	for device := range c.devices {
		devices = append(devices, device)
	}
	sort.Strings(devices)

	return devices
}

// MatchDevices lists the qualified names of devices matching a glob pattern.
func (c *Cache) MatchDevices(devicePattern string) ([]string, error) {
	var devices []string

	if c == nil {
		return nil, nil
	}

	f, err := newFilter(devicePattern)
	if err != nil {
		return nil, err
	}

	for device, d := range c.devices {
		if f == nil || f.match(d.GetSpec().GetVendor(), d.GetSpec().GetClass(), d.Name) {
			devices = append(devices, device)
		}
	}
	sort.Strings(devices)

	return devices, nil
}

// GetDevice returns the requested device from the cache.
func (c *Cache) GetDevice(device string) *Device {
	if c == nil {
		return nil
	}
	return c.devices[device]
}

// Errors returns any errors encountered while loading the cache.
func (c *Cache) Errors() []error {
	if c == nil {
		return []error{errors.New("nil CDI Spec cache")}
	}

	c.Lock()
	defer c.Unlock()

	var errors []error

	if errCnt := len(c.errors); errCnt > 0 {
		errors = make([]error, errCnt)
		copy(errors, c.errors)
	}

	return errors
}

// InjectDevices injects the given qualified devices into an OCI Spec.
// It returns any remaining unresolved devices and errors encountered
// while injecting the devices.
func (c *Cache) InjectDevices(ociSpec *oci.Spec, devices []string) ([]string, error) {
	var unresolved []string

	specs := map[*Spec]struct{}{}

	for _, device := range devices {
		d := c.GetDevice(device)
		if d == nil {
			unresolved = append(unresolved, device)
			continue
		}
		d.ApplyEdits(ociSpec)
		specs[d.GetSpec()] = struct{}{}
	}

	for spec := range specs {
		spec.ApplyEdits(ociSpec)
	}

	return unresolved, nil
}

// ResolveDevices resolves qualified devices present in an OCI Spec.
// These devices are store as OCI Linux Devices with the HostPath
// of the device set to the qualified name of the CDI device. During
// resolution these devices are replaced with their corresponding
// entries found in the cache. Any other related OCI Spec mutations
// (any injection of envronment variables, hooks, and mounts).
func (c *Cache) ResolveDevices(ociSpec *oci.Spec) ([]string, error) {
	var unresolved []string

	if ociSpec.Linux == nil || len(ociSpec.Linux.Devices) == 0 {
		return nil, nil
	}

	devices := ociSpec.Linux.Devices
	specGen := generate.NewFromSpec(ociSpec)
	specGen.ClearLinuxDevices()

	specs := map[*Spec]struct{}{}

	for _, ociDev := range devices {
		device := ociDev.Path
		if !IsQualifiedDevice(device) {
			specGen.AddDevice(ociDev)
			continue
		}
		d := c.GetDevice(device)
		if d == nil {
			unresolved = append(unresolved, device)
			specGen.AddDevice(ociDev)
			continue
		}
		d.ApplyEdits(ociSpec)
		specs[d.GetSpec()] = struct{}{}
	}

	for spec := range specs {
		spec.ApplyEdits(ociSpec)
	}

	return unresolved, nil
}

// Refresh rescans the CDI Spec directories and reloads all Specs.
func (c *Cache) Refresh() error {
	var (
		specs     = map[string][]*Spec{}
		devices   = map[string]*Device{}
		conflicts = map[string]struct{}{}
		result    []error
	)

	_ = c.ScanSpecFiles(func(path string, priority int, spec *Spec, err error) error {
		if err != nil {
			result = append(result, errors.Wrapf(err, "failed to load CDI Spec %q", path))
			return nil
		}

		vendor := spec.GetVendor()
		specs[vendor] = append(specs[vendor], spec)

		for _, dev := range spec.devices {
			qualified := dev.QualifiedName()
			if !c.filter.matchQualifiedDevice(qualified) {
				continue
			}
			other, ok := devices[qualified]
			if ok {
				switch {
				case other.getPriority() == dev.getPriority():
					result = append(result,
						errors.Errorf("device %q defined multiple times (%q, %q)",
							qualified, other.GetSpec().GetPath(), spec.GetPath()))
					conflicts[qualified] = struct{}{}
					continue
				case other.getPriority() > dev.getPriority():
					continue
				}
			}
			devices[qualified] = dev
		}

		return nil
	})

	for conflict := range conflicts {
		delete(devices, conflict)
	}

	c.Lock()
	defer c.Unlock()

	c.specs = specs
	c.devices = devices
	c.errors = result

	if c.errors != nil {
		return multierror.Append(nil, c.errors...)
	}

	return nil
}

// ScanSpecFiles scans the CDI Spec directories, reads each Spec file
// and calls the scan function for each. ScanSpecFile does not modify
// the cache itself. The specs loaded by this function are all orphans.
// They are *NOT* loaded to the cache.
func (c *Cache) ScanSpecFiles(specFn ScanSpecsFunc) error {
	var (
		spec *Spec
		err  error
	)

	for priority, dir := range c.specDirs {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if ext := filepath.Ext(path); ext == ".json" || ext == ".yaml" {
					if err = specFn(path, priority, nil, err); err != nil {
						return err
					}
				}
				return nil
			}

			if info.IsDir() {
				if path != dir {
					return filepath.SkipDir
				}
				return nil
			}

			if name := filepath.Base(info.Name()); name != "" &&
				(name[0] == '.' || name[0] == '#') {
				return nil
			}
			if ext := filepath.Ext(info.Name()); ext != ".json" && ext != ".yaml" {
				return nil
			}

			spec, err = LoadSpec(path, priority)
			if spec == nil && err == nil {
				return nil
			}
			err = specFn(path, priority, spec, err)

			return err
		})

		if err != nil && err != ErrStopScan {
			return err
		}
	}

	return err
}

// SetOptions applies the given options to the cache. If any options
// have changed the cache is refreshed.
func (c *Cache) SetOptions(options ...CacheOption) error {
	if c == nil {
		return errors.New("failed to set options on nil Cache")
	}

	c.Lock()
	dirs := make([]string, len(c.specDirs))
	copy(dirs, c.specDirs)

	for _, o := range options {
		if err := o(c); err != nil {
			return err
		}
	}

	refresh := false
	if len(dirs) != len(c.specDirs) {
		refresh = true
	} else {
		for idx, dir := range c.specDirs {
			if dirs[idx] != dir {
				refresh = true
				break
			}
		}
	}
	c.Unlock()

	if refresh {
		if err := c.Refresh(); err != nil {
			return errors.Wrapf(err, "option-triggered cache refresh failed")
		}
	}

	return nil
}

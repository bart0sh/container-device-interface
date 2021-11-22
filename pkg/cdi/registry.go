package cdi

import (
	"sync"

	oci "github.com/opencontainers/runtime-spec/specs-go"
)

//
// Registry represents the CDI registry. It is a cache of all known CDI
// Specs present in the system and is the primary interface clients use
// to interact with CDI.
//
// Devices are referenced in the registry by fully qualified names which
// are of the following form
//   <device-vendor>/<device-class>=<device-name>
// The registry provides functions for
//   - checking if a fully qualified name is present in the registry
//   - listing qualified names present in the registry
//   - injecting devices by CDI name into an OCI Spec
//   - resolving CDI names present in an OCI Spec to devices
//   - refreshing the registry (rescan and reload Spec files)
//   - querying errors encountered while loading Spec files
//   - setting registry options
//
// The registry is automatically populated upon startup by scanning the
// default CDI Spec directories. These directories can be overridden at
// will using `SetOptions` with the `WithSpecDirs()` registry option.
//
// The registry can be refreshed at any point in time. Refreshing rescans
// the Spec directories, reloads updated Spec files, loads new Spec files
// purges removed Spec files, and finally updates the index of devices to
// reflect the new state of the registry. The registry can be queried for
// any errors encountered during the last (or the initial) refresh cycle.
// Note that there is no hardcoded policy of how to handle errors in CDI.
// It is up to the application to decide how to best handle errors during
// a refresh.
//
// The registry provides two functions for injecting CDI devices into an
// OCI Spec. `InjectDevices` takes an OCI Spec and an explicit list of
// fully qualified CDI device names, resolves these names to actual CDI
// devices and injects these into the OCI Spec. `ResolveDevices` takes
// only an OCI Spec, and scans all `HostPath`'s of the devices present.
// If any device path is found to contain a fully qualified CDI device,
// the devices is replaced with the CDI devices it resolves to.
//

// Registry is a cache of all CDI Specs discovered in the system.
type Registry = Cache

// RegistryOption can be applied to the registry to alter its behavior.
type RegistryOption = CacheOption

var (
	registry *Registry
	initOnce sync.Once
)

// WithSpecDirs returns an option to set the CDI Specs directories.
func WithSpecDirs(dirs ...string) RegistryOption {
	return func(c *Cache) error {
		c.specDirs = make([]string, len(dirs))
		copy(c.specDirs, dirs)
		return nil
	}
}

// ListDevices lists all qualified devices in the registry.
func ListDevices() []string {
	return GetRegistry().ListDevices()
}

// MatchDevices lists all qualified devices matching the glob pattern.
func MatchDevices(pattern string) ([]string, error) {
	return GetRegistry().MatchDevices(pattern)
}

// HasDevice checks if the registry contains the qualified device.
func HasDevice(device string) bool {
	return GetRegistry().HasDevice(device)
}

// Errors returns errors encountered during the last registry refresh.
func Errors() []error {
	return GetRegistry().Errors()
}

// InjectDevices injects qualified devices to an OCI Spec.
// It returns any unresolved devices and an error if injection fails.
func InjectDevices(ociSpec *oci.Spec, devices []string) ([]string, error) {
	return GetRegistry().InjectDevices(ociSpec, devices)
}

// ResolveDevices resolves qualified devices present as a `HostPath` of
// any OCI Device in the OCI Spec. It returns the resolved CDI names of
// resolved devices and an error if resolving fails.
func ResolveDevices(ociSpec *oci.Spec) ([]string, error) {
	return GetRegistry().ResolveDevices(ociSpec)
}

// Refresh rescans all Spec directories and reloads CDI Specs to the registry.
// If any errors are encountered during refresh an error is returned. Note
// however, that upon errors succesfully loaded data is not discarded, neither
// does the registry become necessarily unable to resolve CDI devices. For
// instance, a parse error in a Spec file of vendor A has no effect on trying
// to resolve CDI devices for vendor B.
func Refresh() error {
	return GetRegistry().Refresh()
}

// SetOptions applies the given options to the registry.
func SetOptions(options ...RegistryOption) error {
	var (
		set bool
		err error
	)
	// We initOnce.Do() both here and in GetRegistry() to provide an
	// officially supported way to avoid scanning Spec directories
	// twice (once in GetRegistry()/NewCache() and a second time
	// here/in SetOptions()). Now if SetOptions() is called before
	// any other registry functions, the directories will be scanned
	// only once.
	initOnce.Do(func() {
		registry, err = NewCache(options...)
		set = true
	})
	if !set {
		err = GetRegistry().SetOptions(options...)
	}
	return err
}

// GetRegistry returns the CDI registry.
func GetRegistry() *Registry {
	initOnce.Do(func() { registry, _ = NewCache() })
	return registry
}

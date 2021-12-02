//
// This overview provides a brief high-level introduction to CDI and this
// package. See https://github.com/container-orchestrated-devices/container-device-interface
// for more information about CDI itself.
//
// Container Device Interface
//
// CDI, the Container Device Interface, provides advanced 3rd party device
// support for container runtimes. CDI uses CDI Specs, vendor-provided
// specification files, to describe how a container's runtime environment
// needs to be altered in order for software running inside the container
// to gain access and be able to use vendor-specific devices. In addition
// to the low-level platform-specific details necessary to interact with
// a device, this metadata can contain further OCI metadata, for instance
// extra environment variables to set, extra directories to mount, and
// OCI hooks to run.
//
// In the CDI device model containers request access to a device using a
// so-called fully qualified device name, or qualified name for short. This
// name consists of a vendor name, a device class, and a device identifier.
// These pieces of information together, and consequently the qualified
// likewise uniquely identifies a device among all vendors, device types
// and device instances.
//
// The CDI package provides an API for consuming CDI devices. Among others
// things this API implements discovery, loading and caching of CDI Specs
// and devices, resolution of fully qualified CDI device names to OCI
// devices and other metadata, and injection of CDI devices/OCI metadata
// to a container's OCI Runtime Spec.
//
// Additionally, there is a bunch of extra functionality implemented for
// more special albeit marginal use cases. For the vast majority of CDI
// consumers these are neither of any interest nor any use. See the full
// documentation for the Cache, Spec and Device types for more details
// about these.
//
// Registry
//
// The primary interface to interact with CDI devices is the CDI Registry.
// It is essentially a default cache of all Specs and devices discovered
// in standard CDI directories on the system. The Registry is implicitly
// instantiated and always available once it has been accessed the first
// time. Using the registry an OCI runtime client or an OCI runtime can
// easily inject CDI devices into a container, as represented by an OCI
// Runtime Spec, during the container's creation.
//
//  import (
//      "fmt"
//      "strings"
//
//      "github.com/pkg/errors"
//      log "github.com/sirupsen/logrus"
//
//      "github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
//      oci "github.com/opencontainers/runtime-spec/specs-go"
//  )
//
//  func injectCDIDevices(spec *oci.Spec, devices []string) error {
//      log.Debug("pristine OCI Spec: %s", dumpSpec(spec))
//
//      unresolved, err := cdi.InjectDevices(spec, devices)
//      if err != nil {
//          return errors.Wrap(err, "CDI device injection failed")
//      }
//      if unresolved != nil {
//          return errors.Errorf("CDI resolution failed for %s",
//              strings.Join(unresolved, ", "))
//      }
//
//      log.Debug("CDI-updated OCI Spec: %s", dumpSpec(spec))
//      return nil
//  }
//
// Refresh
//
// While there are cases where the set of CDI devices is static, there are
// systems where the set of devices present (or usable) at any time changes
// dynamically. In these cases it is necessary to update the set of CDI
// Specs dynamically as well, to reflect the changes in the sate of the
// devices. The Registry offers an API for CDI consumers to trigger the
// rediscovery and reloading of CDI Specs and devices.
//
//  import (
//      "fmt"
//      "strings"
//
//      "github.com/pkg/errors"
//      log "github.com/sirupsen/logrus"
//
//      "github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
//      oci "github.com/opencontainers/runtime-spec/specs-go"
//  )
//
//  func injectCDIDevices(spec *oci.Spec, devices []string) error {
//      if err := cdi.Refresh(); err != nil {
//          // See the chapter below on error handling alternatives...
//          return errors.Wrap(err, "registry refresh failed")
//      }
//
//      log.Debug("pristine OCI Spec: %s", dumpSpec(spec))
//
//      unresolved, err := cdi.InjectDevices(spec, devices)
//      if err != nil {
//          return errors.Wrap(err, "CDI device injection failed")
//      }
//      if unresolved != nil {
//          return errors.Errorf("CDI resolution failed for %s",
//              strings.Join(unresolved, ", "))
//      }
//
//      log.Debug("CDI-updated OCI Spec: %s", dumpSpec(spec))
//      return nil
//  }
//
// Error Handling
//
// The CDI package itself does not try to dictate whether CDI device
// resolution and injection should happen in the runtime clients or in
// the runtime implementation itself. The package has been designed with
// both possibilities in mind. Therefore there is no built-in hardwired
// policy for handling Registry refresh errors either. Instead the package
// attempts to provide enough of mechanims for CDI consumers to detect
// errors and decide what is the best course of (re)action to them.
//
// In particular this means that the Registry neither discards all new
// data nor becomes necessarily unable to resolve CDI devices when errors
// occur. For instance, a corrupted dynamically generated Spec file for
// vendor A will have no effect on trying to resolve and inject devices
// of vendor B into an OCI Runtime Spec.
//
// While the above sample code snippet chose to abort device injection
// and bail out with an error on a failed refresh, it is perfectly fine
// for a runtime implementation to go ahead instead and try to perform
// device injection regardless of the error, and only bail out if this
// fails.
//
// Spec and Device Priority/Precendence
//
// It is normal practice to keep dynamically generated CDI Spec files
// separate from any static pre-installed ones. The default configuration
// for Spec directories suggests using /etc for static files and /var/run
// for dynamically generated ones.
//
// It is possible to arrange things so that any dynamically generated
// CDI device entry takes precedence over any static entry for the same
// device. The default configuration is set up so that this is exactly
// what happens if static CDI Specs are installed in /etc/cdi while
// dynamic ones are generated in /var/run/cdi.
//
// The mechanism for determining precedence does not attach or imply
// any semantics to any directories per se. Instead when a CDI Spec file
// is loaded it is assigned a priority, which is simply the integer index
// of the Spec files directory, in the list of all directories. When a
// fully qualified device has more than one Spec entry, the entry chosen
// is the one with the highest priority. IOW, it is taken from the latest
// directory in the list.
//
package cdi

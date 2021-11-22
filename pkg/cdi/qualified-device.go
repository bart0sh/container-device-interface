package cdi

import (
	"strings"

	"github.com/pkg/errors"
)

// QualifiedDevice returns a qualified name for the device.
// The syntax of a qualified device name is
//     "<vendor>/<class>=<name>".
func QualifiedDevice(vendor, class, name string) string {
	return vendor + "/" + class + "=" + name
}

// IsQualifiedDevice checks if a device name is qualified.
func IsQualifiedDevice(device string) bool {
	_, _, _, err := ParseQualifiedDevice(device)
	return err == nil
}

// ParseQualifiedDevice parses a qualified device name into vendor, class
// and name components. The validity of the syntax of each component is
// then verified. If parsing or verification fails, vendor and class are
// returned empty, together with the verbatim input as name, and an error
// describing the reason for failure.
func ParseQualifiedDevice(device string) (string, string, string, error) {
	vendor, class, name := ParseDevice(device)

	if vendor == "" {
		return "", "", device, errors.Errorf("unqualified device %q, missing vendor", device)
	}
	if class == "" {
		return "", "", device, errors.Errorf("unqualified device %q, missing class", device)
	}
	if name == "" {
		return "", "", device, errors.Errorf("unqualified device %q, missing device name", device)
	}

	if err := ValidateVendorName(vendor); err != nil {
		return "", "", device, errors.Wrapf(err, "invalid device %q", device)
	}
	if err := ValidateClassName(class); err != nil {
		return "", "", device, errors.Wrapf(err, "invalid device %q", device)
	}
	if err := ValidateDeviceName(name); err != nil {
		return "", "", device, errors.Wrapf(err, "invalid device %q", device)
	}

	return vendor, class, name, nil
}

// ParseDevice tries to split a device name or pattern into vendor, class, and
// name. If parsing fails, for instance for unqualified device names, an empty
// vendor and class is returned together with name set to the verbatim input.
func ParseDevice(device string) (string, string, string) {
	if device == "" || device[0] == '/' {
		return "", "", device
	}

	parts := strings.SplitN(device, "=", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", device
	}

	name := parts[1]
	vendor, class := ParseQualifier(parts[0])
	if vendor == "" {
		return "", "", device
	}

	return vendor, class, name
}

// ParseQualifier splits a device qualifier into vendor and class.
// The syntax for a device qualifier is
//     "<vendor>/<class>"
// If the input fails to parse as a qualifier, an empty vendor and
// the verbatim input as the class is returned.
func ParseQualifier(kind string) (string, string) {
	parts := strings.SplitN(kind, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", kind
	}
	return parts[0], parts[1]
}

// ValidateVendorName checks the validity of a vendor name.
// A vendor name may contain the following ASCII characters:
//   - upper- and lowercase letters ('A'-'Z', 'a'-'z')
//   - digits ('0'-'9')
//   - underscore, dash, and dot ('_', '-', and '.')
func ValidateVendorName(vendor string) error {
	if vendor == "" {
		return errors.Errorf("invalid (empty) vendor name")
	}
	for _, c := range vendor[:] {
		switch {
		case '0' <= c && c <= '9':
		case 'a' <= c && c <= 'z':
		case 'A' <= c && c <= 'Z':
		case c == '_' || c == '-' || c == '.':
		default:
			return errors.Errorf("invalid character '%c' in vendor name %q",
				c, vendor)
		}
	}
	return nil
}

// ValidateClassName checks the validity of class name.
// A class name may contain the following ASCII characters:
//   - upper- and lowercase letters ('A'-'Z', 'a'-'z')
//   - digits ('0'-'9')
//   - underscore and dash ('_', '-')
func ValidateClassName(class string) error {
	if class == "" {
		return errors.Errorf("invalid (empty) device class")
	}
	for _, c := range class[:] {
		switch {
		case '0' <= c && c <= '9':
		case 'a' <= c && c <= 'z':
		case 'A' <= c && c <= 'Z':
		case c == '_' || c == '-':
		default:
			return errors.Errorf("invalid character '%c' in device class %q",
				c, class)
		}
	}
	return nil
}

// ValidateDeviceName checks the validity of a device name.
// A device name may contain the following ASCII characters:
//   - upper- and lowercase letters ('A'-'Z', 'a'-'z')
//   - digits ('0'-'9')
//   - underscore, dash, dot, colon ('_', '-', '.', ':')
func ValidateDeviceName(name string) error {
	if name == "" {
		return errors.Errorf("invalid (empty) device name")
	}
	for _, c := range name[:] {
		switch {
		case '0' <= c && c <= '9':
		case 'a' <= c && c <= 'z':
		case 'A' <= c && c <= 'Z':
		case c == '_' || c == '-' || c == '.' || c == ':':
		default:
			return errors.Errorf("invalid character '%c' in device name %q",
				c, name)
		}
	}
	return nil
}

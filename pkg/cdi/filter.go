package cdi

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type filter struct {
	comp struct {
		vendor string
		class  string
		name   string
	}
	glob struct {
		vendor string
		class  string
		name   string
	}
}

func newFilter(device string) (*filter, error) {
	const globbing = "?*[-"
	var err error

	if device == "" {
		return nil, nil
	}

	vendor, class, name := ParseDevice(device)
	if vendor == "" {
		return nil, errors.Errorf("invalid filter %q (not qualified)", device)
	}

	f := &filter{}
	if strings.ContainsAny(vendor, globbing) {
		f.glob.vendor, err = testPattern("vendor", vendor)
		if err != nil {
			return nil, err
		}
	} else {
		f.comp.vendor = vendor
	}
	if strings.ContainsAny(class, globbing) {
		f.glob.class, err = testPattern("class", class)
	} else {
		f.comp.class = class
	}
	if strings.ContainsAny(name, globbing) {
		f.glob.name, err = testPattern("name", name)
	} else {
		f.comp.name = name
	}
	return f, nil
}

func (f *filter) compare(o *filter) bool {
	if f == nil {
		return o == nil
	}
	if o == nil {
		return f == nil
	}
	return f.comp.vendor == o.comp.vendor && f.glob.vendor == o.glob.vendor &&
		f.comp.class == o.comp.class && f.glob.class == o.glob.class &&
		f.comp.name == o.comp.name && f.glob.name == o.glob.name
}

func (f *filter) matchVendor(vendor string) bool {
	if f == nil {
		return true
	}
	if f.glob.vendor == "" {
		if f.comp.vendor == "" {
			return true
		}
		return f.comp.vendor == vendor
	}
	match, _ := filepath.Match(f.glob.vendor, vendor)
	return match
}

func (f *filter) matchClass(class string) bool {
	if f == nil {
		return true
	}
	if f.glob.class == "" {
		if f.comp.class == "" {
			return true
		}
		return f.comp.class == class
	}
	match, _ := filepath.Match(f.glob.class, class)
	return match
}

func (f *filter) matchName(name string) bool {
	if f == nil {
		return true
	}
	if f.glob.name == "" {
		if f.comp.name == "" {
			return true
		}
		return f.comp.name == name
	}
	match, _ := filepath.Match(f.glob.name, name)
	return match
}

func (f *filter) matchQualifiedDevice(device string) bool {
	if f == nil {
		return true
	}
	vendor, class, name, err := ParseQualifiedDevice(device)
	if err != nil {
		return false
	}
	return f.matchVendor(vendor) && f.matchClass(class) && f.matchName(name)
}

func (f *filter) match(vendor, class, name string) bool {
	if f == nil {
		return true
	}
	return (vendor == "" || f.matchVendor(vendor)) &&
		(class == "" || f.matchClass(class)) &&
		(name == "" || f.matchName(name))
}

func testPattern(which, pattern string) (string, error) {
	if pattern == "" || pattern == "*" {
		return "", nil
	}
	if _, err := filepath.Match(pattern, "test"); err != nil {
		return "", errors.Wrapf(err, "invalid %s glob pattern %q",
			which, pattern)
	}
	return pattern, nil
}

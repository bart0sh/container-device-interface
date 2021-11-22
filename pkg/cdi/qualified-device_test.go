package cdi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	unqualifiedOrInvalidNames = []string{
		"",
		"/dev/null",
		"whatever",
		"foo=bar",
		"foo=bar/xyzzy",
		"/=",
		"vendor.com/class=",
		"vendor.com/=dev",
		"foo=",
	}
	validQualifiedNamesOrPatterns = []struct {
		device string
		vendor string
		class  string
		name   string
	}{
		{"vendor.com/device=dev1", "vendor.com", "device", "dev1"},
		{"*/device=dev1", "*", "device", "dev1"},
		{"other-vendor.com/device=dev1", "other-vendor.com", "device", "dev1"},
		{"other-vendor.com/*=dev1", "other-vendor.com", "*", "dev1"},
		{"sub.vendor.com/device=dev1", "sub.vendor.com", "device", "dev1"},
		{"sub.vendor.com/device=*", "sub.vendor.com", "device", "*"},
		{"vendor.com/device=dev.2", "vendor.com", "device", "dev.2"},
		{"*/*=dev.2", "*", "*", "dev.2"},
		{"vendor.com/device=dev.2", "vendor.com", "device", "dev.2"},
		{"*/device=*", "*", "device", "*"},
		{"vendor.com/device=dev:3", "vendor.com", "device", "dev:3"},
		{"vendor.com/*=*", "vendor.com", "*", "*"},
		{"vendor.com/device=dev4", "vendor.com", "device", "dev4"},
		{"*/*=*", "*", "*", "*"},
	}
)

func TestIsQualifiedDevice(t *testing.T) {
	for _, device := range unqualifiedOrInvalidNames {
		isQualified := IsQualifiedDevice(device)
		require.False(t, isQualified, "unqualified device %q", device)
	}
	for _, tc := range validQualifiedNamesOrPatterns {
		isQualified := IsQualifiedDevice(tc.device)
		if isQualified {
			v, c, n, err := ParseQualifiedDevice(tc.device)
			require.Nil(t, err)
			require.True(t, tc.vendor == v, "qualified device %q", tc.device)
			require.True(t, tc.class == c, "qualified device %q", tc.device)
			require.True(t, tc.name == n, "qualified device %q", tc.device)
			require.Equal(t, tc.device, QualifiedDevice(v, c, n),
				"qualified device %q", tc.device)
		} else {
			v, c, n := ParseDevice(tc.device)
			require.True(t, v != "" && c != "" && n != "",
				"qualified device or pattern %q", tc.device)
			require.True(t, tc.vendor == v, "qualified device %q", tc.device)
			require.True(t, tc.class == c, "qualified device %q", tc.device)
			require.True(t, tc.name == n, "qualified device %q", tc.device)
			require.Equal(t, tc.device, QualifiedDevice(v, c, n),
				"qualified device %q", tc.device)
		}
	}
}

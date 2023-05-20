package main

import (
	"testing"
)

func TestVersionFormatter(t *testing.T) {
	cases := []struct {
		name    string
		version string
		revision string
		expected string
	}{
		{
			name: "normal case",
			version: "1.0.0",
			revision: "0123456789abcdef",
			expected: "1.0.0 (rev: 0123456789abcdef)",
		},
		{
			name: "revision not specified",
			version: "1.0.0",
			revision: "",
			expected: "1.0.0",
		},
		{
			name: "version not specified",
			version: "",
			revision: "abcdef",
			expected: "devel (rev: abcdef)",
		},
		{
			name: "version and revision not specified",
			version: "",
			revision: "",
			expected: "devel",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := versionFormatter(tt.version, tt.revision)

			if actual != tt.expected {
				t.Errorf("expect %s, got %s", tt.expected, actual)
			}
		})
	}
}

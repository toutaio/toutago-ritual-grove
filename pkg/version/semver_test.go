package version

import (
	"fmt"
	"testing"
)

// TestParseVersion tests parsing semantic versions
func TestParseVersion(t *testing.T) {
	tests := []struct {
		input          string
		wantMajor      int
		wantMinor      int
		wantPatch      int
		wantPreRelease string
		wantError      bool
	}{
		{"1.0.0", 1, 0, 0, "", false},
		{"2.1.3", 2, 1, 3, "", false},
		{"0.1.0", 0, 1, 0, "", false},
		{"1.2.3-alpha", 1, 2, 3, "alpha", false},
		{"1.2.3-beta.1", 1, 2, 3, "beta.1", false},
		{"1.2.3-rc.1", 1, 2, 3, "rc.1", false},
		{"v1.2.3", 1, 2, 3, "", false},
		{"invalid", 0, 0, 0, "", true},
		{"1.2", 0, 0, 0, "", true},
		{"1.2.x", 0, 0, 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ver, err := Parse(tt.input)
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if ver.Major != tt.wantMajor {
				t.Errorf("Major: got %d, want %d", ver.Major, tt.wantMajor)
			}

			if ver.Minor != tt.wantMinor {
				t.Errorf("Minor: got %d, want %d", ver.Minor, tt.wantMinor)
			}

			if ver.Patch != tt.wantPatch {
				t.Errorf("Patch: got %d, want %d", ver.Patch, tt.wantPatch)
			}

			if ver.PreRelease != tt.wantPreRelease {
				t.Errorf("PreRelease: got %s, want %s", ver.PreRelease, tt.wantPreRelease)
			}
		})
	}
}

// TestCompareVersions tests version comparison
func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int // -1: v1 < v2, 0: v1 == v2, 1: v1 > v2
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.2.3", "1.2.3", 0},
		{"1.2.3", "1.2.4", -1},
		{"1.2.4", "1.2.3", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-alpha", 1},
		{"1.0.0-rc.1", "1.0.0-rc.2", -1},
		{"v1.0.0", "v1.0.1", -1},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.v1, tt.v2), func(t *testing.T) {
			result := Compare(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("Compare(%s, %s) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

// TestIsNewer tests checking if version is newer
func TestIsNewer(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected bool
	}{
		{"1.0.1", "1.0.0", true},
		{"1.0.0", "1.0.1", false},
		{"2.0.0", "1.9.9", true},
		{"1.0.0", "1.0.0", false},
		{"1.1.0", "1.0.5", true},
		{"1.0.0", "1.0.0-alpha", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s > %s", tt.v1, tt.v2), func(t *testing.T) {
			result := IsNewer(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("IsNewer(%s, %s) = %v, want %v", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

// TestIsCompatible tests compatibility checking
func TestIsCompatible(t *testing.T) {
	tests := []struct {
		current  string
		required string
		expected bool
	}{
		{"1.2.3", "1.0.0", true},
		{"1.2.3", "1.2.0", true},
		{"1.2.3", "1.2.3", true},
		{"1.2.3", "1.2.4", false},
		{"1.2.3", "1.3.0", false},
		{"1.2.3", "2.0.0", false},
		{"2.0.0", "1.0.0", false},
		{"1.0.0", "0.9.0", false},
		{"0.5.0", "0.4.0", true},
		{"0.5.0", "0.6.0", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s compatible with %s", tt.current, tt.required), func(t *testing.T) {
			result := IsCompatible(tt.current, tt.required)
			if result != tt.expected {
				t.Errorf("IsCompatible(%s, %s) = %v, want %v", tt.current, tt.required, result, tt.expected)
			}
		})
	}
}

// TestNextVersion tests generating next versions
func TestNextVersion(t *testing.T) {
	tests := []struct {
		current  string
		bump     string
		expected string
	}{
		{"1.2.3", "patch", "1.2.4"},
		{"1.2.3", "minor", "1.3.0"},
		{"1.2.3", "major", "2.0.0"},
		{"0.1.0", "patch", "0.1.1"},
		{"0.1.0", "minor", "0.2.0"},
		{"0.1.0", "major", "1.0.0"},
		{"1.0.0-alpha", "patch", "1.0.0"},
		{"1.0.0-alpha", "minor", "1.1.0"},
		{"1.0.0-alpha", "major", "2.0.0"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s bump %s", tt.current, tt.bump), func(t *testing.T) {
			result, err := NextVersion(tt.current, tt.bump)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("NextVersion(%s, %s) = %s, want %s", tt.current, tt.bump, result, tt.expected)
			}
		})
	}
}

// TestIsBreakingChange tests detecting breaking changes
func TestIsBreakingChange(t *testing.T) {
	tests := []struct {
		from     string
		to       string
		expected bool
	}{
		{"1.0.0", "1.0.1", false},
		{"1.0.0", "1.1.0", false},
		{"1.0.0", "2.0.0", true},
		{"1.9.9", "2.0.0", true},
		{"0.1.0", "0.2.0", true},
		{"0.1.0", "0.1.1", false},
		{"0.9.0", "1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s to %s", tt.from, tt.to), func(t *testing.T) {
			result := IsBreakingChange(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsBreakingChange(%s, %s) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

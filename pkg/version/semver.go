// Package version provides semantic versioning utilities
package version

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
}

// Parse parses a semantic version string
func Parse(s string) (*Version, error) {
	// Strip optional 'v' prefix
	s = strings.TrimPrefix(s, "v")

	// Split version and pre-release
	parts := strings.Split(s, "-")
	versionPart := parts[0]
	preRelease := ""
	if len(parts) > 1 {
		preRelease = parts[1]
	}

	// Parse version numbers
	nums := strings.Split(versionPart, ".")
	if len(nums) != 3 {
		return nil, fmt.Errorf("invalid version format: %s (expected X.Y.Z)", s)
	}

	major, err := strconv.Atoi(nums[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", nums[0])
	}

	minor, err := strconv.Atoi(nums[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", nums[1])
	}

	patch, err := strconv.Atoi(nums[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", nums[2])
	}

	return &Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: preRelease,
	}, nil
}

// String returns the string representation of the version
func (v *Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		s += "-" + v.PreRelease
	}
	return s
}

// Compare compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func Compare(v1, v2 string) int {
	ver1, err1 := Parse(v1)
	ver2, err2 := Parse(v2)

	if err1 != nil || err2 != nil {
		return 0
	}

	// Compare major
	if ver1.Major != ver2.Major {
		if ver1.Major > ver2.Major {
			return 1
		}
		return -1
	}

	// Compare minor
	if ver1.Minor != ver2.Minor {
		if ver1.Minor > ver2.Minor {
			return 1
		}
		return -1
	}

	// Compare patch
	if ver1.Patch != ver2.Patch {
		if ver1.Patch > ver2.Patch {
			return 1
		}
		return -1
	}

	// Compare pre-release
	// Releases without pre-release are newer than pre-releases
	if ver1.PreRelease == "" && ver2.PreRelease != "" {
		return 1
	}
	if ver1.PreRelease != "" && ver2.PreRelease == "" {
		return -1
	}

	// Compare pre-release strings lexicographically
	if ver1.PreRelease != ver2.PreRelease {
		if ver1.PreRelease > ver2.PreRelease {
			return 1
		}
		return -1
	}

	return 0
}

// IsNewer returns true if v1 is newer than v2
func IsNewer(v1, v2 string) bool {
	return Compare(v1, v2) > 0
}

// IsCompatible checks if current version satisfies required version
// Uses semantic versioning compatibility rules:
// - For 0.x versions: minor version must match or be higher
// - For 1.x+ versions: major must match, minor.patch must be >= required
func IsCompatible(current, required string) bool {
	cur, err1 := Parse(current)
	req, err2 := Parse(required)

	if err1 != nil || err2 != nil {
		return false
	}

	// For 0.x versions, only compatible within same minor version
	if req.Major == 0 {
		return cur.Major == 0 && cur.Minor >= req.Minor
	}

	// For 1.x+, major must match
	if cur.Major != req.Major {
		return false
	}

	// Minor must be >= required
	if cur.Minor > req.Minor {
		return true
	}

	// If minor matches, patch must be >= required
	if cur.Minor == req.Minor {
		return cur.Patch >= req.Patch
	}

	return false
}

// NextVersion generates the next version based on bump type
// bump can be: "major", "minor", or "patch"
// For pre-release versions, patch bump removes pre-release suffix
func NextVersion(current, bump string) (string, error) {
	ver, err := Parse(current)
	if err != nil {
		return "", err
	}

	// If it's a pre-release and bumping patch, just remove pre-release
	if ver.PreRelease != "" && bump == "patch" {
		return fmt.Sprintf("%d.%d.%d", ver.Major, ver.Minor, ver.Patch), nil
	}

	// Strip pre-release for all other bumps
	ver.PreRelease = ""

	switch bump {
	case "patch":
		ver.Patch++
	case "minor":
		ver.Minor++
		ver.Patch = 0
	case "major":
		ver.Major++
		ver.Minor = 0
		ver.Patch = 0
	default:
		return "", fmt.Errorf("invalid bump type: %s (must be major, minor, or patch)", bump)
	}

	return fmt.Sprintf("%d.%d.%d", ver.Major, ver.Minor, ver.Patch), nil
}

// IsBreakingChange detects if upgrading from 'from' to 'to' is a breaking change
// Breaking changes are:
// - Major version increase (1.x -> 2.x)
// - Minor version increase for 0.x versions (0.1.x -> 0.2.x)
func IsBreakingChange(from, to string) bool {
	fromVer, err1 := Parse(from)
	toVer, err2 := Parse(to)

	if err1 != nil || err2 != nil {
		return false
	}

	// Major version change is always breaking
	if toVer.Major > fromVer.Major {
		return true
	}

	// In 0.x, minor version change is breaking
	if fromVer.Major == 0 && toVer.Minor > fromVer.Minor {
		return true
	}

	return false
}

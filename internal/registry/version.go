package registry

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semantic version number.
type Version struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion parses a semantic version string into a Version struct.
// Accepts formats like "1.2.3" or "v1.2.3".
func ParseVersion(s string) (Version, error) {
	// Remove 'v' prefix if present
	s = strings.TrimPrefix(s, "v")

	if s == "" {
		return Version{}, fmt.Errorf("empty version string")
	}

	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version format: expected x.y.z, got %s", s)
	}

	var v Version
	var err error

	v.Major, err = strconv.Atoi(parts[0])
	if err != nil || v.Major < 0 {
		return Version{}, fmt.Errorf("invalid major version: %s", parts[0])
	}

	v.Minor, err = strconv.Atoi(parts[1])
	if err != nil || v.Minor < 0 {
		return Version{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	v.Patch, err = strconv.Atoi(parts[2])
	if err != nil || v.Patch < 0 {
		return Version{}, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return v, nil
}

// String returns the string representation of the version.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares two versions.
// Returns -1 if v < other, 0 if v == other, 1 if v > other.
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major > other.Major {
			return 1
		}
		return -1
	}

	if v.Minor != other.Minor {
		if v.Minor > other.Minor {
			return 1
		}
		return -1
	}

	if v.Patch != other.Patch {
		if v.Patch > other.Patch {
			return 1
		}
		return -1
	}

	return 0
}

// Constraint represents a version constraint.
type Constraint struct {
	Operator string
	Version  Version
}

var constraintRegex = regexp.MustCompile(`^(=|>=|>|<=|<)(.+)$`)

// ParseConstraint parses a version constraint string.
// Supports operators: =, >=, >, <=, <
func ParseConstraint(s string) (*Constraint, error) {
	matches := constraintRegex.FindStringSubmatch(s)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid constraint format: %s", s)
	}

	operator := matches[1]
	versionStr := matches[2]

	// Validate operator
	validOps := map[string]bool{"=": true, ">=": true, ">": true, "<=": true, "<": true}
	if !validOps[operator] {
		return nil, fmt.Errorf("invalid constraint operator: %s", operator)
	}

	version, err := ParseVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version in constraint: %w", err)
	}

	return &Constraint{
		Operator: operator,
		Version:  version,
	}, nil
}

// Satisfies checks if a version satisfies the constraint.
func (c *Constraint) Satisfies(v Version) bool {
	cmp := v.Compare(c.Version)

	switch c.Operator {
	case "=":
		return cmp == 0
	case ">":
		return cmp > 0
	case ">=":
		return cmp >= 0
	case "<":
		return cmp < 0
	case "<=":
		return cmp <= 0
	default:
		return false
	}
}

// SelectBestVersion selects the highest version that satisfies the constraint.
func SelectBestVersion(versions []string, constraint *Constraint) (string, error) {
	var best *Version
	var bestStr string

	for _, vStr := range versions {
		v, err := ParseVersion(vStr)
		if err != nil {
			continue // Skip invalid versions
		}

		if !constraint.Satisfies(v) {
			continue // Skip versions that don't satisfy constraint
		}

		if best == nil || v.Compare(*best) > 0 {
			best = &v
			bestStr = vStr
		}
	}

	if best == nil {
		return "", fmt.Errorf("no version satisfies constraint %s%s", constraint.Operator, constraint.Version)
	}

	return bestStr, nil
}

// IsVersionNewer checks if newVersion is newer than currentVersion.
// Returns true if newVersion > currentVersion.
func IsVersionNewer(newVersion, currentVersion string) (bool, error) {
	new, err := ParseVersion(newVersion)
	if err != nil {
		return false, fmt.Errorf("invalid new version: %w", err)
	}

	current, err := ParseVersion(currentVersion)
	if err != nil {
		return false, fmt.Errorf("invalid current version: %w", err)
	}

	return new.Compare(current) > 0, nil
}

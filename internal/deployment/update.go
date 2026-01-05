package deployment

import (
	"sort"

	"github.com/Masterminds/semver/v3"
)

// UpdateDetector handles version comparison and update detection
type UpdateDetector struct{}

// NewUpdateDetector creates a new update detector
func NewUpdateDetector() *UpdateDetector {
	return &UpdateDetector{}
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	FromVersion string
	ToVersion   string
	UpdateType  string // "patch", "minor", "major"
	IsBreaking  bool
	Description string
}

// IsUpdateAvailable checks if the available version is newer than current
func (d *UpdateDetector) IsUpdateAvailable(current, available *semver.Version) bool {
	return available.GreaterThan(current)
}

// IsBreakingChange checks if the update is a breaking change (major version bump)
func (d *UpdateDetector) IsBreakingChange(current, available *semver.Version) bool {
	return available.Major() > current.Major()
}

// ListUpdates returns all available updates newer than current version
func (d *UpdateDetector) ListUpdates(current *semver.Version, available []*semver.Version) []*semver.Version {
	var updates []*semver.Version

	for _, v := range available {
		if v.GreaterThan(current) {
			updates = append(updates, v)
		}
	}

	// Sort descending (newest first)
	sort.Slice(updates, func(i, j int) bool {
		return updates[i].GreaterThan(updates[j])
	})

	return updates
}

// GetUpdateInfo returns detailed information about an update
func (d *UpdateDetector) GetUpdateInfo(current, target *semver.Version) UpdateInfo {
	info := UpdateInfo{
		FromVersion: current.String(),
		ToVersion:   target.String(),
		IsBreaking:  d.IsBreakingChange(current, target),
	}

	// Determine update type
	if target.Major() > current.Major() {
		info.UpdateType = "major"
	} else if target.Minor() > current.Minor() {
		info.UpdateType = "minor"
	} else {
		info.UpdateType = "patch"
	}

	return info
}

// GetLatestCompatible returns the latest non-breaking version from available versions
func (d *UpdateDetector) GetLatestCompatible(current *semver.Version, available []*semver.Version) *semver.Version {
	var latest *semver.Version

	for _, v := range available {
		if v.GreaterThan(current) && !d.IsBreakingChange(current, v) {
			if latest == nil || v.GreaterThan(latest) {
				latest = v
			}
		}
	}

	return latest
}

package registry

import (
	"fmt"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// UpdateInfo contains information about available updates for a ritual
type UpdateInfo struct {
	RitualName     string
	CurrentVersion string
	LatestVersion  string
	IsUpdateNeeded bool
	Changelog      string
}

// CheckForUpdates checks if a newer version of a ritual is available
func (r *Registry) CheckForUpdates(ritualName string, currentVersion string) (*UpdateInfo, error) {
	meta, err := r.Get(ritualName)
	if err != nil {
		return nil, fmt.Errorf("ritual not found: %w", err)
	}

	// Compare versions
	isNewer, err := IsVersionNewer(meta.Version, currentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to compare versions: %w", err)
	}

	info := &UpdateInfo{
		RitualName:     ritualName,
		CurrentVersion: currentVersion,
		LatestVersion:  meta.Version,
		IsUpdateNeeded: isNewer,
		Changelog:      "", // TODO: Load from changelog file
	}

	return info, nil
}

// CheckAllUpdates checks for updates for all installed rituals
func (r *Registry) CheckAllUpdates(installedRituals map[string]string) ([]*UpdateInfo, error) {
	var updates []*UpdateInfo

	for name, version := range installedRituals {
		info, err := r.CheckForUpdates(name, version)
		if err != nil {
			// Skip rituals that aren't found in registry
			continue
		}

		if info.IsUpdateNeeded {
			updates = append(updates, info)
		}
	}

	return updates, nil
}

// GetLatestVersion returns the latest version of a ritual
func (r *Registry) GetLatestVersion(ritualName string) (string, error) {
	meta, err := r.Get(ritualName)
	if err != nil {
		return "", err
	}

	return meta.Version, nil
}

// LoadChangelog loads the changelog for a ritual version
func (r *Registry) LoadChangelog(ritualName string, fromVersion string, toVersion string) (string, error) {
	meta, err := r.Get(ritualName)
	if err != nil {
		return "", err
	}

	loader := ritual.NewLoader(meta.Path)
	manifest, err := loader.Load(meta.Path)
	if err != nil {
		return "", fmt.Errorf("failed to load ritual: %w", err)
	}

	// Look for CHANGELOG.md in ritual directory
	_ = manifest // TODO: Actually load and parse changelog

	return fmt.Sprintf("Version %s -> %s\n\nNo changelog available.", fromVersion, toVersion), nil
}

// NotifyUpdate represents an update notification
type NotifyUpdate struct {
	RitualName string
	Message    string
}

// GetUpdateNotifications returns formatted update notifications
func (r *Registry) GetUpdateNotifications(installedRituals map[string]string) ([]NotifyUpdate, error) {
	updates, err := r.CheckAllUpdates(installedRituals)
	if err != nil {
		return nil, err
	}

	var notifications []NotifyUpdate
	for _, update := range updates {
		msg := fmt.Sprintf("Update available for %s: %s -> %s",
			update.RitualName,
			update.CurrentVersion,
			update.LatestVersion,
		)
		notifications = append(notifications, NotifyUpdate{
			RitualName: update.RitualName,
			Message:    msg,
		})
	}

	return notifications, nil
}

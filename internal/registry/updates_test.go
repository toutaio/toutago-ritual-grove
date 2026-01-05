package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckForUpdates(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ritual-updates-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test ritual
	ritualDir := filepath.Join(tmpDir, "test-ritual")
	if err := os.MkdirAll(ritualDir, 0755); err != nil {
		t.Fatalf("Failed to create ritual dir: %v", err)
	}

	ritualContent := `ritual:
  name: update-test
  version: 2.0.0
  description: Test ritual for updates

compatibility:
  min_touta_version: 0.1.0

dependencies:
  go_packages: []
  other_rituals: []

questions: []

files:
  templates: []
  static: []
  protected: []
`
	if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), []byte(ritualContent), 0644); err != nil {
		t.Fatalf("Failed to write ritual.yaml: %v", err)
	}

	// Create registry and index the ritual
	registry := &Registry{
		cache:       make(map[string]*RitualMetadata),
		cacheDir:    tmpDir,
		searchPaths: []string{tmpDir},
	}

	if err := registry.indexRitual(ritualDir, SourceLocal); err != nil {
		t.Fatalf("Failed to index ritual: %v", err)
	}

	tests := []struct {
		name           string
		currentVersion string
		wantUpdate     bool
	}{
		{
			name:           "older version",
			currentVersion: "1.0.0",
			wantUpdate:     true,
		},
		{
			name:           "same version",
			currentVersion: "2.0.0",
			wantUpdate:     false,
		},
		{
			name:           "newer version (local is old)",
			currentVersion: "3.0.0",
			wantUpdate:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := registry.CheckForUpdates("update-test", tt.currentVersion)
			if err != nil {
				t.Fatalf("CheckForUpdates failed: %v", err)
			}

			if info.IsUpdateNeeded != tt.wantUpdate {
				t.Errorf("IsUpdateNeeded = %v, want %v", info.IsUpdateNeeded, tt.wantUpdate)
			}

			if info.RitualName != "update-test" {
				t.Errorf("RitualName = %s, want update-test", info.RitualName)
			}

			if info.CurrentVersion != tt.currentVersion {
				t.Errorf("CurrentVersion = %s, want %s", info.CurrentVersion, tt.currentVersion)
			}

			if info.LatestVersion != "2.0.0" {
				t.Errorf("LatestVersion = %s, want 2.0.0", info.LatestVersion)
			}
		})
	}
}

func TestCheckForUpdatesNotFound(t *testing.T) {
	registry := &Registry{
		cache:    make(map[string]*RitualMetadata),
		cacheDir: "/tmp",
	}

	_, err := registry.CheckForUpdates("nonexistent", "1.0.0")
	if err == nil {
		t.Error("Expected error for nonexistent ritual, got nil")
	}
}

func TestCheckAllUpdates(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ritual-all-updates-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple test rituals
	rituals := []struct {
		name    string
		version string
	}{
		{"ritual-a", "2.0.0"},
		{"ritual-b", "1.5.0"},
		{"ritual-c", "3.0.0"},
	}

	registry := &Registry{
		cache:       make(map[string]*RitualMetadata),
		cacheDir:    tmpDir,
		searchPaths: []string{tmpDir},
	}

	for _, r := range rituals {
		ritualDir := filepath.Join(tmpDir, r.name)
		if err := os.MkdirAll(ritualDir, 0755); err != nil {
			t.Fatalf("Failed to create ritual dir: %v", err)
		}

		content := `ritual:
  name: ` + r.name + `
  version: ` + r.version + `
  description: Test ritual

compatibility:
  min_touta_version: 0.1.0

dependencies:
  go_packages: []
  other_rituals: []

questions: []

files:
  templates: []
  static: []
  protected: []
`
		if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write ritual.yaml: %v", err)
		}

		if err := registry.indexRitual(ritualDir, SourceLocal); err != nil {
			t.Fatalf("Failed to index ritual: %v", err)
		}
	}

	// Check for updates with various installed versions
	installed := map[string]string{
		"ritual-a": "1.0.0", // Update available
		"ritual-b": "1.5.0", // No update
		"ritual-c": "2.0.0", // Update available
	}

	updates, err := registry.CheckAllUpdates(installed)
	if err != nil {
		t.Fatalf("CheckAllUpdates failed: %v", err)
	}

	if len(updates) != 2 {
		t.Errorf("Expected 2 updates, got %d", len(updates))
	}

	// Verify the updates
	updateMap := make(map[string]*UpdateInfo)
	for _, update := range updates {
		updateMap[update.RitualName] = update
	}

	if update, exists := updateMap["ritual-a"]; !exists {
		t.Error("Expected update for ritual-a")
	} else {
		if update.LatestVersion != "2.0.0" {
			t.Errorf("Expected latest version 2.0.0 for ritual-a, got %s", update.LatestVersion)
		}
	}

	if update, exists := updateMap["ritual-c"]; !exists {
		t.Error("Expected update for ritual-c")
	} else {
		if update.LatestVersion != "3.0.0" {
			t.Errorf("Expected latest version 3.0.0 for ritual-c, got %s", update.LatestVersion)
		}
	}

	if _, exists := updateMap["ritual-b"]; exists {
		t.Error("Did not expect update for ritual-b (same version)")
	}
}

func TestGetLatestVersion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ritual-latest-version-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test ritual
	ritualDir := filepath.Join(tmpDir, "version-test")
	if err := os.MkdirAll(ritualDir, 0755); err != nil {
		t.Fatalf("Failed to create ritual dir: %v", err)
	}

	ritualContent := `ritual:
  name: version-test
  version: 5.2.1
  description: Test version retrieval

compatibility:
  min_touta_version: 0.1.0

dependencies:
  go_packages: []
  other_rituals: []

questions: []

files:
  templates: []
  static: []
  protected: []
`
	if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), []byte(ritualContent), 0644); err != nil {
		t.Fatalf("Failed to write ritual.yaml: %v", err)
	}

	registry := &Registry{
		cache:       make(map[string]*RitualMetadata),
		cacheDir:    tmpDir,
		searchPaths: []string{tmpDir},
	}

	if err := registry.indexRitual(ritualDir, SourceLocal); err != nil {
		t.Fatalf("Failed to index ritual: %v", err)
	}

	version, err := registry.GetLatestVersion("version-test")
	if err != nil {
		t.Fatalf("GetLatestVersion failed: %v", err)
	}

	if version != "5.2.1" {
		t.Errorf("GetLatestVersion() = %s, want 5.2.1", version)
	}
}

func TestGetUpdateNotifications(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ritual-notifications-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test ritual
	ritualDir := filepath.Join(tmpDir, "notify-test")
	if err := os.MkdirAll(ritualDir, 0755); err != nil {
		t.Fatalf("Failed to create ritual dir: %v", err)
	}

	ritualContent := `ritual:
  name: notify-test
  version: 2.5.0
  description: Test notifications

compatibility:
  min_touta_version: 0.1.0

dependencies:
  go_packages: []
  other_rituals: []

questions: []

files:
  templates: []
  static: []
  protected: []
`
	if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), []byte(ritualContent), 0644); err != nil {
		t.Fatalf("Failed to write ritual.yaml: %v", err)
	}

	registry := &Registry{
		cache:       make(map[string]*RitualMetadata),
		cacheDir:    tmpDir,
		searchPaths: []string{tmpDir},
	}

	if err := registry.indexRitual(ritualDir, SourceLocal); err != nil {
		t.Fatalf("Failed to index ritual: %v", err)
	}

	installed := map[string]string{
		"notify-test": "1.0.0",
	}

	notifications, err := registry.GetUpdateNotifications(installed)
	if err != nil {
		t.Fatalf("GetUpdateNotifications failed: %v", err)
	}

	if len(notifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(notifications))
	}

	notification := notifications[0]
	if notification.RitualName != "notify-test" {
		t.Errorf("RitualName = %s, want notify-test", notification.RitualName)
	}

	expectedMsg := "Update available for notify-test: 1.0.0 -> 2.5.0"
	if notification.Message != expectedMsg {
		t.Errorf("Message = %s, want %s", notification.Message, expectedMsg)
	}
}

func TestLoadChangelog(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ritual-changelog-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test ritual
	ritualDir := filepath.Join(tmpDir, "changelog-test")
	if err := os.MkdirAll(ritualDir, 0755); err != nil {
		t.Fatalf("Failed to create ritual dir: %v", err)
	}

	ritualContent := `ritual:
  name: changelog-test
  version: 2.0.0
  description: Test changelog loading

compatibility:
  min_touta_version: 0.1.0

dependencies:
  go_packages: []
  other_rituals: []

questions: []

files:
  templates: []
  static: []
  protected: []
`
	if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), []byte(ritualContent), 0644); err != nil {
		t.Fatalf("Failed to write ritual.yaml: %v", err)
	}

	registry := &Registry{
		cache:       make(map[string]*RitualMetadata),
		cacheDir:    tmpDir,
		searchPaths: []string{tmpDir},
	}

	if err := registry.indexRitual(ritualDir, SourceLocal); err != nil {
		t.Fatalf("Failed to index ritual: %v", err)
	}

	changelog, err := registry.LoadChangelog("changelog-test", "1.0.0", "2.0.0")
	if err != nil {
		t.Fatalf("LoadChangelog failed: %v", err)
	}

	// For now, we expect a placeholder message
	if changelog == "" {
		t.Error("Expected non-empty changelog")
	}
}

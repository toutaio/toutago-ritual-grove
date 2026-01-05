package deployment

import (
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestUpdateDetector_CompareVersions(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		availableVer   string
		wantUpdate     bool
		wantBreaking   bool
	}{
		{
			name:           "patch update available",
			currentVersion: "1.0.0",
			availableVer:   "1.0.1",
			wantUpdate:     true,
			wantBreaking:   false,
		},
		{
			name:           "minor update available",
			currentVersion: "1.0.0",
			availableVer:   "1.1.0",
			wantUpdate:     true,
			wantBreaking:   false,
		},
		{
			name:           "major update available (breaking)",
			currentVersion: "1.0.0",
			availableVer:   "2.0.0",
			wantUpdate:     true,
			wantBreaking:   true,
		},
		{
			name:           "same version",
			currentVersion: "1.0.0",
			availableVer:   "1.0.0",
			wantUpdate:     false,
			wantBreaking:   false,
		},
		{
			name:           "older version available",
			currentVersion: "1.1.0",
			availableVer:   "1.0.0",
			wantUpdate:     false,
			wantBreaking:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewUpdateDetector()
			current := mustParseVersion(t, tt.currentVersion)
			available := mustParseVersion(t, tt.availableVer)

			updateAvailable := detector.IsUpdateAvailable(current, available)
			if updateAvailable != tt.wantUpdate {
				t.Errorf("IsUpdateAvailable() = %v, want %v", updateAvailable, tt.wantUpdate)
			}

			isBreaking := detector.IsBreakingChange(current, available)
			if isBreaking != tt.wantBreaking {
				t.Errorf("IsBreakingChange() = %v, want %v", isBreaking, tt.wantBreaking)
			}
		})
	}
}

func TestUpdateDetector_ListUpdates(t *testing.T) {
	detector := NewUpdateDetector()
	current := mustParseVersion(t, "1.0.0")
	
	available := []*semver.Version{
		mustParseVersion(t, "0.9.0"),
		mustParseVersion(t, "1.0.0"),
		mustParseVersion(t, "1.0.1"),
		mustParseVersion(t, "1.1.0"),
		mustParseVersion(t, "2.0.0"),
	}

	updates := detector.ListUpdates(current, available)
	
	expectedCount := 3 // 1.0.1, 1.1.0, 2.0.0
	if len(updates) != expectedCount {
		t.Errorf("ListUpdates() returned %d updates, want %d", len(updates), expectedCount)
	}

	// Verify updates are sorted (newest first)
	if len(updates) > 1 {
		for i := 0; i < len(updates)-1; i++ {
			if updates[i].LessThan(updates[i+1]) {
				t.Errorf("Updates not sorted correctly: %v should be >= %v", updates[i], updates[i+1])
			}
		}
	}
}

func TestUpdateDetector_GetUpdateInfo(t *testing.T) {
	detector := NewUpdateDetector()
	current := mustParseVersion(t, "1.0.0")
	target := mustParseVersion(t, "1.1.0")

	info := detector.GetUpdateInfo(current, target)

	if info.FromVersion != "1.0.0" {
		t.Errorf("FromVersion = %v, want 1.0.0", info.FromVersion)
	}
	if info.ToVersion != "1.1.0" {
		t.Errorf("ToVersion = %v, want 1.1.0", info.ToVersion)
	}
	if info.UpdateType != "minor" {
		t.Errorf("UpdateType = %v, want minor", info.UpdateType)
	}
	if info.IsBreaking {
		t.Errorf("IsBreaking = true, want false for minor update")
	}
}

func TestUpdateDetector_GetLatestCompatible(t *testing.T) {
	detector := NewUpdateDetector()
	current := mustParseVersion(t, "1.2.3")
	
	available := []*semver.Version{
		mustParseVersion(t, "1.2.4"),
		mustParseVersion(t, "1.3.0"),
		mustParseVersion(t, "2.0.0"), // breaking
	}

	// Get latest non-breaking
	latest := detector.GetLatestCompatible(current, available)
	if latest == nil {
		t.Fatal("GetLatestCompatible() returned nil")
	}
	if latest.String() != "1.3.0" {
		t.Errorf("GetLatestCompatible() = %v, want 1.3.0", latest)
	}
}

func mustParseVersion(t *testing.T, v string) *semver.Version {
	t.Helper()
	version, err := semver.NewVersion(v)
	if err != nil {
		t.Fatalf("Failed to parse version %s: %v", v, err)
	}
	return version
}

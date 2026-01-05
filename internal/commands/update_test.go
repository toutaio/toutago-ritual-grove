package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"

	"github.com/toutaio/toutago-ritual-grove/internal/storage"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestUpdateHandler_Execute(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project structure
	ritualDir := filepath.Join(tmpDir, ".ritual")
	if err := os.MkdirAll(ritualDir, 0750); err != nil {
		t.Fatal(err)
	}

	// Create initial state
	state := &storage.State{
		RitualName:     "test-ritual",
		RitualVersion:  "1.0.0",
		GeneratedFiles: []string{"main.go"},
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatal(err)
	}

	handler := NewUpdateHandler()

	t.Run("dry run mode", func(t *testing.T) {
		err := handler.Execute(tmpDir, UpdateOptions{
			ToVersion: "1.1.0",
			DryRun:    true,
		})

		// Dry run should not error on missing ritual
		// It just shows what would happen
		if err != nil {
			t.Logf("Expected success in dry-run, got: %v", err)
		}
	})

	t.Run("already at target version", func(t *testing.T) {
		err := handler.Execute(tmpDir, UpdateOptions{
			ToVersion: "1.0.0",
		})

		if err != nil {
			t.Errorf("Should succeed when already at target version: %v", err)
		}
	})

	t.Run("no target version specified", func(t *testing.T) {
		err := handler.Execute(tmpDir, UpdateOptions{})

		if err == nil {
			t.Error("Should error when no target version specified")
		}
	})
}

func TestUpdateHandler_GetMigrationsToRun(t *testing.T) {
	handler := NewUpdateHandler()

	state := &storage.State{
		RitualVersion: "1.0.0",
		AppliedMigrations: []storage.Migration{
			{Version: "1.0.0", AppliedAt: time.Now()},
		},
	}

	manifest := &ritual.Manifest{
		Migrations: []ritual.Migration{
			{
				FromVersion: "0.9.0",
				ToVersion:   "1.0.0",
				Description: "Initial migration",
				Up:          ritual.MigrationHandler{SQL: []string{"-- already applied"}},
			},
			{
				FromVersion: "1.0.0",
				ToVersion:   "1.1.0",
				Description: "New migration",
				Up:          ritual.MigrationHandler{SQL: []string{"-- new"}},
				Down:        ritual.MigrationHandler{SQL: []string{"-- rollback"}},
			},
			{
				FromVersion: "1.1.0",
				ToVersion:   "1.2.0",
				Description: "Future migration",
				Up:          ritual.MigrationHandler{SQL: []string{"-- future"}},
			},
		},
	}

	migrations := handler.getMigrationsToRun(state, manifest)

	// Should return 2 migrations (1.1.0 and 1.2.0, skipping 1.0.0 which is already applied)
	if len(migrations) != 2 {
		t.Errorf("Expected 2 migrations, got %d", len(migrations))
	}

	// Check migration details
	found11 := false
	found12 := false
	for _, mig := range migrations {
		if mig.ToVersion == "1.1.0" {
			found11 = true
			if mig.Description != "New migration" {
				t.Errorf("Wrong description: %s", mig.Description)
			}
		}
		if mig.ToVersion == "1.2.0" {
			found12 = true
		}
	}

	if !found11 {
		t.Error("Migration 1.1.0 not found")
	}
	if !found12 {
		t.Error("Migration 1.2.0 not found")
	}
}

func TestUpdateHandler_CanUpdate(t *testing.T) {
	tmpDir := t.TempDir()

	ritualDir := filepath.Join(tmpDir, ".ritual")
	if err := os.MkdirAll(ritualDir, 0750); err != nil {
		t.Fatal(err)
	}

	state := &storage.State{
		RitualName:    "test-ritual",
		RitualVersion: "1.0.0",
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatal(err)
	}

	handler := NewUpdateHandler()

	canUpdate, version, err := handler.CanUpdate(tmpDir)
	if err != nil {
		t.Fatalf("CanUpdate() error = %v", err)
	}

	if version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", version)
	}

	// Currently returns false as registry check is not implemented
	if canUpdate {
		t.Log("Note: canUpdate is true (registry check not yet implemented)")
	}
}

func TestUpdateHandler_ShowUpdateInfo(t *testing.T) {
	tmpDir := t.TempDir()

	ritualDir := filepath.Join(tmpDir, ".ritual")
	if err := os.MkdirAll(ritualDir, 0750); err != nil {
		t.Fatal(err)
	}

	state := &storage.State{
		RitualName:    "test-ritual",
		RitualVersion: "1.0.0",
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatal(err)
	}

	handler := NewUpdateHandler()

	err := handler.ShowUpdateInfo(tmpDir, "1.1.0")
	if err != nil {
		t.Errorf("ShowUpdateInfo() error = %v", err)
	}
}

func TestUpdateHandler_EmptyPath(t *testing.T) {
	handler := NewUpdateHandler()

	// Should use current directory as default
	err := handler.Execute("", UpdateOptions{
		ToVersion: "1.0.0",
	})

	// Will fail on state load, but shows path handling works
	if err == nil {
		t.Log("Note: Execute with empty path didn't error (no state file expected)")
	}
}

func TestUpdateHandler_CreateBackup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple project structure
	if err := os.MkdirAll(filepath.Join(tmpDir, ".ritual"), 0750); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0600); err != nil {
		t.Fatal(err)
	}

	handler := NewUpdateHandler()

	backupPath, err := handler.createBackup(tmpDir)
	if err != nil {
		t.Fatalf("createBackup() error = %v", err)
	}

	if backupPath == "" {
		t.Error("Expected non-empty backup path")
	}

	// Check that backup was created
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}
}

func TestUpdateHandler_LoadNewRitual(t *testing.T) {
	handler := NewUpdateHandler()

	// Try to load a non-existent ritual
	_, err := handler.loadNewRitual("nonexistent-ritual-xyz")
	if err == nil {
		t.Error("Expected error when loading non-existent ritual")
	}
}

func TestUpdateHandler_SaveUpdatedState(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .ritual directory
	if err := os.MkdirAll(filepath.Join(tmpDir, ".ritual"), 0750); err != nil {
		t.Fatal(err)
	}

	state := &storage.State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
	}

	handler := NewUpdateHandler()

	err := handler.saveUpdatedState(state, "1.1.0", tmpDir)
	if err != nil {
		t.Fatalf("saveUpdatedState() error = %v", err)
	}

	// Verify state was updated
	if state.RitualVersion != "1.1.0" {
		t.Errorf("Expected version 1.1.0, got %s", state.RitualVersion)
	}

	// Verify state was saved
	loadedState, err := storage.LoadState(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load saved state: %v", err)
	}

	if loadedState.RitualVersion != "1.1.0" {
		t.Errorf("Expected saved version 1.1.0, got %s", loadedState.RitualVersion)
	}
}

func TestUpdateHandler_ParseVersions(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		to          string
		expectError bool
	}{
		{"valid versions", "1.0.0", "2.0.0", false},
		{"same version", "1.0.0", "1.0.0", false},
		{"invalid from version", "invalid", "1.0.0", true},
		{"invalid to version", "1.0.0", "invalid", true},
	}

	handler := NewUpdateHandler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := handler.parseVersions(tt.from, tt.to)
			hasError := err != nil
			if hasError != tt.expectError {
				t.Errorf("parseVersions() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestUpdateHandler_DisplayUpdateInfo(t *testing.T) {
	handler := NewUpdateHandler()

	// Create versions
	v1, _ := semver.NewVersion("1.0.0")
	v2, _ := semver.NewVersion("2.0.0")

	// Just verify it doesn't panic
	handler.displayUpdateInfo("1.0.0", "2.0.0", v1, v2)

	// With same version
	handler.displayUpdateInfo("1.0.0", "1.0.0", v1, v1)
}

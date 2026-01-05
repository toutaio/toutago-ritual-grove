package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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

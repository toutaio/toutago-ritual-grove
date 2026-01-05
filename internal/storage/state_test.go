package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestState_Save(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, ".ritual", "state.yaml")

	state := &State{
		RitualName:    "test-ritual",
		RitualVersion: "1.0.0",
		InstalledAt:   time.Now(),
		AppliedMigrations: []Migration{
			{Version: "1.0.0", AppliedAt: time.Now()},
		},
		GeneratedFiles: []string{"main.go", "config.yaml"},
		ProtectedFiles: []string{"custom.go"},
	}

	err := state.Save(tmpDir)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Errorf("State file was not created at %s", statePath)
	}
}

func TestState_Load(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".ritual")
	os.MkdirAll(stateDir, 0755)

	// Create a state file manually
	originalState := &State{
		RitualName:    "test-ritual",
		RitualVersion: "1.2.3",
		InstalledAt:   time.Now().Round(time.Second),
		AppliedMigrations: []Migration{
			{Version: "1.0.0", AppliedAt: time.Now().Round(time.Second)},
		},
		GeneratedFiles: []string{"main.go"},
		ProtectedFiles: []string{"custom.go"},
	}

	err := originalState.Save(tmpDir)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Load the state
	loadedState, err := LoadState(tmpDir)
	if err != nil {
		t.Fatalf("LoadState() failed: %v", err)
	}

	// Verify loaded state matches
	if loadedState.RitualName != originalState.RitualName {
		t.Errorf("RitualName = %v, want %v", loadedState.RitualName, originalState.RitualName)
	}
	if loadedState.RitualVersion != originalState.RitualVersion {
		t.Errorf("RitualVersion = %v, want %v", loadedState.RitualVersion, originalState.RitualVersion)
	}
	if len(loadedState.AppliedMigrations) != len(originalState.AppliedMigrations) {
		t.Errorf("AppliedMigrations count = %v, want %v", len(loadedState.AppliedMigrations), len(originalState.AppliedMigrations))
	}
	if len(loadedState.GeneratedFiles) != len(originalState.GeneratedFiles) {
		t.Errorf("GeneratedFiles count = %v, want %v", len(loadedState.GeneratedFiles), len(originalState.GeneratedFiles))
	}
}

func TestLoadState_NotExists(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadState(tmpDir)
	if err == nil {
		t.Error("LoadState() should fail when state file doesn't exist")
	}
}

func TestState_AddMigration(t *testing.T) {
	state := &State{
		RitualName:        "test",
		RitualVersion:     "1.0.0",
		AppliedMigrations: []Migration{},
	}

	state.AddMigration("1.0.0")

	if len(state.AppliedMigrations) != 1 {
		t.Errorf("Expected 1 migration, got %d", len(state.AppliedMigrations))
	}

	if state.AppliedMigrations[0].Version != "1.0.0" {
		t.Errorf("Migration version = %v, want 1.0.0", state.AppliedMigrations[0].Version)
	}
}

func TestState_IsMigrationApplied(t *testing.T) {
	state := &State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
		AppliedMigrations: []Migration{
			{Version: "1.0.0", AppliedAt: time.Now()},
			{Version: "1.1.0", AppliedAt: time.Now()},
		},
	}

	tests := []struct {
		version string
		want    bool
	}{
		{"1.0.0", true},
		{"1.1.0", true},
		{"1.2.0", false},
		{"2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := state.IsMigrationApplied(tt.version); got != tt.want {
				t.Errorf("IsMigrationApplied(%v) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestState_MarkFileAsGenerated(t *testing.T) {
	state := &State{
		GeneratedFiles: []string{},
	}

	state.MarkFileAsGenerated("main.go")
	state.MarkFileAsGenerated("config.yaml")

	if len(state.GeneratedFiles) != 2 {
		t.Errorf("Expected 2 generated files, got %d", len(state.GeneratedFiles))
	}

	// Should not add duplicates
	state.MarkFileAsGenerated("main.go")
	if len(state.GeneratedFiles) != 2 {
		t.Errorf("Expected 2 generated files (no duplicates), got %d", len(state.GeneratedFiles))
	}
}

func TestState_MarkFileAsProtected(t *testing.T) {
	state := &State{
		ProtectedFiles: []string{},
	}

	state.MarkFileAsProtected("custom.go")

	if len(state.ProtectedFiles) != 1 {
		t.Errorf("Expected 1 protected file, got %d", len(state.ProtectedFiles))
	}

	// Should not add duplicates
	state.MarkFileAsProtected("custom.go")
	if len(state.ProtectedFiles) != 1 {
		t.Errorf("Expected 1 protected file (no duplicates), got %d", len(state.ProtectedFiles))
	}
}

func TestState_IsFileGenerated(t *testing.T) {
	state := &State{
		GeneratedFiles: []string{"main.go", "config.yaml"},
	}

	tests := []struct {
		file string
		want bool
	}{
		{"main.go", true},
		{"config.yaml", true},
		{"custom.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			if got := state.IsFileGenerated(tt.file); got != tt.want {
				t.Errorf("IsFileGenerated(%v) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}

func TestState_IsFileProtected(t *testing.T) {
	state := &State{
		ProtectedFiles: []string{"custom.go", "user-config.yaml"},
	}

	tests := []struct {
		file string
		want bool
	}{
		{"custom.go", true},
		{"user-config.yaml", true},
		{"main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			if got := state.IsFileProtected(tt.file); got != tt.want {
				t.Errorf("IsFileProtected(%v) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}

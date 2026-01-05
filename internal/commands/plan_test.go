package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/storage"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestPlanCommand(t *testing.T) {
	// Create temp project directory
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Create .ritual directory and state
	ritualDir := filepath.Join(tmpDir, ".ritual")
	if err := os.MkdirAll(ritualDir, 0755); err != nil {
		t.Fatalf("Failed to create .ritual directory: %v", err)
	}

	// Create initial state
	state := &storage.State{
		RitualName:    "test-ritual",
		RitualVersion: "1.0.0",
		GeneratedFiles: []string{
			"main.go",
		},
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create current ritual manifest
	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{Source: "main.go.tmpl", Destination: "main.go"},
			},
		},
	}

	// Write manifest to yaml file manually for now (SaveManifest doesn't exist in API)
	manifestPath := filepath.Join(ritualDir, "ritual.yaml")
	manifestData := `ritual:
  name: test-ritual
  version: 1.0.0
files:
  templates:
    - src: main.go.tmpl
      dest: main.go
`
	if err := os.WriteFile(manifestPath, []byte(manifestData), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Test command creation
	cmd := NewPlanCommand()
	if cmd == nil {
		t.Fatal("NewPlanCommand() returned nil")
	}

	if cmd.Use != "plan" {
		t.Errorf("Expected Use='plan', got %s", cmd.Use)
	}

	// Test flags
	if !cmd.Flags().HasFlags() {
		t.Error("Expected command to have flags")
	}

	toVersionFlag := cmd.Flags().Lookup("to-version")
	if toVersionFlag == nil {
		t.Error("Expected --to-version flag")
	}

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("Expected --json flag")
	}

	// Suppress output
	_ = manifest // Use variable to avoid unused warnings
}

func TestPlanCommand_NoState(t *testing.T) {
	// Create temp directory without state
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	cmd := NewPlanCommand()
	err := cmd.RunE(cmd, []string{})

	if err == nil {
		t.Error("Expected error when no state file exists")
	}
}

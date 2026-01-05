package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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
	if err := os.MkdirAll(ritualDir, 0750); err != nil {
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
	if err := os.WriteFile(manifestPath, []byte(manifestData), 0600); err != nil {
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

func TestPlanCommand_EmptyRitualName(t *testing.T) {
	// Create temp directory with empty ritual state
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Create state with empty ritual name
	state := &storage.State{
		RitualName:    "",
		RitualVersion: "1.0.0",
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	cmd := NewPlanCommand()
	err := cmd.RunE(cmd, []string{})

	if err == nil {
		t.Error("Expected error when ritual name is empty")
	}
	if !strings.Contains(err.Error(), "no ritual found") {
		t.Errorf("Expected 'no ritual found' error, got: %v", err)
	}
}

func TestPlanCommand_JSONOutputNotImplemented(t *testing.T) {
	// Create temp project directory
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Create .ritual directory and state
	ritualDir := filepath.Join(tmpDir, ".ritual")
	if err := os.MkdirAll(ritualDir, 0750); err != nil {
		t.Fatalf("Failed to create .ritual directory: %v", err)
	}

	// Create state
	state := &storage.State{
		RitualName:    "basic-site",
		RitualVersion: "0.1.0",
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Write manifest
	manifestPath := filepath.Join(ritualDir, "ritual.yaml")
	manifestData := `ritual:
  name: basic-site
  version: 0.1.0
`
	if err := os.WriteFile(manifestPath, []byte(manifestData), 0600); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Create a built-in ritual in rituals directory (basic-site exists)
	// The registry should find it

	cmd := NewPlanCommand()
	cmd.SetArgs([]string{"--json"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()

	// Should error since JSON output is not implemented
	if err == nil {
		t.Error("Expected error for JSON output not implemented")
	}
	if !strings.Contains(err.Error(), "JSON output not yet implemented") {
		t.Errorf("Expected 'JSON output not yet implemented' error, got: %v", err)
	}
}

func TestPlanCommand_CommandStructure(t *testing.T) {
	cmd := NewPlanCommand()

	tests := []struct {
		name     string
		check    func() bool
		expected bool
	}{
		{"has Use field", func() bool { return cmd.Use == "plan" }, true},
		{"has Short description", func() bool { return cmd.Short != "" }, true},
		{"has Long description", func() bool { return cmd.Long != "" }, true},
		{"has RunE function", func() bool { return cmd.RunE != nil }, true},
		{"has to-version flag", func() bool { return cmd.Flags().Lookup("to-version") != nil }, true},
		{"has json flag", func() bool { return cmd.Flags().Lookup("json") != nil }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.check(); result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRunPlan_InvalidDirectory(t *testing.T) {
	// Save current dir
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	// Try to change to non-existent directory (will fail, but that's ok for test)
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	os.RemoveAll(tmpDir) // Remove it to make Getwd potentially fail in some edge cases

	cmd := NewPlanCommand()

	// This should fail when trying to load state
	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when project directory is invalid")
	}
}

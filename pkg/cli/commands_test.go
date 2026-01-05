package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestInitRitual(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Create a simple test ritual
	ritualDir := filepath.Join(tmpDir, "test-ritual")
	if err := os.MkdirAll(ritualDir, 0750); err != nil {
		t.Fatalf("Failed to create ritual dir: %v", err)
	}

	// Create ritual.yaml
	ritualYAML := `ritual:
  name: test-ritual
  version: 1.0.0
  description: Test ritual
  author: Test

compatibility:
  min_touta_version: "0.1.0"
  min_go_version: "1.22"

questions:
  - name: project_name
    type: text
    prompt: "Project name?"
    required: true
    default: "test-project"

files:
  templates:
    - src: templates/main.go.tmpl
      dest: main.go

hooks:
  pre_install: []
  post_install: []
`

	ritualYAMLPath := filepath.Join(ritualDir, "ritual.yaml")
	if err := os.WriteFile(ritualYAMLPath, []byte(ritualYAML), 0600); err != nil {
		t.Fatalf("Failed to create ritual.yaml: %v", err)
	}

	// Create templates directory
	templatesDir := filepath.Join(ritualDir, "templates")
	if err := os.MkdirAll(templatesDir, 0750); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Create a simple template
	mainTemplate := `package main

func main() {
	println("Hello {{ .project_name }}")
}
`

	mainTemplatePath := filepath.Join(templatesDir, "main.go.tmpl")
	if err := os.WriteFile(mainTemplatePath, []byte(mainTemplate), 0600); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Test initialization
	outputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Note: This test requires a registry with the ritual available
	// For now, we'll just test the validation
	t.Log("Init ritual test setup complete")
}

func TestRitualCommand(t *testing.T) {
	cmd := RitualCommand()
	
	if cmd == nil {
		t.Fatal("RitualCommand() returned nil")
	}
	
	if cmd.Use != "ritual" {
		t.Errorf("Expected Use='ritual', got %s", cmd.Use)
	}
	
	// Check subcommands exist
	subcommands := []string{"init", "list", "info", "validate", "create", "plan"}
	for _, subcmd := range subcommands {
		found := false
		for _, c := range cmd.Commands() {
			if c.Use == subcmd || c.Use == subcmd+" <ritual-name>" || c.Use == subcmd+" <name>" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Subcommand %s not found", subcmd)
		}
	}
}

func TestInitCommand(t *testing.T) {
	ritualCmd := RitualCommand()
	var initCmd *cobra.Command
	for _, cmd := range ritualCmd.Commands() {
		if cmd.Name() == "init" {
			initCmd = cmd
			break
		}
	}
	
	if initCmd == nil {
		t.Fatal("init command not found")
	}
	
	// Check flags
	outputFlag := initCmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("Expected --output flag")
	}
	
	yesFlag := initCmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Error("Expected --yes flag")
	}
}

func TestListCommand(t *testing.T) {
	ritualCmd := RitualCommand()
	var listCmd *cobra.Command
	for _, cmd := range ritualCmd.Commands() {
		if cmd.Name() == "list" {
			listCmd = cmd
			break
		}
	}
	
	if listCmd == nil {
		t.Fatal("list command not found")
	}
	
	if listCmd.Use != "list" {
		t.Errorf("Expected Use='list', got %s", listCmd.Use)
	}
}

func TestInfoCommand(t *testing.T) {
	ritualCmd := RitualCommand()
	var infoCmd *cobra.Command
	for _, cmd := range ritualCmd.Commands() {
		if cmd.Name() == "info" {
			infoCmd = cmd
			break
		}
	}
	
	if infoCmd == nil {
		t.Fatal("info command not found")
	}
}

func TestValidateCommand(t *testing.T) {
	ritualCmd := RitualCommand()
	var validateCmd *cobra.Command
	for _, cmd := range ritualCmd.Commands() {
		if cmd.Name() == "validate" {
			validateCmd = cmd
			break
		}
	}
	
	if validateCmd == nil {
		t.Fatal("validate command not found")
	}
}

func TestPlanCommand(t *testing.T) {
	ritualCmd := RitualCommand()
	var planCmd *cobra.Command
	for _, cmd := range ritualCmd.Commands() {
		if cmd.Name() == "plan" {
			planCmd = cmd
			break
		}
	}
	
	if planCmd == nil {
		t.Fatal("plan command not found")
	}
}

func TestValidateRitual(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid ritual.yaml
	ritualYAML := `ritual:
  name: test
  version: 1.0.0
  description: Test
  author: Test

compatibility:
  min_touta_version: "0.1.0"

questions: []
files:
  templates: []
hooks:
  pre_install: []
`

	ritualPath := filepath.Join(tmpDir, "ritual.yaml")
	if err := os.WriteFile(ritualPath, []byte(ritualYAML), 0600); err != nil {
		t.Fatalf("Failed to create ritual.yaml: %v", err)
	}

	// Test validation
	if err := validateRitual(tmpDir); err != nil {
		t.Errorf("Expected valid ritual, got error: %v", err)
	}
}

func TestValidateRitualInvalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid ritual.yaml (missing required fields)
	ritualYAML := `ritual:
  description: Test
`

	ritualPath := filepath.Join(tmpDir, "ritual.yaml")
	if err := os.WriteFile(ritualPath, []byte(ritualYAML), 0600); err != nil {
		t.Fatalf("Failed to create ritual.yaml: %v", err)
	}

	// Test validation should fail
	if err := validateRitual(tmpDir); err == nil {
		t.Error("Expected validation error for invalid ritual")
	}
}

func TestCreateRitual(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(oldDir)

	ritualName := "my-ritual"
	if err := createRitual(ritualName); err != nil {
		t.Fatalf("Failed to create ritual: %v", err)
	}

	// Check that directory was created
	ritualPath := filepath.Join(tmpDir, ritualName)
	if _, err := os.Stat(ritualPath); os.IsNotExist(err) {
		t.Error("Ritual directory was not created")
	}

	// Check that ritual.yaml was created
	yamlPath := filepath.Join(ritualPath, "ritual.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Error("ritual.yaml was not created")
	}

	// Check that subdirectories were created
	for _, dir := range []string{"templates", "static", "migrations"} {
		dirPath := filepath.Join(ritualPath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", dir)
		}
	}

	// Check that README.md was created
	readmePath := filepath.Join(ritualPath, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md was not created")
	}

	// Validate the created ritual
	if err := validateRitual(ritualPath); err != nil {
		t.Errorf("Created ritual is invalid: %v", err)
	}
}

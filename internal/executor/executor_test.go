package executor

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/generator"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestExecutor_Execute_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "ritual")
	outputDir := filepath.Join(tmpDir, "output")

	// Create ritual structure
	os.MkdirAll(filepath.Join(ritualDir, "templates"), 0755)

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{Source: "templates/main.go", Destination: "main.go"},
			},
		},
	}

	// Create template file
	templatePath := filepath.Join(ritualDir, "templates", "main.go")
	os.WriteFile(templatePath, []byte("package main"), 0644)

	vars := generator.NewVariables()
	
	context := &ExecutionContext{
		RitualPath: ritualDir,
		OutputPath: outputDir,
		Variables:  vars,
		DryRun:     true,
		Logger:     log.New(os.Stdout, "[test] ", 0),
	}

	executor := NewExecutor(context)
	
	// Should not error in dry run
	if err := executor.Execute(manifest); err != nil {
		t.Errorf("Execute failed in dry run: %v", err)
	}

	// Output directory should not be created in dry run
	if _, err := os.Stat(outputDir); err == nil {
		t.Error("Output directory should not exist in dry run")
	}
}

func TestExecutor_Execute_Real(t *testing.T) {
	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "ritual")
	outputDir := filepath.Join(tmpDir, "output")

	// Create ritual structure
	os.MkdirAll(filepath.Join(ritualDir, "templates"), 0755)

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{Source: "templates/main.go", Destination: "main.go"},
			},
		},
	}

	// Create template file
	templatePath := filepath.Join(ritualDir, "templates", "main.go")
	templateContent := "package main\n\nconst App = \"{{ .app_name }}\""
	os.WriteFile(templatePath, []byte(templateContent), 0644)

	vars := generator.NewVariables()
	vars.Set("app_name", "test-app")
	
	context := &ExecutionContext{
		RitualPath: ritualDir,
		OutputPath: outputDir,
		Variables:  vars,
		DryRun:     false,
		Logger:     log.New(os.Stdout, "[test] ", 0),
	}

	executor := NewExecutor(context)
	
	if err := executor.Execute(manifest); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check that file was generated
	outputFile := filepath.Join(outputDir, "main.go")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expected := "package main\n\nconst App = \"test-app\""
	if string(content) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(content))
	}
}

func TestExecutor_RunHooks_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	
	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
		Hooks: ritual.ManifestHooks{
			PreInstall: []string{"echo 'Setting up'"},
			PostInstall: []string{"echo 'Cleaning up'"},
		},
	}

	vars := generator.NewVariables()
	
	context := &ExecutionContext{
		RitualPath: tmpDir,
		OutputPath: tmpDir,
		Variables:  vars,
		DryRun:     true,
		Logger:     log.New(os.Stdout, "[test] ", 0),
	}

	executor := NewExecutor(context)
	
	// Should not error in dry run
	if err := executor.Execute(manifest); err != nil {
		t.Errorf("Execute with hooks failed in dry run: %v", err)
	}
}

func TestExecutor_InstallPackages_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	
	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
		Dependencies: ritual.Dependencies{
			Packages: []string{"github.com/lib/pq"},
		},
	}

	vars := generator.NewVariables()
	
	context := &ExecutionContext{
		RitualPath: tmpDir,
		OutputPath: tmpDir,
		Variables:  vars,
		DryRun:     true,
		Logger:     log.New(os.Stdout, "[test] ", 0),
	}

	executor := NewExecutor(context)
	
	// Should not error in dry run
	if err := executor.Execute(manifest); err != nil {
		t.Errorf("Execute with packages failed in dry run: %v", err)
	}
}

func TestExecutor_Rollback(t *testing.T) {
	tmpDir := t.TempDir()
	
	vars := generator.NewVariables()
	
	context := &ExecutionContext{
		RitualPath: tmpDir,
		OutputPath: tmpDir,
		Variables:  vars,
		DryRun:     false,
		Logger:     log.New(os.Stdout, "[test] ", 0),
	}

	executor := NewExecutor(context)
	
	// Should not error (even though not fully implemented)
	if err := executor.Rollback(); err != nil {
		t.Errorf("Rollback failed: %v", err)
	}
}

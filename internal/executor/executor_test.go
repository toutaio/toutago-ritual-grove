package executor

import (
	"log"
	"os"
	"os/exec"
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
			PreInstall:  []string{"echo 'Setting up'"},
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

func TestExecutor_Execute_WithStaticFiles(t *testing.T) {
	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "ritual")
	outputDir := filepath.Join(tmpDir, "output")

	// Create ritual structure
	os.MkdirAll(filepath.Join(ritualDir, "static"), 0755)

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
		Files: ritual.FilesSection{
			Static: []ritual.FileMapping{
				{Source: "static/config.json", Destination: "config.json"},
			},
		},
	}

	// Create static file
	staticPath := filepath.Join(ritualDir, "static", "config.json")
	os.WriteFile(staticPath, []byte(`{"app": "test"}`), 0644)

	vars := generator.NewVariables()

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

	// Check that file was copied
	outputFile := filepath.Join(outputDir, "config.json")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expected := `{"app": "test"}`
	if string(content) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(content))
	}
}

func TestExecutor_Execute_MissingSourceFile(t *testing.T) {
	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "ritual")
	outputDir := filepath.Join(tmpDir, "output")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{Source: "nonexistent.go", Destination: "main.go"},
			},
		},
	}

	vars := generator.NewVariables()

	context := &ExecutionContext{
		RitualPath: ritualDir,
		OutputPath: outputDir,
		Variables:  vars,
		DryRun:     false,
		Logger:     log.New(os.Stdout, "[test] ", 0),
	}

	executor := NewExecutor(context)

	// Should error when source file doesn't exist
	if err := executor.Execute(manifest); err == nil {
		t.Error("Expected error for missing source file, got nil")
	}
}

func TestExecutor_Execute_WithHooksAndPackages(t *testing.T) {
	// Skip if no go command available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}

	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "ritual")
	outputDir := filepath.Join(tmpDir, "output")

	// Create ritual structure
	os.MkdirAll(filepath.Join(ritualDir, "templates"), 0755)
	os.MkdirAll(outputDir, 0755)

	// Create go.mod first
	goModContent := "module testapp\n\ngo 1.21\n"
	os.WriteFile(filepath.Join(outputDir, "go.mod"), []byte(goModContent), 0644)

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
		Dependencies: ritual.Dependencies{
			Packages: []string{"github.com/stretchr/testify@v1.8.4"},
		},
		Hooks: ritual.ManifestHooks{
			PreInstall:  []string{"echo 'Pre-install'"},
			PostInstall: []string{"echo 'Post-install'"},
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
		DryRun:     false,
		Logger:     log.New(os.Stdout, "[test] ", 0),
	}

	executor := NewExecutor(context)

	if err := executor.Execute(manifest); err != nil {
		t.Fatalf("Execute with hooks and packages failed: %v", err)
	}

	// Check that file was generated
	outputFile := filepath.Join(outputDir, "main.go")
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("Output file not created: %v", err)
	}
}

func TestExecutor_Execute_NilLogger(t *testing.T) {
	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "ritual")
	outputDir := filepath.Join(tmpDir, "output")

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

	templatePath := filepath.Join(ritualDir, "templates", "main.go")
	os.WriteFile(templatePath, []byte("package main"), 0644)

	vars := generator.NewVariables()

	context := &ExecutionContext{
		RitualPath: ritualDir,
		OutputPath: outputDir,
		Variables:  vars,
		DryRun:     false,
		Logger:     nil, // Test with nil logger
	}

	executor := NewExecutor(context)

	if err := executor.Execute(manifest); err != nil {
		t.Fatalf("Execute failed with nil logger: %v", err)
	}
}

func TestExecutor_Execute_HookFailure(t *testing.T) {
	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "ritual")
	outputDir := filepath.Join(tmpDir, "output")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
		Hooks: ritual.ManifestHooks{
			PreInstall: []string{"exit 1"}, // Command that fails
		},
	}

	vars := generator.NewVariables()

	context := &ExecutionContext{
		RitualPath: ritualDir,
		OutputPath: outputDir,
		Variables:  vars,
		DryRun:     false,
		Logger:     log.New(os.Stdout, "[test] ", 0),
	}

	executor := NewExecutor(context)

	// Should error when hook fails
	if err := executor.Execute(manifest); err == nil {
		t.Error("Expected error for failed hook, got nil")
	}
}

func TestExecutor_Execute_PackageInstallFailure(t *testing.T) {
	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "ritual")
	outputDir := filepath.Join(tmpDir, "output")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
		Dependencies: ritual.Dependencies{
			Packages: []string{"invalid/nonexistent/package@v999.999.999"},
		},
	}

	vars := generator.NewVariables()

	context := &ExecutionContext{
		RitualPath: ritualDir,
		OutputPath: outputDir,
		Variables:  vars,
		DryRun:     false,
		Logger:     log.New(os.Stdout, "[test] ", 0),
	}

	executor := NewExecutor(context)

	// Should handle package install failure
	if err := executor.Execute(manifest); err == nil {
		t.Error("Expected error for failed package install, got nil")
	}
}

package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestGenerateFile(t *testing.T) {
	gen := NewFileGenerator("go-template")

	vars := NewVariables()
	vars.Set("app_name", "test-app")
	vars.Set("port", 8080)
	gen.SetVariables(vars)

	tmpDir := t.TempDir()

	// Create a template file
	templatePath := filepath.Join(tmpDir, "src.tmpl")
	templateContent := "App: [[ .app_name ]]\nPort: [[ .port ]]"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0600); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Generate file
	destPath := filepath.Join(tmpDir, "output.txt")
	if err := gen.GenerateFile(templatePath, destPath, true); err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}

	// Check output
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	expected := "App: test-app\nPort: 8080"
	if string(content) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(content))
	}
}

func TestGenerateFileStatic(t *testing.T) {
	gen := NewFileGenerator("go-template")

	tmpDir := t.TempDir()

	// Create a static file
	staticPath := filepath.Join(tmpDir, "static.txt")
	staticContent := "This is static content"
	if err := os.WriteFile(staticPath, []byte(staticContent), 0600); err != nil {
		t.Fatalf("Failed to create static file: %v", err)
	}

	// Copy file
	destPath := filepath.Join(tmpDir, "output.txt")
	if err := gen.GenerateFile(staticPath, destPath, false); err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}

	// Check output
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	if string(content) != staticContent {
		t.Errorf("Expected '%s', got '%s'", staticContent, string(content))
	}
}

func TestProtectedFiles(t *testing.T) {
	gen := NewFileGenerator("go-template")

	vars := NewVariables()
	vars.Set("app_name", "test-app")
	gen.SetVariables(vars)

	tmpDir := t.TempDir()

	// Create existing protected file
	protectedPath := filepath.Join(tmpDir, "config.yaml")
	originalContent := "original content"
	if err := os.WriteFile(protectedPath, []byte(originalContent), 0600); err != nil {
		t.Fatalf("Failed to create protected file: %v", err)
	}

	// Mark as protected
	gen.SetProtectedFiles([]string{"config.yaml"})

	// Create template
	templatePath := filepath.Join(tmpDir, "config.tmpl")
	templateContent := "app: [[ .app_name ]]"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0600); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Try to generate (should skip)
	if err := gen.GenerateFile(templatePath, protectedPath, true); err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}

	// Check that original content is preserved
	content, err := os.ReadFile(protectedPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != originalContent {
		t.Error("Protected file was overwritten")
	}
}

func TestCreateDirectoryStructure(t *testing.T) {
	gen := NewFileGenerator("go-template")

	tmpDir := t.TempDir()

	dirs := []string{
		"cmd/app",
		"internal/handler",
		"pkg/util",
		"config",
	}

	if err := gen.CreateDirectoryStructure(tmpDir, dirs); err != nil {
		t.Fatalf("CreateDirectoryStructure failed: %v", err)
	}

	// Check that directories were created
	for _, dir := range dirs {
		fullPath := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Directory not created: %s", dir)
		}
	}
}

func TestGenerateFiles(t *testing.T) {
	gen := NewFileGenerator("go-template")

	vars := NewVariables()
	vars.Set("app_name", "my-app")
	gen.SetVariables(vars)

	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "ritual")
	outputDir := filepath.Join(tmpDir, "output")

	// Create ritual structure
	os.MkdirAll(filepath.Join(ritualDir, "templates"), 0750)
	os.MkdirAll(filepath.Join(ritualDir, "static"), 0750)

	// Create template file
	templatePath := filepath.Join(ritualDir, "templates", "main.go.tmpl")
	templateContent := "package main\n\nconst AppName = \"[[ .app_name ]]\""
	os.WriteFile(templatePath, []byte(templateContent), 0600)

	// Create static file
	staticPath := filepath.Join(ritualDir, "static", "README.md")
	staticContent := "# My App"
	os.WriteFile(staticPath, []byte(staticContent), 0600)

	// Create manifest
	manifest := &ritual.Manifest{
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{Source: "templates/main.go.tmpl", Destination: "main.go"},
			},
			Static: []ritual.FileMapping{
				{Source: "static/README.md", Destination: "README.md"},
			},
			Protected: []string{"README.md"},
		},
	}

	// Generate files
	if err := gen.GenerateFiles(manifest, ritualDir, outputDir); err != nil {
		t.Fatalf("GenerateFiles failed: %v", err)
	}

	// Check template output
	mainContent, err := os.ReadFile(filepath.Join(outputDir, "main.go"))
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	expectedMain := "package main\n\nconst AppName = \"my-app\""
	if string(mainContent) != expectedMain {
		t.Errorf("Expected '%s', got '%s'", expectedMain, string(mainContent))
	}

	// Check static output
	readmeContent, err := os.ReadFile(filepath.Join(outputDir, "README.md"))
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	if string(readmeContent) != staticContent {
		t.Errorf("Expected '%s', got '%s'", staticContent, string(readmeContent))
	}
}

func TestGenerateFile_MissingSource(t *testing.T) {
	gen := NewFileGenerator("go-template")
	tmpDir := t.TempDir()

	destPath := filepath.Join(tmpDir, "output.txt")
	err := gen.GenerateFile("/nonexistent/file.txt", destPath, false)
	if err == nil {
		t.Error("Expected error for missing source file, got nil")
	}
}

func TestGenerateFile_InvalidTemplate(t *testing.T) {
	gen := NewFileGenerator("go-template")
	vars := NewVariables()
	vars.Set("app_name", "test")
	gen.SetVariables(vars)

	tmpDir := t.TempDir()

	// Create invalid template
	templatePath := filepath.Join(tmpDir, "bad.tmpl")
	templateContent := "[[ .app_name" // Missing closing ]]
	if err := os.WriteFile(templatePath, []byte(templateContent), 0600); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	destPath := filepath.Join(tmpDir, "output.txt")
	err := gen.GenerateFile(templatePath, destPath, true)
	if err == nil {
		t.Error("Expected error for invalid template, got nil")
	}
}

func TestGenerateFiles_MissingSourceFile(t *testing.T) {
	gen := NewFileGenerator("go-template")

	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "ritual")
	outputDir := filepath.Join(tmpDir, "output")

	os.MkdirAll(ritualDir, 0750)

	manifest := &ritual.Manifest{
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{Source: "nonexistent.go", Destination: "main.go"},
			},
		},
	}

	err := gen.GenerateFiles(manifest, ritualDir, outputDir)
	if err == nil {
		t.Error("Expected error for missing source file, got nil")
	}
}

func TestGenerateFiles_OptionalMissingFile(t *testing.T) {
	gen := NewFileGenerator("go-template")

	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "ritual")
	outputDir := filepath.Join(tmpDir, "output")

	os.MkdirAll(ritualDir, 0750)

	manifest := &ritual.Manifest{
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{Source: "nonexistent.go", Destination: "main.go", Optional: true},
			},
		},
	}

	// Should not error for optional missing file
	err := gen.GenerateFiles(manifest, ritualDir, outputDir)
	if err != nil {
		t.Errorf("Unexpected error for optional missing file: %v", err)
	}
}

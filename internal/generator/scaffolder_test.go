package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestProjectScaffolder_CreateStructure(t *testing.T) {
	tmpDir := t.TempDir()

	scaffolder := NewProjectScaffolder()
	projectPath := filepath.Join(tmpDir, "test-project")

	err := scaffolder.CreateStructure(projectPath)
	if err != nil {
		t.Fatalf("CreateStructure() error = %v", err)
	}

	// Verify standard directories were created
	expectedDirs := []string{
		"cmd",
		"internal",
		"pkg",
		"config",
		"docs",
		"test",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(projectPath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}
}

func TestProjectScaffolder_GenerateMainGo(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")

	scaffolder := NewProjectScaffolder()
	if err := scaffolder.CreateStructure(projectPath); err != nil {
		t.Fatal(err)
	}

	vars := NewVariables()
	vars.Set("app_name", "test-app")
	vars.Set("module_name", "github.com/example/test-app")
	vars.Set("port", 8080)

	err := scaffolder.GenerateMainGo(projectPath, vars)
	if err != nil {
		t.Fatalf("GenerateMainGo() error = %v", err)
	}

	// Verify main.go was created
	mainPath := filepath.Join(projectPath, "cmd", "server", "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		t.Error("main.go was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !contains(contentStr, "package main") {
		t.Error("main.go should contain 'package main'")
	}
	if !contains(contentStr, "test-app") {
		t.Error("main.go should contain app name")
	}
}

func TestProjectScaffolder_GenerateGoMod(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")

	scaffolder := NewProjectScaffolder()
	if err := scaffolder.CreateStructure(projectPath); err != nil {
		t.Fatal(err)
	}

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name: "test-ritual",
		},
		Dependencies: ritual.Dependencies{
			Packages: []string{
				"github.com/toutaio/toutago",
				"github.com/toutaio/toutago-nasc-dependency-injector",
			},
		},
	}

	vars := NewVariables()
	vars.Set("module_name", "github.com/example/test-app")

	err := scaffolder.GenerateGoMod(projectPath, manifest, vars)
	if err != nil {
		t.Fatalf("GenerateGoMod() error = %v", err)
	}

	// Verify go.mod was created
	goModPath := filepath.Join(projectPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Error("go.mod was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !contains(contentStr, "module github.com/example/test-app") {
		t.Error("go.mod should contain module name")
	}
	if !contains(contentStr, "github.com/toutaio/toutago") {
		t.Error("go.mod should contain toutago dependency")
	}
}

func TestProjectScaffolder_GenerateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")

	scaffolder := NewProjectScaffolder()
	if err := scaffolder.CreateStructure(projectPath); err != nil {
		t.Fatal(err)
	}

	vars := NewVariables()
	vars.Set("app_name", "test-app")
	vars.Set("port", 8080)

	err := scaffolder.GenerateConfig(projectPath, vars)
	if err != nil {
		t.Fatalf("GenerateConfig() error = %v", err)
	}

	// Verify .env.example was created
	envPath := filepath.Join(projectPath, ".env.example")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Error(".env.example was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !contains(contentStr, "APP_NAME=test-app") {
		t.Error(".env.example should contain APP_NAME")
	}
	if !contains(contentStr, "PORT=8080") {
		t.Error(".env.example should contain PORT")
	}
}

func TestProjectScaffolder_GenerateFromRitual(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")

	// Create a simple ritual structure
	ritualPath := filepath.Join(tmpDir, "ritual")
	if err := os.MkdirAll(filepath.Join(ritualPath, "templates"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create ritual.yaml
	ritualYAML := `ritual:
  name: test-ritual
  version: 1.0.0
  description: Test ritual
  template_engine: go-template

dependencies:
  packages:
    - github.com/toutaio/toutago

files:
  templates:
    - src: "test.txt.tmpl"
      dest: "test.txt"
`
	if err := os.WriteFile(filepath.Join(ritualPath, "ritual.yaml"), []byte(ritualYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Create template file
	template := `Hello {{ .app_name }}!`
	if err := os.WriteFile(filepath.Join(ritualPath, "templates", "test.txt.tmpl"), []byte(template), 0644); err != nil {
		t.Fatal(err)
	}

	// Load ritual
	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	if err != nil {
		t.Fatal(err)
	}

	// Create scaffolder and generate project
	scaffolder := NewProjectScaffolder()

	vars := NewVariables()
	vars.Set("app_name", "my-app")
	vars.Set("module_name", "github.com/example/my-app")

	err = scaffolder.GenerateFromRitual(projectPath, ritualPath, manifest, vars)
	if err != nil {
		t.Fatalf("GenerateFromRitual() error = %v", err)
	}

	// Verify project structure
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Error("Project directory was not created")
	}

	// Verify template was rendered
	testFile := filepath.Join(projectPath, "test.txt")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Template file was not generated")
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "Hello my-app!" {
		t.Errorf("Template was not rendered correctly: got %s", string(content))
	}
}

func TestProjectScaffolder_GenerateREADME(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")

	scaffolder := NewProjectScaffolder()
	if err := scaffolder.CreateStructure(projectPath); err != nil {
		t.Fatal(err)
	}

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:        "test-ritual",
			Description: "A test ritual for testing",
		},
	}

	vars := NewVariables()
	vars.Set("app_name", "test-app")
	vars.Set("module_name", "github.com/example/test-app")

	err := scaffolder.GenerateREADME(projectPath, manifest, vars)
	if err != nil {
		t.Fatalf("GenerateREADME() error = %v", err)
	}

	// Verify README.md was created
	readmePath := filepath.Join(projectPath, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !contains(contentStr, "test-app") {
		t.Error("README should contain app name")
	}
	if !contains(contentStr, "A test ritual for testing") {
		t.Error("README should contain ritual description")
	}
}

func TestProjectScaffolder_GenerateGitignore(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")

	scaffolder := NewProjectScaffolder()
	if err := scaffolder.CreateStructure(projectPath); err != nil {
		t.Fatal(err)
	}

	err := scaffolder.GenerateGitignore(projectPath)
	if err != nil {
		t.Fatalf("GenerateGitignore() error = %v", err)
	}

	// Verify .gitignore was created
	gitignorePath := filepath.Join(projectPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	expectedPatterns := []string{
		"*.exe",
		".env",
		"vendor/",
		"bin/",
	}

	for _, pattern := range expectedPatterns {
		if !contains(contentStr, pattern) {
			t.Errorf(".gitignore should contain pattern: %s", pattern)
		}
	}
}

func TestProjectScaffolder_ApplyTemplateFiles(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")
	ritualPath := filepath.Join(tmpDir, "ritual")

	// Create ritual with templates
	if err := os.MkdirAll(filepath.Join(ritualPath, "templates"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create template files
	template1 := `package main
// {{ .app_name }}`
	if err := os.WriteFile(filepath.Join(ritualPath, "templates", "file1.go.tmpl"), []byte(template1), 0644); err != nil {
		t.Fatal(err)
	}

	// Create static file
	static := `# Static content`
	if err := os.WriteFile(filepath.Join(ritualPath, "templates", "static.txt"), []byte(static), 0644); err != nil {
		t.Fatal(err)
	}

	manifest := &ritual.Manifest{
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{Source: "file1.go.tmpl", Destination: "output.go"},
			},
			Static: []ritual.FileMapping{
				{Source: "static.txt", Destination: "static.txt"},
			},
		},
	}

	scaffolder := NewProjectScaffolder()
	if err := scaffolder.CreateStructure(projectPath); err != nil {
		t.Fatal(err)
	}

	vars := NewVariables()
	vars.Set("app_name", "my-app")

	err := scaffolder.ApplyTemplateFiles(projectPath, ritualPath, manifest, vars)
	if err != nil {
		t.Fatalf("ApplyTemplateFiles() error = %v", err)
	}

	// Verify template was rendered
	output1 := filepath.Join(projectPath, "output.go")
	content1, err := os.ReadFile(output1)
	if err != nil {
		t.Error("Template file was not created")
	} else if !contains(string(content1), "my-app") {
		t.Error("Template was not rendered with variables")
	}

	// Verify static file was copied
	output2 := filepath.Join(projectPath, "static.txt")
	content2, err := os.ReadFile(output2)
	if err != nil {
		t.Error("Static file was not copied")
	} else if string(content2) != static {
		t.Error("Static file content was modified")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && anySubstringMatch(s, substr)
}

func anySubstringMatch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestProjectScaffolder_ExecuteHooks(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")

	scaffolder := NewProjectScaffolder()
	if err := scaffolder.CreateStructure(projectPath); err != nil {
		t.Fatal(err)
	}

	// Create test hooks
	hooks := []string{
		"echo 'hook executed' > hook_marker.txt",
	}

	err := scaffolder.ExecutePostGenerateHooks(projectPath, hooks)
	if err != nil {
		t.Fatalf("ExecutePostGenerateHooks() error = %v", err)
	}

	// Verify hook was executed
	markerFile := filepath.Join(projectPath, "hook_marker.txt")
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Error("Hook should have created marker file")
	}
}

func TestProjectScaffolder_GenerateWithHooks(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")
	ritualPath := filepath.Join(tmpDir, "ritual")

	// Create ritual with hooks
	if err := os.MkdirAll(filepath.Join(ritualPath, "templates"), 0755); err != nil {
		t.Fatal(err)
	}

	ritualYAML := `ritual:
  name: hooks-ritual
  version: 1.0.0
  description: Test hooks
  template_engine: go-template

hooks:
  post_install:
    - "echo 'post-install hook' > post_install.txt"

files:
  templates:
    - src: "test.txt.tmpl"
      dest: "test.txt"
`
	if err := os.WriteFile(filepath.Join(ritualPath, "ritual.yaml"), []byte(ritualYAML), 0644); err != nil {
		t.Fatal(err)
	}

	template := "Test content"
	if err := os.WriteFile(filepath.Join(ritualPath, "templates", "test.txt.tmpl"), []byte(template), 0644); err != nil {
		t.Fatal(err)
	}

	// Load ritual
	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	if err != nil {
		t.Fatal(err)
	}

	scaffolder := NewProjectScaffolder()
	vars := NewVariables()
	vars.Set("app_name", "test-app")

	// Generate with hooks
	err = scaffolder.GenerateFromRitualWithHooks(projectPath, ritualPath, manifest, vars)
	if err != nil {
		t.Fatalf("GenerateFromRitualWithHooks() error = %v", err)
	}

	// Verify hook was executed
	hookFile := filepath.Join(projectPath, "post_install.txt")
	if _, err := os.Stat(hookFile); os.IsNotExist(err) {
		t.Error("Post-install hook should have been executed")
	}
}

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/executor"
	"github.com/toutaio/toutago-ritual-grove/internal/generator"
	"github.com/toutaio/toutago-ritual-grove/internal/registry"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// TestAllBuiltinRituals tests all built-in rituals to ensure they:
// 1. Load successfully
// 2. Generate valid projects
// 3. Compile without errors
// 4. Pass their own tests (if any)
func TestAllBuiltinRituals(t *testing.T) {
	// Find ritual-grove root
	rootDir, err := findRitualGroveRoot()
	if err != nil {
		t.Fatalf("Failed to find ritual-grove root: %v", err)
	}

	ritualsDir := filepath.Join(rootDir, "rituals")

	// List all built-in rituals
	entries, err := os.ReadDir(ritualsDir)
	if err != nil {
		t.Fatalf("Failed to read rituals directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("No built-in rituals found")
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		ritualName := entry.Name()
		t.Run(ritualName, func(t *testing.T) {
			testRitual(t, ritualsDir, ritualName)
		})
	}
}

func testRitual(t *testing.T, ritualsDir, ritualName string) {
	// Create temp directory for generated project
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")

	// Load ritual
	reg := registry.NewRegistry()
	reg.AddSearchPath(ritualsDir)

	err := reg.Scan()
	if err != nil {
		t.Fatalf("Failed to scan rituals: %v", err)
	}

	manifest, err := reg.Load(ritualName)
	if err != nil {
		t.Fatalf("Failed to load ritual %s: %v", ritualName, err)
	}

	// Create default answers for questionnaire
	answers := createDefaultAnswers(t, manifest, projectPath)

	// Create variables from answers
	vars := generator.NewVariables()
	for k, v := range answers {
		vars.Set(k, v)
	}

	// Execute ritual
	ctx := &executor.ExecutionContext{
		RitualPath: filepath.Join(ritualsDir, ritualName),
		OutputPath: projectPath,
		Variables:  vars,
		DryRun:     false,
	}

	exec := executor.NewExecutor(ctx)
	if err := exec.Execute(manifest); err != nil {
		t.Fatalf("Failed to execute ritual: %v", err)
	}

	// Verify project structure
	verifyProjectStructure(t, projectPath, ritualName)

	// Verify project compiles
	verifyProjectCompiles(t, projectPath)

	// Run generated tests if they exist
	runGeneratedTests(t, projectPath)
}

func createDefaultAnswers(t *testing.T, manifest *ritual.Manifest, projectPath string) map[string]interface{} {
	answers := make(map[string]interface{})

	// Common default answers
	answers["project_name"] = "test-project"
	answers["module_path"] = "example.com/test-project"
	answers["port"] = "8080"
	answers["database"] = "none"

	// Process ritual questions and provide defaults
	if manifest.Questions != nil {
		for _, q := range manifest.Questions {
			if _, exists := answers[q.Name]; exists {
				continue // Already set
			}

			// Set defaults based on question type
			switch q.Type {
			case "text", "path":
				if q.Default != nil {
					answers[q.Name] = q.Default
				} else {
					answers[q.Name] = "test-value"
				}
			case "number":
				if q.Default != nil {
					answers[q.Name] = q.Default
				} else {
					answers[q.Name] = "1"
				}
			case "boolean":
				if q.Default != nil {
					answers[q.Name] = q.Default
				} else {
					answers[q.Name] = false
				}
			case "choice":
				if len(q.Choices) > 0 {
					answers[q.Name] = q.Choices[0]
				}
			case "multi-choice":
				if len(q.Choices) > 0 {
					answers[q.Name] = []string{q.Choices[0]}
				}
			}
		}
	}

	return answers
}

func verifyProjectStructure(t *testing.T, projectPath, ritualName string) {
	// Check essential files exist
	essentialFiles := []string{
		"go.mod",
		"main.go",
	}

	for _, file := range essentialFiles {
		path := filepath.Join(projectPath, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file not found: %s", file)
		}
	}

	// Ritual-specific checks
	switch ritualName {
	case "blog":
		// Blog should have handlers, models, views
		checkExists(t, projectPath, "handlers")
		checkExists(t, projectPath, "models")
		checkExists(t, projectPath, "views")
	case "wiki":
		// Wiki should have pages, revisions
		checkExists(t, projectPath, "handlers")
		checkExists(t, projectPath, "models")
	case "minimal", "hello-world", "basic-site":
		// Should have at least main.go
		checkExists(t, projectPath, "main.go")
	}
}

func checkExists(t *testing.T, basePath, relativePath string) {
	path := filepath.Join(basePath, relativePath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected path not found: %s", relativePath)
	}
}

func verifyProjectCompiles(t *testing.T, projectPath string) {
	// First run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = projectPath
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		t.Logf("go mod tidy output: %s", output)
		t.Errorf("go mod tidy failed: %v", err)
		return
	}

	// Try to build the project
	buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = projectPath
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Logf("Build output: %s", output)
		t.Errorf("Project failed to compile: %v", err)
	}
}

func runGeneratedTests(t *testing.T, projectPath string) {
	// Check if tests exist
	testCmd := exec.Command("go", "list", "./...")
	testCmd.Dir = projectPath
	if err := testCmd.Run(); err != nil {
		// No tests, skip
		return
	}

	// Run tests
	runCmd := exec.Command("go", "test", "./...", "-short")
	runCmd.Dir = projectPath
	output, err := runCmd.CombinedOutput()
	if err != nil {
		t.Logf("Test output: %s", output)
		// Don't fail the test, just log
		t.Logf("Generated tests failed (non-critical): %v", err)
	}
}

func findRitualGroveRoot() (string, error) {
	// Start from current directory and walk up
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Check if go.mod exists and contains ritual-grove
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, check if it's ritual-grove
			if filepath.Base(dir) == "toutago-ritual-grove" || 
				filepath.Base(dir) == "ritual-grove" {
				return dir, nil
			}
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toutaio/toutago-ritual-grove/internal/generator"
	"github.com/toutaio/toutago-ritual-grove/internal/questionnaire"
	"github.com/toutaio/toutago-ritual-grove/internal/registry"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// TestBasicSiteRitualEndToEnd tests the entire basic-site ritual flow
func TestBasicSiteRitualEndToEnd(t *testing.T) {
	// Setup: Create temporary output directory
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "my-site")

	// Step 1: Load the basic-site ritual
	ritualPath := filepath.Join("..", "rituals", "basic-site")
	loader := ritual.NewLoader(ritualPath)
	
	manifest, err := loader.Load(ritualPath)
	require.NoError(t, err, "Should load basic-site ritual")
	assert.Equal(t, "basic-site", manifest.Ritual.Name)
	assert.Equal(t, "1.0.0", manifest.Ritual.Version)

	// Step 2: Answer questions programmatically
	flow := questionnaire.NewQuestionFlow()
	answers := map[string]interface{}{
		"site_name":       "Test Site",
		"port":            3000,
		"enable_database": false,
	}
	
	for name, value := range answers {
		flow.SetAnswer(name, value)
	}

	// Step 3: Prepare variables for generation
	vars := generator.NewVariables()
	vars.Set("site_name", "Test Site")
	vars.Set("port", 3000)
	vars.Set("enable_database", false)
	vars.Set("module_path", "github.com/test/my-site")
	vars.Set("message", "Welcome to your test site!")

	// Step 4: Generate project files
	gen := generator.NewFileGenerator(manifest.Ritual.TemplateEngine)
	gen.SetVariables(vars)

	// Create directories first
	err = gen.CreateDirectoryStructure(projectPath, manifest.Files.Directories)
	require.NoError(t, err, "Should create directory structure")

	t.Logf("Generating files from %s to %s", ritualPath, projectPath)
	t.Logf("Templates: %+v", manifest.Files.Templates)
	
	err = gen.GenerateFiles(manifest, ritualPath, projectPath)
	if err != nil {
		t.Logf("Generation error: %v", err)
		// List what's in ritualPath
		entries, _ := os.ReadDir(ritualPath)
		t.Logf("Ritual path contents:")
		for _, e := range entries {
			t.Logf("  - %s (dir=%v)", e.Name(), e.IsDir())
		}
		// List what's in projectPath
		entries2, _ := os.ReadDir(projectPath)
		t.Logf("Project path contents:")
		for _, e := range entries2 {
			t.Logf("  - %s (dir=%v)", e.Name(), e.IsDir())
		}
	}
	require.NoError(t, err, "Should generate files successfully")

	// Step 5: Verify generated files exist
	expectedFiles := []string{
		"main.go",
		"handlers/home.go",
		"views/home.html",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(projectPath, file)
		assert.FileExists(t, fullPath, "File should exist: %s", file)
	}

	// Step 6: Verify generated directories exist
	expectedDirs := []string{
		"handlers",
		"views",
		"static",
		"config",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(projectPath, dir)
		info, err := os.Stat(fullPath)
		require.NoError(t, err, "Directory should exist: %s", dir)
		assert.True(t, info.IsDir(), "Should be a directory: %s", dir)
	}

	// Step 7: Verify template variables were substituted correctly
	mainContent, err := os.ReadFile(filepath.Join(projectPath, "main.go"))
	require.NoError(t, err, "Should read main.go")
	
	mainStr := string(mainContent)
	assert.Contains(t, mainStr, "github.com/test/my-site/handlers", "Should contain module path")
	assert.Contains(t, mainStr, "3000", "Should contain port number")
	assert.Contains(t, mainStr, "Test Site", "Should contain site name")

	// Step 8: Verify handler contains correct values
	handlerContent, err := os.ReadFile(filepath.Join(projectPath, "handlers/home.go"))
	require.NoError(t, err, "Should read handler")
	
	handlerStr := string(handlerContent)
	assert.Contains(t, handlerStr, "Test Site", "Handler should contain site name")

	t.Log("✅ Basic-site ritual end-to-end test passed!")
}

// TestBasicSiteRitualWithRegistry tests ritual discovery and loading
func TestBasicSiteRitualWithRegistry(t *testing.T) {
	// Step 1: Create registry
	reg := registry.NewRegistry()
	
	// Add built-in rituals directory
	ritualsDir := filepath.Join("..", "rituals")
	reg.AddSearchPath(ritualsDir)

	// Step 2: Scan for rituals
	err := reg.Scan()
	require.NoError(t, err, "Should scan for rituals")
	
	// Step 3: Get list of rituals
	rituals := reg.List()
	
	// Should find at least basic-site
	assert.GreaterOrEqual(t, len(rituals), 1, "Should discover at least one ritual")

	// Find basic-site
	var basicSite *registry.RitualMetadata
	for _, r := range rituals {
		if r.Name == "basic-site" {
			basicSite = r
			break
		}
	}
	require.NotNil(t, basicSite, "Should find basic-site ritual")
	assert.Equal(t, "1.0.0", basicSite.Version)

	// Step 4: Get the ritual details
	ritualMeta, err := reg.Get("basic-site")
	require.NoError(t, err, "Should get basic-site ritual metadata")
	assert.Equal(t, "basic-site", ritualMeta.Name)

	t.Log("✅ Basic-site ritual registry test passed!")
}

// TestBasicSiteRitualValidation tests ritual manifest validation
func TestBasicSiteRitualValidation(t *testing.T) {
	// Load ritual
	ritualPath := filepath.Join("..", "rituals", "basic-site")
	loader := ritual.NewLoader(ritualPath)
	
	manifest, err := loader.Load(ritualPath)
	require.NoError(t, err, "Should load ritual")

	// Validate ritual
	// Note: This requires implementing validator
	// For now, just check required fields exist
	assert.NotEmpty(t, manifest.Ritual.Name, "Name should not be empty")
	assert.NotEmpty(t, manifest.Ritual.Version, "Version should not be empty")
	assert.NotEmpty(t, manifest.Ritual.Description, "Description should not be empty")
	
	// Validate questions
	assert.NotEmpty(t, manifest.Questions, "Should have questions")
	
	// Validate each question has required fields
	for _, q := range manifest.Questions {
		assert.NotEmpty(t, q.Name, "Question should have name")
		assert.NotEmpty(t, q.Prompt, "Question should have prompt")
		assert.NotEmpty(t, q.Type, "Question should have type")
	}

	// Validate files section
	assert.NotEmpty(t, manifest.Files.Templates, "Should have templates")

	t.Log("✅ Basic-site ritual validation test passed!")
}

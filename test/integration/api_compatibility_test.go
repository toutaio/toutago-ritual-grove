package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toutaio/toutago-ritual-grove/internal/generator"
	"github.com/toutaio/toutago-ritual-grove/internal/questionnaire"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// TestBasicSiteAPICompatibility verifies that the generated basic-site code
// uses the correct Cosan and Fith APIs
func TestBasicSiteAPICompatibility(t *testing.T) {
	tmpDir := t.TempDir()

	// Load basic-site ritual
	ritualPath := filepath.Join("..", "..", "rituals", "basic-site", "ritual.yaml")
	r, err := ritual.LoadFromFile(ritualPath)
	require.NoError(t, err)

	// Prepare answers
	answers := map[string]interface{}{
		"site_name":   "Test Site",
		"module_path": "example.com/testsite",
		"port":        8080,
	}

	// Validate answers
	qa := questionnaire.NewEngine()
	validAnswers, err := qa.ValidateAnswers(r.Questions, answers)
	require.NoError(t, err)

	// Generate files
	gen := generator.NewEngine()
	err = gen.Generate(tmpDir, r, validAnswers)
	require.NoError(t, err)

	// Test 1: Verify main.go uses cosan.New() and cosan.Context
	mainPath := filepath.Join(tmpDir, "main.go")
	mainContent, err := os.ReadFile(mainPath)
	require.NoError(t, err)
	mainStr := string(mainContent)

	assert.Contains(t, mainStr, `router := cosan.New()`, "Should create router with cosan.New()")
	assert.Contains(t, mainStr, `router.GET("/", homeHandler.Index)`, "Should register GET route")
	assert.Contains(t, mainStr, `router.Listen(`, "Should use router.Listen() method")

	// Test 2: Verify Fith configuration
	assert.Contains(t, mainStr, `fith.New(fith.Config{`, "Should use fith.Config struct")
	assert.Contains(t, mainStr, `TemplateDir:`, "Should configure template directory")

	// Test 3: Verify handler uses cosan.Context
	handlerPath := filepath.Join(tmpDir, "handlers", "home.go")
	handlerContent, err := os.ReadFile(handlerPath)
	require.NoError(t, err)
	handlerStr := string(handlerContent)

	assert.Contains(t, handlerStr, `func (h *HomeHandler) Index(ctx cosan.Context) error`, "Handler should accept cosan.Context")
	assert.Contains(t, handlerStr, `ctx.Render(`, "Should use ctx.Render() method")
	assert.NotContains(t, handlerStr, `http.ResponseWriter`, "Should not use http.ResponseWriter directly")
	assert.NotContains(t, handlerStr, `*http.Request`, "Should not use *http.Request directly")

	// Test 4: Verify Fith template syntax
	viewPath := filepath.Join(tmpDir, "views", "home.html")
	viewContent, err := os.ReadFile(viewPath)
	require.NoError(t, err)
	viewStr := string(viewContent)

	assert.Contains(t, viewStr, `{{.site_name}}`, "Should use Fith template syntax")
	assert.Contains(t, viewStr, `{{.message}}`, "Should use Fith variable syntax")

	// Test 5: Verify go.mod has correct dependencies
	gomodPath := filepath.Join(tmpDir, "go.mod")
	gomodContent, err := os.ReadFile(gomodPath)
	require.NoError(t, err)
	gomodStr := string(gomodContent)

	assert.Contains(t, gomodStr, `github.com/toutaio/toutago-cosan-router`, "Should depend on Cosan")
	assert.Contains(t, gomodStr, `github.com/toutaio/toutago-fith-renderer`, "Should depend on Fith")
}

// TestBasicSiteStructure verifies the generated project structure
func TestBasicSiteStructure(t *testing.T) {
	tmpDir := t.TempDir()

	// Load and generate
	ritualPath := filepath.Join("..", "..", "rituals", "basic-site", "ritual.yaml")
	r, err := ritual.LoadFromFile(ritualPath)
	require.NoError(t, err)

	answers := map[string]interface{}{
		"site_name":   "Test Site",
		"module_path": "example.com/testsite",
		"port":        8080,
	}

	qa := questionnaire.NewEngine()
	validAnswers, err := qa.ValidateAnswers(r.Questions, answers)
	require.NoError(t, err)

	gen := generator.NewEngine()
	err = gen.Generate(tmpDir, r, validAnswers)
	require.NoError(t, err)

	// Verify expected files exist
	expectedFiles := []string{
		"main.go",
		"go.mod",
		"README.md",
		"handlers/home.go",
		"views/home.html",
		"static/style.css",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		_, err := os.Stat(path)
		assert.NoError(t, err, "Expected file %s should exist", file)
	}
}

// TestRitualAnswerSubstitution verifies template variable substitution
func TestRitualAnswerSubstitution(t *testing.T) {
	tmpDir := t.TempDir()

	ritualPath := filepath.Join("..", "..", "rituals", "basic-site", "ritual.yaml")
	r, err := ritual.LoadFromFile(ritualPath)
	require.NoError(t, err)

	answers := map[string]interface{}{
		"site_name":   "My Awesome Site",
		"module_path": "github.com/myorg/mysite",
		"port":        3000,
	}

	qa := questionnaire.NewEngine()
	validAnswers, err := qa.ValidateAnswers(r.Questions, answers)
	require.NoError(t, err)

	gen := generator.NewEngine()
	err = gen.Generate(tmpDir, r, validAnswers)
	require.NoError(t, err)

	// Verify substitutions in main.go
	mainContent, err := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	require.NoError(t, err)
	mainStr := string(mainContent)

	assert.Contains(t, mainStr, `"github.com/myorg/mysite/handlers"`, "Should substitute module_path")
	assert.Contains(t, mainStr, `:= 3000`, "Should substitute port")
	assert.Contains(t, mainStr, `My Awesome Site`, "Should substitute site_name")

	// Verify substitutions in go.mod
	gomodContent, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	require.NoError(t, err)
	gomodStr := string(gomodContent)

	assert.Contains(t, gomodStr, `module github.com/myorg/mysite`, "Should substitute module in go.mod")
}

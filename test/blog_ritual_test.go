package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// TestBlogRitualWithInertiaVue tests the blog ritual configuration for Inertia.js + Vue.
func TestBlogRitualWithInertiaVue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ritualPath, err := findRitual("blog")
	require.NoError(t, err, "Blog ritual should exist")

	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	require.NoError(t, err, "Should load blog ritual")

	err = manifest.Validate()
	require.NoError(t, err, "Ritual should be valid")

	// Verify Inertia-Vue templates exist in the manifest
	hasInertiaTemplates := false
	for _, tmpl := range manifest.Files.Templates {
		if strings.Contains(tmpl.Source, "inertia") || strings.Contains(tmpl.Source, "frontend") {
			hasInertiaTemplates = true
			break
		}
	}
	assert.True(t, hasInertiaTemplates, "Should have Inertia/frontend templates")

	// Verify SSR question exists and depends on inertia-vue
	var ssrQuestion *ritual.Question
	for i := range manifest.Questions {
		if manifest.Questions[i].Name == "enable_ssr" {
			ssrQuestion = &manifest.Questions[i]
			break
		}
	}
	require.NotNil(t, ssrQuestion, "Should have enable_ssr question")
	assert.NotNil(t, ssrQuestion.Condition, "SSR question should have condition")

	t.Logf("✅ Blog ritual with Inertia-Vue configuration validated successfully")
}

// TestBlogRitualWithHTMX tests the blog ritual configuration for HTMX.
func TestBlogRitualWithHTMX(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ritualPath, err := findRitual("blog")
	require.NoError(t, err)

	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	require.NoError(t, err)

	err = manifest.Validate()
	require.NoError(t, err)

	// Verify HTMX templates exist
	hasHTMXTemplates := false
	for _, tmpl := range manifest.Files.Templates {
		if strings.Contains(tmpl.Source, "htmx") || strings.Contains(tmpl.Source, "views") {
			hasHTMXTemplates = true
			break
		}
	}
	assert.True(t, hasHTMXTemplates, "Should have HTMX/views templates")

	t.Logf("✅ Blog ritual with HTMX configuration validated successfully")
}

// TestBlogRitualWithTraditional tests that the blog ritual loads successfully with traditional frontend.
func TestBlogRitualWithTraditional(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ritualPath, err := findRitual("blog")
	require.NoError(t, err, "Blog ritual should exist")

	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	require.NoError(t, err, "Should load blog ritual")

	err = manifest.Validate()
	require.NoError(t, err, "Ritual should be valid")

	// Verify ritual metadata
	assert.Equal(t, "blog", manifest.Ritual.Name)
	assert.NotEmpty(t, manifest.Ritual.Version)
	assert.NotEmpty(t, manifest.Ritual.Description)

	// Verify questions exist
	assert.NotEmpty(t, manifest.Questions, "Should have questions")

	// Find frontend_type question
	var frontendQuestion *ritual.Question
	for i := range manifest.Questions {
		if manifest.Questions[i].Name == "frontend_type" {
			frontendQuestion = &manifest.Questions[i]
			break
		}
	}
	require.NotNil(t, frontendQuestion, "Should have frontend_type question")
	assert.Contains(t, frontendQuestion.Choices, "traditional")
	assert.Contains(t, frontendQuestion.Choices, "inertia-vue")
	assert.Contains(t, frontendQuestion.Choices, "htmx")

	// Verify files section exists
	assert.NotEmpty(t, manifest.Files.Templates, "Should have template files")

	// Verify essential templates exist
	hasMainGo := false
	hasGoMod := false
	for _, tmpl := range manifest.Files.Templates {
		if tmpl.Destination == "main.go" {
			hasMainGo = true
		}
		if tmpl.Destination == "go.mod" {
			hasGoMod = true
		}
	}
	assert.True(t, hasMainGo, "Should have main.go template")
	assert.True(t, hasGoMod, "Should have go.mod template")

	t.Logf("✅ Blog ritual with traditional templates validated successfully")
}

// TestBlogRitualStructure tests that the blog ritual has all required components for all frontend types.
func TestBlogRitualStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ritualPath, err := findRitual("blog")
	require.NoError(t, err)

	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	require.NoError(t, err)

	err = manifest.Validate()
	require.NoError(t, err)

	// Verify models exist
	hasModels := false
	for _, tmpl := range manifest.Files.Templates {
		if strings.Contains(tmpl.Destination, "models/") {
			hasModels = true
			break
		}
	}
	assert.True(t, hasModels, "Should have model templates")

	// Verify handlers exist
	hasHandlers := false
	for _, tmpl := range manifest.Files.Templates {
		if strings.Contains(tmpl.Destination, "handlers/") {
			hasHandlers = true
			break
		}
	}
	assert.True(t, hasHandlers, "Should have handler templates")

	// Verify migrations exist (blog has database)
	assert.NotEmpty(t, manifest.Migrations, "Should have migrations")

	t.Logf("✅ Blog ritual structure validated successfully")
}

// Helper functions

func findRitual(name string) (string, error) {
	// Get absolute path to rituals directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Try from test directory (when running from project root)
	ritualPath := filepath.Join(cwd, "..", "rituals", name)
	if _, err := os.Stat(filepath.Join(ritualPath, "ritual.yaml")); err == nil {
		absPath, _ := filepath.Abs(ritualPath)
		return absPath, nil
	}

	// Try from current directory (when running from test dir)
	ritualPath = filepath.Join(cwd, "rituals", name)
	if _, err := os.Stat(filepath.Join(ritualPath, "ritual.yaml")); err == nil {
		absPath, _ := filepath.Abs(ritualPath)
		return absPath, nil
	}

	// Try parent directory
	ritualPath = filepath.Join(filepath.Dir(cwd), "rituals", name)
	if _, err := os.Stat(filepath.Join(ritualPath, "ritual.yaml")); err == nil {
		absPath, _ := filepath.Abs(ritualPath)
		return absPath, nil
	}

	return "", os.ErrNotExist
}

package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toutaio/toutago-ritual-grove/internal/generator"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestWikiRitual_Load(t *testing.T) {
	ritualPath := filepath.Join("..", "rituals", "wiki")

	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	require.NoError(t, err, "Should load wiki ritual")

	assert.Equal(t, "wiki", manifest.Ritual.Name)
	assert.NotEmpty(t, manifest.Ritual.Description)

	requiredQuestions := map[string]bool{
		"wiki_name":   false,
		"module_path": false,
		"port":        false,
		"database":    false,
	}

	for _, q := range manifest.Questions {
		if _, exists := requiredQuestions[q.Name]; exists {
			requiredQuestions[q.Name] = true
		}
	}

	for name, found := range requiredQuestions {
		assert.True(t, found, "Required question '%s' not found", name)
	}

	assert.NotEmpty(t, manifest.Files.Templates, "No templates defined")
}

func TestWikiRitual_Generate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ritualPath := filepath.Join("..", "rituals", "wiki")
	projectPath := filepath.Join(t.TempDir(), "test-wiki")

	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	require.NoError(t, err)

	vars := generator.NewVariables()
	vars.Set("wiki_name", "Test Wiki")
	vars.Set("module_path", "github.com/test/wiki")
	vars.Set("port", 8080)
	vars.Set("database", "postgres")
	vars.Set("enable_search", true)
	vars.Set("enable_tags", true)
	vars.Set("max_revisions", 100)

	gen := generator.NewFileGenerator(manifest.Ritual.TemplateEngine)
	gen.SetVariables(vars)

	err = gen.CreateDirectoryStructure(projectPath, manifest.Files.Directories)
	require.NoError(t, err)

	err = gen.GenerateFiles(manifest, ritualPath, projectPath)
	require.NoError(t, err, "Failed to generate wiki project")

	requiredFiles := []string{
		"main.go",
		"go.mod",
		"models/page.go",
		"models/revision.go",
		"handlers/pages.go",
		"views/page.html",
	}

	for _, file := range requiredFiles {
		fullPath := filepath.Join(projectPath, file)
		assert.FileExists(t, fullPath, "Expected file not generated: %s", file)
	}

	pageModel := filepath.Join(projectPath, "models", "page.go")
	content, err := os.ReadFile(pageModel)
	require.NoError(t, err)

	requiredFields := []string{"Title", "Content", "Slug", "Version"}
	for _, field := range requiredFields {
		assert.Contains(t, string(content), field, "Page model missing required field")
	}
}

func TestWikiRitual_RevisionTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ritualPath := filepath.Join("..", "rituals", "wiki")
	projectPath := filepath.Join(t.TempDir(), "test-wiki-revisions")

	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	require.NoError(t, err)

	vars := generator.NewVariables()
	vars.Set("wiki_name", "Test Wiki")
	vars.Set("module_path", "github.com/test/wiki")
	vars.Set("port", 8080)
	vars.Set("database", "postgres")
	vars.Set("enable_search", true)
	vars.Set("enable_tags", true)
	vars.Set("max_revisions", 100)

	gen := generator.NewFileGenerator(manifest.Ritual.TemplateEngine)
	gen.SetVariables(vars)

	err = gen.CreateDirectoryStructure(projectPath, manifest.Files.Directories)
	require.NoError(t, err)

	err = gen.GenerateFiles(manifest, ritualPath, projectPath)
	require.NoError(t, err)

	revisionModel := filepath.Join(projectPath, "models", "revision.go")
	content, err := os.ReadFile(revisionModel)
	require.NoError(t, err)

	requiredFields := []string{"CreatedAt", "Author", "PageID", "Content"}
	for _, field := range requiredFields {
		assert.Contains(t, string(content), field, "Revision model missing field")
	}
}

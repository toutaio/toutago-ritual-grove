package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toutaio/toutago-ritual-grove/internal/generator"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// TestSharedTemplateSupport tests loading templates from _shared directory
func TestSharedTemplateSupport(t *testing.T) {
	// Setup test directories
	tmpDir := t.TempDir()
	ritualsDir := filepath.Join(tmpDir, "rituals")
	sharedDir := filepath.Join(ritualsDir, "_shared", "docker")
	testRitualDir := filepath.Join(ritualsDir, "test-ritual", "templates")
	outputDir := filepath.Join(tmpDir, "output")

	// Create directories
	require.NoError(t, os.MkdirAll(sharedDir, 0750))
	require.NoError(t, os.MkdirAll(testRitualDir, 0750))

	// Create a shared template
	sharedTemplate := `FROM golang:[[.go_version]]-alpine
WORKDIR /app
EXPOSE [[.port]]`
	err := os.WriteFile(filepath.Join(sharedDir, "Dockerfile.tmpl"), []byte(sharedTemplate), 0600)
	require.NoError(t, err)

	// Create ritual manifest that references shared template
	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test",
			Version: "1.0.0",
		},
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{
					Source:      "_shared:docker/Dockerfile.tmpl",
					Destination: "Dockerfile",
				},
			},
		},
	}

	// Setup generator
	gen := generator.NewFileGenerator("fith")
	vars := generator.NewVariables()
	vars.Set("go_version", "1.21")
	vars.Set("port", 8080)
	gen.SetVariables(vars)

	// Set rituals base path
	gen.SetRitualsBasePath(ritualsDir)

	// Generate files
	err = gen.GenerateFiles(manifest, testRitualDir, outputDir)
	require.NoError(t, err)

	// Verify Dockerfile was created
	dockerfilePath := filepath.Join(outputDir, "Dockerfile")
	require.FileExists(t, dockerfilePath)

	// Verify content was rendered correctly
	content, err := os.ReadFile(dockerfilePath)
	require.NoError(t, err)
	
	assert.Contains(t, string(content), "FROM golang:1.21-alpine")
	assert.Contains(t, string(content), "EXPOSE 8080")
}

// TestSharedTemplateWithCondition tests conditional shared template loading
func TestSharedTemplateWithCondition(t *testing.T) {
	tmpDir := t.TempDir()
	ritualsDir := filepath.Join(tmpDir, "rituals")
	sharedDir := filepath.Join(ritualsDir, "_shared", "docker")
	testRitualDir := filepath.Join(ritualsDir, "test-ritual", "templates")
	outputDir := filepath.Join(tmpDir, "output")

	require.NoError(t, os.MkdirAll(sharedDir, 0750))
	require.NoError(t, os.MkdirAll(testRitualDir, 0750))

	// Create shared docker-compose template
	composeTemplate := `version: '3.9'
services:
  app:
    build: .`
	err := os.WriteFile(filepath.Join(sharedDir, "docker-compose.yml.tmpl"), []byte(composeTemplate), 0600)
	require.NoError(t, err)

	t.Run("generate when condition is true", func(t *testing.T) {
		manifest := &ritual.Manifest{
			Files: ritual.FilesSection{
				Templates: []ritual.FileMapping{
					{
						Source:      "_shared:docker/docker-compose.yml.tmpl",
						Destination: "docker-compose.yml",
						Condition:   "use_docker == true",
					},
				},
			},
		}

		gen := generator.NewFileGenerator("fith")
		vars := generator.NewVariables()
		vars.Set("use_docker", true)
		gen.SetVariables(vars)
		gen.SetRitualsBasePath(ritualsDir)

		err := gen.GenerateFiles(manifest, testRitualDir, outputDir)
		require.NoError(t, err)

		composePath := filepath.Join(outputDir, "docker-compose.yml")
		assert.FileExists(t, composePath)
	})

	t.Run("skip when condition is false", func(t *testing.T) {
		outputDir2 := filepath.Join(tmpDir, "output2")
		
		manifest := &ritual.Manifest{
			Files: ritual.FilesSection{
				Templates: []ritual.FileMapping{
					{
						Source:      "_shared:docker/docker-compose.yml.tmpl",
						Destination: "docker-compose.yml",
						Condition:   "use_docker == true",
					},
				},
			},
		}

		gen := generator.NewFileGenerator("fith")
		vars := generator.NewVariables()
		vars.Set("use_docker", false)
		gen.SetVariables(vars)
		gen.SetRitualsBasePath(ritualsDir)

		err := gen.GenerateFiles(manifest, testRitualDir, outputDir2)
		require.NoError(t, err)

		composePath := filepath.Join(outputDir2, "docker-compose.yml")
		assert.NoFileExists(t, composePath)
	})
}

// TestSharedTemplateDirectory tests copying entire shared directory
func TestSharedTemplateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	ritualsDir := filepath.Join(tmpDir, "rituals")
	sharedDockerDir := filepath.Join(ritualsDir, "_shared", "docker")
	testRitualDir := filepath.Join(ritualsDir, "test-ritual", "templates")
	outputDir := filepath.Join(tmpDir, "output")

	require.NoError(t, os.MkdirAll(sharedDockerDir, 0750))
	require.NoError(t, os.MkdirAll(testRitualDir, 0750))

	// Create multiple shared templates
	templates := map[string]string{
		"Dockerfile.tmpl":     "FROM golang:[[.go_version]]",
		".dockerignore.tmpl":  "node_modules\n.git",
		".air.toml.tmpl":      "root = \".\"",
	}

	for name, content := range templates {
		err := os.WriteFile(filepath.Join(sharedDockerDir, name), []byte(content), 0600)
		require.NoError(t, err)
	}

	// Manifest referencing entire shared directory
	manifest := &ritual.Manifest{
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{
					Source:      "_shared:docker/",
					Destination: "./",
				},
			},
		},
	}

	gen := generator.NewFileGenerator("fith")
	vars := generator.NewVariables()
	vars.Set("go_version", "1.21")
	gen.SetVariables(vars)
	gen.SetRitualsBasePath(ritualsDir)

	err := gen.GenerateFiles(manifest, testRitualDir, outputDir)
	require.NoError(t, err)

	// Verify all files were created
	assert.FileExists(t, filepath.Join(outputDir, "Dockerfile"))
	assert.FileExists(t, filepath.Join(outputDir, ".dockerignore"))
	assert.FileExists(t, filepath.Join(outputDir, ".air.toml"))
}

// TestSharedTemplateNotFound tests error handling for missing shared templates
func TestSharedTemplateNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	ritualsDir := filepath.Join(tmpDir, "rituals")
	testRitualDir := filepath.Join(ritualsDir, "test-ritual", "templates")
	outputDir := filepath.Join(tmpDir, "output")

	require.NoError(t, os.MkdirAll(testRitualDir, 0750))

	manifest := &ritual.Manifest{
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{
					Source:      "_shared:docker/nonexistent.tmpl",
					Destination: "file.txt",
				},
			},
		},
	}

	gen := generator.NewFileGenerator("fith")
	gen.SetRitualsBasePath(ritualsDir)

	err := gen.GenerateFiles(manifest, testRitualDir, outputDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestSharedTemplateOptional tests optional shared templates
func TestSharedTemplateOptional(t *testing.T) {
	tmpDir := t.TempDir()
	ritualsDir := filepath.Join(tmpDir, "rituals")
	testRitualDir := filepath.Join(ritualsDir, "test-ritual", "templates")
	outputDir := filepath.Join(tmpDir, "output")

	require.NoError(t, os.MkdirAll(testRitualDir, 0750))

	manifest := &ritual.Manifest{
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{
					Source:      "_shared:docker/optional.tmpl",
					Destination: "optional.txt",
					Optional:    true,
				},
			},
		},
	}

	gen := generator.NewFileGenerator("fith")
	gen.SetRitualsBasePath(ritualsDir)

	// Should not error even though file doesn't exist
	err := gen.GenerateFiles(manifest, testRitualDir, outputDir)
	assert.NoError(t, err)
}

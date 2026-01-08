package docker_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerfileTemplateRendering tests that Dockerfile template renders correctly
func TestDockerfileTemplateRendering(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		contains []string
		notContains []string
	}{
		{
			name: "basic go application",
			data: map[string]interface{}{
				"app_name": "myapp",
				"port":     8080,
			},
			contains: []string{
				"FROM golang:",
				"alpine",
				"WORKDIR /app",
				"EXPOSE [[.port]]", // Template variable, not rendered
				"air",
			},
			notContains: []string{
				"node",
				"npm",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Navigate to project root for tests
			tmplPath := filepath.Join("..", "..", "rituals", "_shared", "docker", "Dockerfile.go.tmpl")
			
			// Check template exists
			_, err := os.Stat(tmplPath)
			require.NoError(t, err, "Dockerfile template should exist")

			// Parse and render template
			tmpl, err := template.ParseFiles(tmplPath)
			require.NoError(t, err, "Should parse Dockerfile template")

			var result strings.Builder
			err = tmpl.Execute(&result, tt.data)
			require.NoError(t, err, "Should execute template")

			output := result.String()

			// Verify content
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Dockerfile should contain: %s", expected)
			}

			for _, unexpected := range tt.notContains {
				assert.NotContains(t, output, unexpected, "Dockerfile should not contain: %s", unexpected)
			}
		})
	}
}

// TestDockerfileValidSyntax tests that generated Dockerfile has valid syntax
func TestDockerfileValidSyntax(t *testing.T) {
	t.Skip("Requires docker to be installed - will implement after template creation")
	
	// This test will:
	// 1. Generate Dockerfile from template
	// 2. Use `docker build --dry-run` or hadolint to validate
	// 3. Check for common issues (missing FROM, invalid commands, etc.)
}

// TestDockerIgnoreTemplateRendering tests .dockerignore template
func TestDockerIgnoreTemplateRendering(t *testing.T) {
	tmplPath := filepath.Join("..", "..", "rituals", "_shared", "docker", ".dockerignore.tmpl")
	
	// Check template exists
	_, err := os.Stat(tmplPath)
	require.NoError(t, err, ".dockerignore template should exist")

	// Parse template
	tmpl, err := template.ParseFiles(tmplPath)
	require.NoError(t, err, "Should parse .dockerignore template")

	// Render with empty data
	var result strings.Builder
	err = tmpl.Execute(&result, map[string]interface{}{})
	require.NoError(t, err, "Should execute template")

	output := result.String()

	// Should contain common exclusions
	expectedPatterns := []string{
		"node_modules",
		".git",
		"tmp",
		".env",
		"*.log",
	}

	for _, pattern := range expectedPatterns {
		assert.Contains(t, output, pattern, ".dockerignore should exclude: %s", pattern)
	}
}

// TestAirConfigTemplateRendering tests Air hot reload configuration
func TestAirConfigTemplateRendering(t *testing.T) {
	tmplPath := filepath.Join("..", "..", "rituals", "_shared", "docker", ".air.toml.tmpl")
	
	// Check template exists
	_, err := os.Stat(tmplPath)
	require.NoError(t, err, ".air.toml template should exist")

	// Parse template
	tmpl, err := template.ParseFiles(tmplPath)
	require.NoError(t, err, "Should parse .air.toml template")

	data := map[string]interface{}{
		"app_name": "myapp",
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	require.NoError(t, err, "Should execute template")

	output := result.String()

	// Verify Air configuration essentials
	expectedConfigs := []string{
		"root = \".\"",
		"tmp_dir = \"tmp\"",
		"[build]",
		"cmd =",
		"exclude_dir",
		"include_ext",
		"\"go\"",
	}

	for _, expected := range expectedConfigs {
		assert.Contains(t, output, expected, ".air.toml should contain: %s", expected)
	}

	// Should exclude frontend directories
	assert.Contains(t, output, "frontend", "Should exclude frontend directory")
	assert.Contains(t, output, "node_modules", "Should exclude node_modules")
}

// TestDockerImageSize tests that built images are reasonably sized
func TestDockerImageSize(t *testing.T) {
	t.Skip("Requires docker build - integration test")
	
	// This test will:
	// 1. Build the Docker image
	// 2. Check image size is < 500MB (with dev tools)
	// 3. Verify layers are optimized
}

// TestSecurityNoRootUser tests that containers don't run as root
func TestSecurityNoRootUser(t *testing.T) {
	t.Skip("Future enhancement - current version uses root for development")
	
	// This test will verify:
	// 1. USER directive is set to non-root
	// 2. Files have correct ownership
}

// TestSecurityNoSecrets tests that no secrets are in Dockerfile
func TestSecurityNoSecrets(t *testing.T) {
	tmplPath := filepath.Join("..", "..", "rituals", "_shared", "docker", "Dockerfile.go.tmpl")
	
	_, err := os.Stat(tmplPath)
	require.NoError(t, err)

	content, err := os.ReadFile(tmplPath)
	require.NoError(t, err)

	// Check for common secret patterns
	forbiddenPatterns := []string{
		"password=",
		"secret=",
		"token=",
		"api_key=",
		"private_key",
	}

	for _, pattern := range forbiddenPatterns {
		assert.NotContains(t, string(content), pattern, 
			"Dockerfile should not contain hardcoded secrets: %s", pattern)
	}
}

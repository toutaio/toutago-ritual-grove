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

// TestDockerComposeTemplateRendering tests docker-compose.yml template rendering
func TestDockerComposeTemplateRendering(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		contains []string
	}{
		{
			name: "postgres database",
			data: map[string]interface{}{
				"app_name":      "myblog",
				"database_type": "postgres",
				"port":          8080,
			},
			contains: []string{
				"version:",
				"services:",
				"app:",
				"db:",
				"postgres:16-alpine",
				"POSTGRES_USER",
				"networks:",
				"volumes:",
			},
		},
		{
			name: "mysql database",
			data: map[string]interface{}{
				"app_name":      "myblog",
				"database_type": "mysql",
				"port":          8080,
			},
			contains: []string{
				"version:",
				"services:",
				"app:",
				"db:",
				"mysql:8-alpine",
				"MYSQL_USER",
			},
		},
		{
			name: "with frontend",
			data: map[string]interface{}{
				"app_name":      "myblog",
				"database_type": "postgres",
				"has_frontend":  true,
			},
			contains: []string{
				"frontend:",
				"node:20-alpine",
				"npm run dev",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmplPath := filepath.Join("rituals", "_shared", "docker", "docker-compose.yml.tmpl")
			
			_, err := os.Stat(tmplPath)
			require.NoError(t, err, "docker-compose template should exist")

			tmpl, err := template.ParseFiles(tmplPath)
			require.NoError(t, err, "Should parse docker-compose template")

			var result strings.Builder
			err = tmpl.Execute(&result, tt.data)
			require.NoError(t, err, "Should execute template")

			output := result.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, 
					"docker-compose.yml should contain: %s", expected)
			}
		})
	}
}

// TestComposeFileValidation tests compose file syntax validation
func TestComposeFileValidation(t *testing.T) {
	t.Skip("Requires docker-compose - will implement as integration test")
	
	// This test will:
	// 1. Generate docker-compose.yml from template
	// 2. Run `docker-compose config --quiet`
	// 3. Check for validation errors
}

// TestServiceDependencies tests that service dependencies are correct
func TestServiceDependencies(t *testing.T) {
	tmplPath := filepath.Join("rituals", "_shared", "docker", "docker-compose.yml.tmpl")
	
	_, err := os.Stat(tmplPath)
	require.NoError(t, err)

	// Parse template with database
	tmpl, err := template.ParseFiles(tmplPath)
	require.NoError(t, err)

	data := map[string]interface{}{
		"app_name":      "testapp",
		"database_type": "postgres",
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	require.NoError(t, err)

	output := result.String()

	// App should depend on db
	assert.Contains(t, output, "depends_on:", "App service should have dependencies")
	assert.Contains(t, output, "db:", "App should depend on db service")
	assert.Contains(t, output, "condition: service_healthy", 
		"Should wait for db health check")
}

// TestHealthCheckConfiguration tests health check setup
func TestHealthCheckConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		databaseType string
		healthCheck  string
	}{
		{
			name:         "postgres health check",
			databaseType: "postgres",
			healthCheck:  "pg_isready",
		},
		{
			name:         "mysql health check",
			databaseType: "mysql",
			healthCheck:  "mysqladmin ping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmplPath := filepath.Join("rituals", "_shared", "docker", "docker-compose.yml.tmpl")
			
			tmpl, err := template.ParseFiles(tmplPath)
			require.NoError(t, err)

			data := map[string]interface{}{
				"database_type": tt.databaseType,
			}

			var result strings.Builder
			err = tmpl.Execute(&result, data)
			require.NoError(t, err)

			output := result.String()
			assert.Contains(t, output, "healthcheck:", "Should have health check")
			assert.Contains(t, output, tt.healthCheck, 
				"Should use correct health check command")
		})
	}
}

// TestVolumeConfiguration tests volume setup
func TestVolumeConfiguration(t *testing.T) {
	tmplPath := filepath.Join("rituals", "_shared", "docker", "docker-compose.yml.tmpl")
	
	tmpl, err := template.ParseFiles(tmplPath)
	require.NoError(t, err)

	data := map[string]interface{}{
		"app_name":      "testapp",
		"database_type": "postgres",
		"has_frontend":  true,
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	require.NoError(t, err)

	output := result.String()

	// Should have named volumes
	assert.Contains(t, output, "volumes:", "Should define volumes section")
	assert.Contains(t, output, "db-data:", "Should have db-data volume")
	assert.Contains(t, output, "go-cache:", "Should have go-cache volume")

	// Frontend should have node-modules volume
	assert.Contains(t, output, "node-modules:", "Should have node-modules volume for frontend")

	// App should mount source code
	assert.Contains(t, output, "./:/app", "Should mount source code for hot reload")
}

// TestNetworkConfiguration tests network setup
func TestNetworkConfiguration(t *testing.T) {
	tmplPath := filepath.Join("rituals", "_shared", "docker", "docker-compose.yml.tmpl")
	
	tmpl, err := template.ParseFiles(tmplPath)
	require.NoError(t, err)

	data := map[string]interface{}{
		"app_name": "testapp",
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	require.NoError(t, err)

	output := result.String()

	assert.Contains(t, output, "networks:", "Should define networks")
	assert.Contains(t, output, "app-network:", "Should have app-network")
	
	// All services should be on the same network
	serviceCount := strings.Count(output, "- app-network")
	assert.GreaterOrEqual(t, serviceCount, 2, 
		"At least app and db should be on app-network")
}

// TestEnvironmentVariables tests env var configuration
func TestEnvironmentVariables(t *testing.T) {
	tmplPath := filepath.Join("rituals", "_shared", "docker", "docker-compose.yml.tmpl")
	
	tmpl, err := template.ParseFiles(tmplPath)
	require.NoError(t, err)

	data := map[string]interface{}{
		"app_name":      "testapp",
		"database_type": "postgres",
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	require.NoError(t, err)

	output := result.String()

	// Should use environment variables with defaults
	envVars := []string{
		"${APP_PORT",
		"${DB_USER",
		"${DB_PASSWORD",
		"${DB_NAME",
	}

	for _, envVar := range envVars {
		assert.Contains(t, output, envVar, 
			"Should use environment variable: %s", envVar)
	}

	// Should have default values
	assert.Contains(t, output, ":-", "Should have default values for env vars")
}

// TestEnvExampleTemplate tests .env.example template
func TestEnvExampleTemplate(t *testing.T) {
	tmplPath := filepath.Join("rituals", "_shared", "docker", ".env.example.tmpl")
	
	_, err := os.Stat(tmplPath)
	require.NoError(t, err, ".env.example template should exist")

	tmpl, err := template.ParseFiles(tmplPath)
	require.NoError(t, err)

	data := map[string]interface{}{
		"app_name":      "myblog",
		"database_type": "postgres",
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	require.NoError(t, err)

	output := result.String()

	// Should have all necessary variables
	requiredVars := []string{
		"APP_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"COMPOSE_PROJECT_NAME",
	}

	for _, varName := range requiredVars {
		assert.Contains(t, output, varName, 
			".env.example should define: %s", varName)
	}

	// Should have comments
	assert.Contains(t, output, "#", "Should have comments explaining variables")
}

// TestConditionalFrontendService tests frontend service generation
func TestConditionalFrontendService(t *testing.T) {
	tmplPath := filepath.Join("rituals", "_shared", "docker", "docker-compose.yml.tmpl")
	
	tmpl, err := template.ParseFiles(tmplPath)
	require.NoError(t, err)

	// Test with frontend
	t.Run("with frontend", func(t *testing.T) {
		data := map[string]interface{}{
			"has_frontend": true,
		}

		var result strings.Builder
		err = tmpl.Execute(&result, data)
		require.NoError(t, err)

		output := result.String()
		assert.Contains(t, output, "frontend:", "Should have frontend service")
		assert.Contains(t, output, "node:20-alpine", "Should use Node.js image")
	})

	// Test without frontend
	t.Run("without frontend", func(t *testing.T) {
		data := map[string]interface{}{
			"has_frontend": false,
		}

		var result strings.Builder
		err = tmpl.Execute(&result, data)
		require.NoError(t, err)

		output := result.String()
		assert.NotContains(t, output, "frontend:", "Should not have frontend service")
		assert.NotContains(t, output, "node:20-alpine", "Should not have Node image")
	})
}

package generator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toutaio/toutago-ritual-grove/internal/generator"
)

// TestDockerImageHelper tests dockerImage template function
func TestDockerImageHelper(t *testing.T) {
	tests := []struct {
		name         string
		databaseType string
		expected     string
	}{
		{
			name:         "postgres database",
			databaseType: "postgres",
			expected:     "postgres:16-alpine",
		},
		{
			name:         "mysql database",
			databaseType: "mysql",
			expected:     "mysql:8-alpine",
		},
		{
			name:         "empty database",
			databaseType: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.DockerImage(tt.databaseType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDockerPortHelper tests dockerPort template function
func TestDockerPortHelper(t *testing.T) {
	tests := []struct {
		name         string
		databaseType string
		expected     int
	}{
		{
			name:         "postgres port",
			databaseType: "postgres",
			expected:     5432,
		},
		{
			name:         "mysql port",
			databaseType: "mysql",
			expected:     3306,
		},
		{
			name:         "empty database",
			databaseType: "",
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.DockerPort(tt.databaseType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHealthCheckHelper tests healthCheck template function
func TestHealthCheckHelper(t *testing.T) {
	tests := []struct {
		name         string
		databaseType string
		expected     string
	}{
		{
			name:         "postgres health check",
			databaseType: "postgres",
			expected:     "pg_isready -U ${DB_USER}",
		},
		{
			name:         "mysql health check",
			databaseType: "mysql",
			expected:     "mysqladmin ping -h localhost -u root -p${DB_ROOT_PASSWORD}",
		},
		{
			name:         "empty database",
			databaseType: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.HealthCheck(tt.databaseType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHasFrontendHelper tests hasFrontend template function
func TestHasFrontendHelper(t *testing.T) {
	tests := []struct {
		name         string
		frontendType string
		expected     bool
	}{
		{
			name:         "inertia-vue frontend",
			frontendType: "inertia-vue",
			expected:     true,
		},
		{
			name:         "htmx frontend",
			frontendType: "htmx",
			expected:     false, // htmx doesn't need separate build service
		},
		{
			name:         "traditional frontend",
			frontendType: "traditional",
			expected:     false,
		},
		{
			name:         "empty frontend",
			frontendType: "",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.HasFrontend(tt.frontendType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTemplateHelpersIntegration tests using helpers in templates
func TestTemplateHelpersIntegration(t *testing.T) {
	engine := generator.NewTemplateEngine("fith")
	
	// Register template helpers
	engine.RegisterFunc("dockerImage", generator.DockerImage)
	engine.RegisterFunc("dockerPort", generator.DockerPort)
	engine.RegisterFunc("healthCheck", generator.HealthCheck)
	engine.RegisterFunc("hasFrontend", generator.HasFrontend)

	t.Run("use dockerImage in template", func(t *testing.T) {
		template := `image: [[dockerImage .database_type]]`
		vars := map[string]interface{}{
			"database_type": "postgres",
		}

		result, err := engine.Render(template, vars)
		assert.NoError(t, err)
		assert.Equal(t, "image: postgres:16-alpine", result)
	})

	t.Run("use dockerPort in template", func(t *testing.T) {
		template := `port: [[dockerPort .database_type]]`
		vars := map[string]interface{}{
			"database_type": "mysql",
		}

		result, err := engine.Render(template, vars)
		assert.NoError(t, err)
		assert.Equal(t, "port: 3306", result)
	})

	t.Run("use healthCheck in template", func(t *testing.T) {
		template := `healthcheck: [[healthCheck .database_type]]`
		vars := map[string]interface{}{
			"database_type": "postgres",
		}

		result, err := engine.Render(template, vars)
		assert.NoError(t, err)
		assert.Contains(t, result, "pg_isready")
	})

	t.Run("use hasFrontend in conditional", func(t *testing.T) {
		template := `[[- if hasFrontend .frontend_type]]
frontend: true
[[- end]]`
		vars := map[string]interface{}{
			"frontend_type": "inertia-vue",
		}

		result, err := engine.Render(template, vars)
		assert.NoError(t, err)
		assert.Contains(t, result, "frontend: true")
	})

	t.Run("hasFrontend returns false for htmx", func(t *testing.T) {
		template := `[[- if hasFrontend .frontend_type]]
frontend: true
[[- else]]
frontend: false
[[- end]]`
		vars := map[string]interface{}{
			"frontend_type": "htmx",
		}

		result, err := engine.Render(template, vars)
		assert.NoError(t, err)
		assert.Contains(t, result, "frontend: false")
	})
}

// TestDBUserDefaultHelper tests dbUser default value generation
func TestDBUserDefaultHelper(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		expected    string
	}{
		{
			name:        "simple project name",
			projectName: "myapp",
			expected:    "myapp_user",
		},
		{
			name:        "project with dashes",
			projectName: "my-app",
			expected:    "my_app_user",
		},
		{
			name:        "long project name",
			projectName: "very-long-project-name",
			expected:    "very_long_project_name_user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.DBUser(tt.projectName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDBNameDefaultHelper tests dbName default value generation
func TestDBNameDefaultHelper(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		expected    string
	}{
		{
			name:        "simple project name",
			projectName: "myapp",
			expected:    "myapp_db",
		},
		{
			name:        "project with dashes",
			projectName: "my-app",
			expected:    "my_app_db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.DBName(tt.projectName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

package integration_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMinimalRitualWithDocker tests the minimal ritual with Docker support
func TestMinimalRitualWithDocker(t *testing.T) {
	t.Skip("Full integration test - requires ritual grove and Docker")
	
	tmpDir, err := ioutil.TempDir("", "ritual-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Generate project using ritual
	// This will be implemented after generator supports Docker
	t.Run("generate project", func(t *testing.T) {
		// cmd := exec.Command("touta", "ritual", "init", "minimal", 
		//	"--output", tmpDir, "--app-name", "testapp")
		// err := cmd.Run()
		// require.NoError(t, err, "Should generate project")
	})

	t.Run("verify docker files exist", func(t *testing.T) {
		requiredFiles := []string{
			"Dockerfile",
			"docker-compose.yml",
			".dockerignore",
			".air.toml",
			".env.example",
		}

		for _, file := range requiredFiles {
			path := filepath.Join(tmpDir, file)
			_, err := os.Stat(path)
			assert.NoError(t, err, "File should exist: %s", file)
		}
	})

	t.Run("docker build succeeds", func(t *testing.T) {
		cmd := exec.Command("docker-compose", "build")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Docker build should succeed: %s", string(output))
	})

	t.Run("containers start successfully", func(t *testing.T) {
		cmd := exec.Command("docker-compose", "up", "-d")
		cmd.Dir = tmpDir
		err := cmd.Run()
		require.NoError(t, err, "Containers should start")

		// Wait for services
		time.Sleep(10 * time.Second)
	})

	t.Run("http endpoint responds", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080")
		require.NoError(t, err, "Should connect to app")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return 200 OK")
	})

	t.Run("cleanup", func(t *testing.T) {
		cmd := exec.Command("docker-compose", "down", "-v")
		cmd.Dir = tmpDir
		err := cmd.Run()
		require.NoError(t, err, "Should cleanup containers")
	})
}

// TestBlogRitualWithPostgres tests blog ritual with PostgreSQL
func TestBlogRitualWithPostgres(t *testing.T) {
	t.Skip("Full integration test")
	
	tmpDir, err := ioutil.TempDir("", "ritual-blog-postgres-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test would:
	// 1. Generate blog project with PostgreSQL
	// 2. Verify Docker files
	// 3. Start containers
	// 4. Test app + database + frontend all working
	// 5. Test hot reload
	// 6. Cleanup
}

// TestBlogRitualWithMySQL tests blog ritual with MySQL
func TestBlogRitualWithMySQL(t *testing.T) {
	t.Skip("Full integration test")
	
	tmpDir, err := ioutil.TempDir("", "ritual-blog-mysql-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Similar to PostgreSQL test but with MySQL
}

// TestWikiRitualWithDocker tests wiki ritual
func TestWikiRitualWithDocker(t *testing.T) {
	t.Skip("Full integration test")
	
	// Test wiki ritual with Docker support
}

// TestFullstackInertiaVueWithDocker tests fullstack-inertia-vue ritual
func TestFullstackInertiaVueWithDocker(t *testing.T) {
	t.Skip("Full integration test")
	
	tmpDir, err := ioutil.TempDir("", "ritual-fullstack-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test would verify:
	// 1. Frontend service starts
	// 2. esbuild watch mode works
	// 3. Frontend hot reload functions
	// 4. App can serve built assets
}

// TestHotReloadIntegration tests end-to-end hot reload
func TestHotReloadIntegration(t *testing.T) {
	t.Skip("Full integration test")
	
	tmpDir, err := ioutil.TempDir("", "ritual-hotreload-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Generate project
	// Start containers
	// Get initial response from endpoint
	// Modify a .go file
	// Wait for Air to rebuild
	// Get new response
	// Verify response changed
	// Cleanup
}

// TestVolumePersistenceIntegration tests data persistence
func TestVolumePersistenceIntegration(t *testing.T) {
	t.Skip("Full integration test")
	
	// Create data in database
	// Stop containers (without -v)
	// Start containers again
	// Verify data still exists
}

// TestDatabaseMigrations tests that migrations run on startup
func TestDatabaseMigrations(t *testing.T) {
	t.Skip("Full integration test")
	
	// Generate project with migrations
	// Start containers
	// Verify migrations table exists
	// Verify schema is created
}

// TestEnvironmentVariableOverrides tests .env file customization
func TestEnvironmentVariableOverrides(t *testing.T) {
	t.Skip("Full integration test")
	
	tmpDir, err := ioutil.TempDir("", "ritual-env-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Generate project
	// Create custom .env file
	// Start containers
	// Verify custom values are used
}

// TestMultipleProjectsSimultaneously tests running multiple ritual projects
func TestMultipleProjectsSimultaneously(t *testing.T) {
	t.Skip("Full integration test - port conflicts")
	
	// This test would verify:
	// 1. Generate two projects
	// 2. Customize ports in .env files
	// 3. Start both projects
	// 4. Verify both are running
	// 5. No port conflicts
}

// TestDockerComposeValidation tests all generated docker-compose files are valid
func TestDockerComposeValidation(t *testing.T) {
	t.Skip("Integration test")
	
	rituals := []string{"minimal", "hello-world", "basic-site", "blog", "wiki", "fullstack-inertia-vue"}
	
	for _, ritual := range rituals {
		t.Run(ritual, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "ritual-validate-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Generate project
			// Run docker-compose config --quiet
			// Verify no errors
		})
	}
}

// TestBuildPerformance tests that builds complete in reasonable time
func TestBuildPerformance(t *testing.T) {
	t.Skip("Performance test")
	
	tmpDir, err := ioutil.TempDir("", "ritual-perf-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Generate project
	start := time.Now()
	
	cmd := exec.Command("docker-compose", "build")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	duration := time.Since(start)
	
	// Build should complete in < 2 minutes
	assert.Less(t, duration, 2*time.Minute, 
		"Build should complete quickly")
	
	t.Logf("Build completed in %v", duration)
}

// TestImageSizes tests that Docker images are reasonably sized
func TestImageSizes(t *testing.T) {
	t.Skip("Integration test")
	
	// Build images
	// Check image sizes
	// Development image should be < 500MB
	// Verify no unnecessary layers
}

// TestSecurityScan tests Docker images for vulnerabilities
func TestSecurityScan(t *testing.T) {
	t.Skip("Security test - requires docker scan or trivy")
	
	// Build image
	// Run security scanner
	// Verify no critical vulnerabilities
}

// TestDockerIgnoreEffectiveness tests that .dockerignore works
func TestDockerIgnoreEffectiveness(t *testing.T) {
	t.Skip("Integration test")
	
	tmpDir, err := ioutil.TempDir("", "ritual-dockerignore-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Generate project
	// Create files that should be ignored (node_modules, .git, etc.)
	// Build image
	// Inspect image layers
	// Verify ignored files are not in image
}

// Helper: cleanupDockerResources removes all test containers and volumes
func cleanupDockerResources(t *testing.T, projectDir string) {
	cmd := exec.Command("docker-compose", "down", "-v", "--remove-orphans")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Cleanup warning: %s: %s", err, string(output))
	}
}

// Helper: waitForHealthy waits for all services to be healthy
func waitForHealthy(ctx context.Context, projectDir string) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			cmd := exec.Command("docker-compose", "ps")
			cmd.Dir = projectDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				continue
			}

			// Check if all services are up and healthy
			lines := strings.Split(string(output), "\n")
			allHealthy := true
			for _, line := range lines {
				if strings.Contains(line, "unhealthy") || 
				   strings.Contains(line, "starting") {
					allHealthy = false
					break
				}
			}

			if allHealthy {
				return nil
			}
		}
	}
}

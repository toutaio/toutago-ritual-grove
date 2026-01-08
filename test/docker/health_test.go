package docker_test

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContainerStarts tests that containers start successfully
func TestContainerStarts(t *testing.T) {
	t.Skip("Integration test - requires Docker")
	
	ctx := context.Background()
	
	// Start containers
	cmd := exec.CommandContext(ctx, "docker-compose", "up", "-d")
	err := cmd.Run()
	require.NoError(t, err, "Containers should start")

	// Wait a bit for startup
	time.Sleep(5 * time.Second)

	// Check containers are running
	cmd = exec.CommandContext(ctx, "docker-compose", "ps")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	assert.Contains(t, string(output), "Up", "Containers should be running")
}

// TestHealthChecksPass tests that all health checks pass
func TestHealthChecksPass(t *testing.T) {
	t.Skip("Integration test - requires Docker")
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Wait for all services to be healthy
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Timeout waiting for services to be healthy")
		case <-ticker.C:
			cmd := exec.Command("docker-compose", "ps")
			output, err := cmd.CombinedOutput()
			require.NoError(t, err)

			// Check if all services are healthy
			lines := strings.Split(string(output), "\n")
			allHealthy := true
			for _, line := range lines {
				if strings.Contains(line, "unhealthy") || strings.Contains(line, "starting") {
					allHealthy = false
					break
				}
			}

			if allHealthy {
				return // Test passed
			}
		}
	}
}

// TestHTTPEndpointResponds tests that the app HTTP endpoint responds
func TestHTTPEndpointResponds(t *testing.T) {
	t.Skip("Integration test - requires Docker")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try to connect to app endpoint
	client := &http.Client{Timeout: 2 * time.Second}
	
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Timeout waiting for HTTP endpoint")
		case <-ticker.C:
			resp, err := client.Get("http://localhost:8080")
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return // Success
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}

// TestDatabaseConnectivity tests database accepts connections
func TestDatabaseConnectivity(t *testing.T) {
	t.Skip("Integration test - requires Docker")
	
	tests := []struct {
		name     string
		dbType   string
		command  []string
	}{
		{
			name:   "postgres connection",
			dbType: "postgres",
			command: []string{
				"docker-compose", "exec", "-T", "db",
				"psql", "-U", "testuser", "-d", "testdb", "-c", "SELECT 1",
			},
		},
		{
			name:   "mysql connection",
			dbType: "mysql",
			command: []string{
				"docker-compose", "exec", "-T", "db",
				"mysql", "-u", "testuser", "-ptestpass", "-e", "SELECT 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(tt.command[0], tt.command[1:]...)
			output, err := cmd.CombinedOutput()
			
			require.NoError(t, err, "Should connect to database")
			assert.NotEmpty(t, output, "Should get response from database")
		})
	}
}

// TestServiceDiscovery tests that services can communicate
func TestServiceDiscovery(t *testing.T) {
	t.Skip("Integration test - requires Docker")
	
	// Test that app can resolve 'db' hostname
	cmd := exec.Command("docker-compose", "exec", "-T", "app", "ping", "-c", "1", "db")
	err := cmd.Run()
	require.NoError(t, err, "App should be able to ping db service")

	// Test DNS resolution
	cmd = exec.Command("docker-compose", "exec", "-T", "app", "nslookup", "db")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Should resolve db hostname")
	assert.Contains(t, string(output), "Address", "DNS should resolve")
}

// TestVolumePersistence tests that data persists in volumes
func TestVolumePersistence(t *testing.T) {
	t.Skip("Integration test - requires Docker")
	
	// Write test data to database
	cmd := exec.Command("docker-compose", "exec", "-T", "db",
		"psql", "-U", "testuser", "-d", "testdb",
		"-c", "CREATE TABLE test (id INTEGER PRIMARY KEY, data TEXT)",
	)
	err := cmd.Run()
	require.NoError(t, err, "Should create test table")

	cmd = exec.Command("docker-compose", "exec", "-T", "db",
		"psql", "-U", "testuser", "-d", "testdb",
		"-c", "INSERT INTO test VALUES (1, 'test_data')",
	)
	err = cmd.Run()
	require.NoError(t, err, "Should insert test data")

	// Restart database container
	cmd = exec.Command("docker-compose", "restart", "db")
	err = cmd.Run()
	require.NoError(t, err, "Should restart db container")

	// Wait for database to be ready
	time.Sleep(10 * time.Second)

	// Verify data still exists
	cmd = exec.Command("docker-compose", "exec", "-T", "db",
		"psql", "-U", "testuser", "-d", "testdb",
		"-c", "SELECT data FROM test WHERE id = 1",
	)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Should query data after restart")
	assert.Contains(t, string(output), "test_data", "Data should persist")
}

// TestHotReloadFunctionality tests that hot reload works
func TestHotReloadFunctionality(t *testing.T) {
	t.Skip("Integration test - requires Docker and file watching")
	
	// This test would:
	// 1. Start containers
	// 2. Modify a .go file
	// 3. Watch logs for Air rebuild message
	// 4. Verify new code is running
	// 5. Check HTTP endpoint returns updated response
}

// TestContainerHealthEndpoints tests health check endpoints
func TestContainerHealthEndpoints(t *testing.T) {
	t.Skip("Integration test - requires Docker")
	
	// Get container health status from Docker
	cmd := exec.Command("docker", "inspect", 
		"--format='{{.State.Health.Status}}'",
		"testapp_db")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	status := strings.TrimSpace(string(output))
	assert.Equal(t, "'healthy'", status, "Container should be healthy")
}

// TestLogOutput tests that containers produce expected logs
func TestLogOutput(t *testing.T) {
	t.Skip("Integration test - requires Docker")
	
	// Get app logs
	cmd := exec.Command("docker-compose", "logs", "app")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	logs := string(output)
	
	// Should show Air starting
	assert.Contains(t, logs, "air", "Logs should show Air is running")
	
	// Should show app starting
	assert.Contains(t, logs, "Starting", "Logs should show app starting")
}

// TestResourceUsage tests that containers don't use excessive resources
func TestResourceUsage(t *testing.T) {
	t.Skip("Integration test - requires Docker")
	
	// Get container stats
	cmd := exec.Command("docker", "stats", "--no-stream", "--format", 
		"{{.Container}}\t{{.MemUsage}}\t{{.CPUPerc}}")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	stats := string(output)
	t.Logf("Container stats:\n%s", stats)
	
	// Basic sanity check - containers should be using some resources
	assert.NotEmpty(t, stats, "Should get container stats")
}

// TestCleanup tests that docker-compose down cleans up properly
func TestCleanup(t *testing.T) {
	t.Skip("Integration test - requires Docker")
	
	// Stop and remove containers
	cmd := exec.Command("docker-compose", "down")
	err := cmd.Run()
	require.NoError(t, err, "Should stop containers")

	// Verify containers are gone
	cmd = exec.Command("docker-compose", "ps", "-q")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(string(output)), 
		"No containers should be running")
}

// TestCleanupWithVolumes tests cleanup including volumes
func TestCleanupWithVolumes(t *testing.T) {
	t.Skip("Integration test - requires Docker")
	
	// Get volume names before cleanup
	cmd := exec.Command("docker", "volume", "ls", "-q")
	beforeOutput, err := cmd.CombinedOutput()
	require.NoError(t, err)
	volumesBefore := strings.Split(strings.TrimSpace(string(beforeOutput)), "\n")

	// Stop and remove containers with volumes
	cmd = exec.Command("docker-compose", "down", "-v")
	err = cmd.Run()
	require.NoError(t, err, "Should stop containers and remove volumes")

	// Get volume names after cleanup
	cmd = exec.Command("docker", "volume", "ls", "-q")
	afterOutput, err := cmd.CombinedOutput()
	require.NoError(t, err)
	volumesAfter := strings.Split(strings.TrimSpace(string(afterOutput)), "\n")

	// Volumes should be removed
	assert.Less(t, len(volumesAfter), len(volumesBefore), 
		"Volumes should be removed")
}

// Helper function to wait for service readiness
func waitForService(address string, timeout time.Duration) error {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(address)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("service not ready after %v", timeout)
}

package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDeploymentHistory(t *testing.T) {
	tmpDir := t.TempDir()

	history := &DeploymentHistory{}

	// Test 1: Add successful deployment
	t.Run("AddSuccessfulDeployment", func(t *testing.T) {
		history.AddDeployment(DeploymentRecord{
			Timestamp:   time.Now(),
			FromVersion: "1.0.0",
			ToVersion:   "1.1.0",
			Status:      "success",
			Message:     "Successfully updated to v1.1.0",
		})

		if len(history.Deployments) != 1 {
			t.Errorf("Expected 1 deployment, got %d", len(history.Deployments))
		}

		if history.Deployments[0].Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", history.Deployments[0].Status)
		}
	})

	// Test 2: Add failed deployment
	t.Run("AddFailedDeployment", func(t *testing.T) {
		history.AddDeployment(DeploymentRecord{
			Timestamp:   time.Now(),
			FromVersion: "1.1.0",
			ToVersion:   "1.2.0",
			Status:      "failure",
			Message:     "Migration failed",
			Errors:      []string{"database connection failed"},
		})

		if len(history.Deployments) != 2 {
			t.Errorf("Expected 2 deployments, got %d", len(history.Deployments))
		}

		if history.Deployments[1].Status != "failure" {
			t.Errorf("Expected status 'failure', got '%s'", history.Deployments[1].Status)
		}

		if len(history.Deployments[1].Errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(history.Deployments[1].Errors))
		}
	})

	// Test 3: Save and load history
	t.Run("SaveAndLoadHistory", func(t *testing.T) {
		err := history.Save(tmpDir)
		if err != nil {
			t.Fatalf("Failed to save history: %v", err)
		}

		// Verify file exists
		historyPath := filepath.Join(tmpDir, ".ritual", "history.yaml")
		if _, err := os.Stat(historyPath); os.IsNotExist(err) {
			t.Fatal("History file was not created")
		}

		// Load history
		loaded, err := LoadDeploymentHistory(tmpDir)
		if err != nil {
			t.Fatalf("Failed to load history: %v", err)
		}

		if len(loaded.Deployments) != 2 {
			t.Errorf("Expected 2 deployments after loading, got %d", len(loaded.Deployments))
		}
	})

	// Test 4: Get latest successful deployment
	t.Run("GetLatestSuccessful", func(t *testing.T) {
		latest := history.GetLatestSuccessful()
		if latest == nil {
			t.Fatal("Expected to find latest successful deployment")
		}

		if latest.ToVersion != "1.1.0" {
			t.Errorf("Expected latest successful version 1.1.0, got %s", latest.ToVersion)
		}
	})

	// Test 5: Get all failures
	t.Run("GetFailures", func(t *testing.T) {
		failures := history.GetFailures()
		if len(failures) != 1 {
			t.Errorf("Expected 1 failure, got %d", len(failures))
		}

		if failures[0].ToVersion != "1.2.0" {
			t.Errorf("Expected failure version 1.2.0, got %s", failures[0].ToVersion)
		}
	})

	// Test 6: Limit history size
	t.Run("LimitHistorySize", func(t *testing.T) {
		// Add many deployments
		for i := 0; i < 105; i++ {
			history.AddDeployment(DeploymentRecord{
				Timestamp:   time.Now(),
				FromVersion: "1.0.0",
				ToVersion:   "2.0.0",
				Status:      "success",
			})
		}

		// Should keep only last 100
		if len(history.Deployments) > 100 {
			t.Errorf("Expected max 100 deployments, got %d", len(history.Deployments))
		}
	})

	// Test 7: Load non-existent history
	t.Run("LoadNonExistentHistory", func(t *testing.T) {
		nonExistentDir := filepath.Join(tmpDir, "non-existent")
		loaded, err := LoadDeploymentHistory(nonExistentDir)
		if err == nil {
			t.Fatal("Expected error when loading non-existent history")
		}
		if loaded != nil {
			t.Error("Expected nil history when file doesn't exist")
		}
	})
}

package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProtectedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test state
	state := &State{
		RitualName:     "test-ritual",
		RitualVersion:  "1.0.0",
		ProtectedFiles: []string{"config.yaml", "secrets.env"},
	}

	pm := NewProtectedFileManager(state)

	// Test 1: Check if file is protected
	t.Run("IsProtected", func(t *testing.T) {
		if !pm.IsProtected("config.yaml") {
			t.Error("Expected config.yaml to be protected")
		}

		if !pm.IsProtected("secrets.env") {
			t.Error("Expected secrets.env to be protected")
		}

		if pm.IsProtected("other.go") {
			t.Error("Expected other.go to NOT be protected")
		}
	})

	// Test 2: Add protected file
	t.Run("AddProtectedFile", func(t *testing.T) {
		pm.AddProtectedFile("custom.yaml")

		if !pm.IsProtected("custom.yaml") {
			t.Error("Expected custom.yaml to be protected after adding")
		}

		if len(state.ProtectedFiles) != 3 {
			t.Errorf("Expected 3 protected files, got %d", len(state.ProtectedFiles))
		}
	})

	// Test 3: Remove protected file
	t.Run("RemoveProtectedFile", func(t *testing.T) {
		pm.RemoveProtectedFile("secrets.env")

		if pm.IsProtected("secrets.env") {
			t.Error("Expected secrets.env to NOT be protected after removing")
		}

		if len(state.ProtectedFiles) != 2 {
			t.Errorf("Expected 2 protected files, got %d", len(state.ProtectedFiles))
		}
	})

	// Test 4: Load user-defined protected files
	t.Run("LoadUserProtectedFiles", func(t *testing.T) {
		// Create .ritual/protected.txt
		ritualDir := filepath.Join(tmpDir, ".ritual")
		err := os.MkdirAll(ritualDir, 0750)
		if err != nil {
			t.Fatalf("Failed to create ritual dir: %v", err)
		}

		protectedFile := filepath.Join(ritualDir, "protected.txt")
		content := []byte("user-config.yaml\n# This is a comment\nmy-secret.env\n\n")
		err = os.WriteFile(protectedFile, content, 0600)
		if err != nil {
			t.Fatalf("Failed to write protected file: %v", err)
		}

		userFiles, err := pm.LoadUserProtectedFiles(tmpDir)
		if err != nil {
			t.Fatalf("Failed to load user protected files: %v", err)
		}

		if len(userFiles) != 2 {
			t.Errorf("Expected 2 user protected files, got %d", len(userFiles))
		}

		expected := map[string]bool{
			"user-config.yaml": true,
			"my-secret.env":    true,
		}

		for _, file := range userFiles {
			if !expected[file] {
				t.Errorf("Unexpected protected file: %s", file)
			}
		}
	})

	// Test 5: Get all protected files
	t.Run("GetAllProtectedFiles", func(t *testing.T) {
		all := pm.GetAllProtectedFiles()

		// Should include both state protected files and might include user files
		if len(all) < 2 {
			t.Errorf("Expected at least 2 protected files, got %d", len(all))
		}
	})

	// Test 6: Avoid duplicate protected files
	t.Run("AvoidDuplicates", func(t *testing.T) {
		pm.AddProtectedFile("config.yaml") // Already exists

		count := 0
		for _, f := range state.ProtectedFiles {
			if f == "config.yaml" {
				count++
			}
		}

		if count > 1 {
			t.Error("Protected file should not be added twice")
		}
	})
}

func TestProtectedFilePattern(t *testing.T) {
	state := &State{
		RitualName:     "test",
		RitualVersion:  "1.0.0",
		ProtectedFiles: []string{"*.env", "config/*.yaml"},
	}

	pm := NewProtectedFileManager(state)

	// Test pattern matching
	t.Run("PatternMatching", func(t *testing.T) {
		if !pm.IsProtected("secrets.env") {
			t.Error("Expected *.env pattern to match secrets.env")
		}

		if !pm.IsProtected("config/database.yaml") {
			t.Error("Expected config/*.yaml pattern to match config/database.yaml")
		}

		if pm.IsProtected("other/file.txt") {
			t.Error("Did not expect other/file.txt to match any pattern")
		}
	})
}

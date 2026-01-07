package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/storage"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// TestPlanCommandWithProtectedFiles tests that the plan command respects protected files
func TestPlanCommandWithProtectedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a .ritual directory with state
	ritualDir := filepath.Join(tmpDir, ".ritual")
	if err := os.MkdirAll(ritualDir, 0750); err != nil {
		t.Fatalf("Failed to create ritual dir: %v", err)
	}

	// Create state
	state := &storage.State{
		RitualName:    "test-ritual",
		RitualVersion: "1.0.0",
		ProtectedFiles: []string{
			"config.yaml",
			"*.env",
		},
	}

	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create protected.txt with user-defined protected files
	protectedTxt := filepath.Join(ritualDir, "protected.txt")
	content := "# User protected files\nmy-custom.conf\n"
	if err := os.WriteFile(protectedTxt, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write protected.txt: %v", err)
	}

	// Load state and protected files
	loadedState, err := storage.LoadState(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	pm := storage.NewProtectedFileManager(loadedState)

	// Load user-defined protected files
	userFiles, err := pm.LoadUserProtectedFiles(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load user protected files: %v", err)
	}

	// Verify user file was loaded
	if len(userFiles) != 1 {
		t.Errorf("Expected 1 user protected file, got %d", len(userFiles))
	}

	// Verify all protected files are recognized
	testCases := []struct {
		file       string
		shouldBeProtected bool
	}{
		{"config.yaml", true},
		{"secrets.env", true},
		{"my-custom.conf", true},
		{"main.go", false},
	}

	for _, tc := range testCases {
		isProtected := pm.IsProtected(tc.file)
		if isProtected != tc.shouldBeProtected {
			t.Errorf("File %s: expected protected=%v, got %v", tc.file, tc.shouldBeProtected, isProtected)
		}
	}
}

// TestProtectedFilesInDiffGeneration tests that protected files appear in conflicts
func TestProtectedFilesInDiffGeneration(t *testing.T) {
	// This is tested in deployment/diff_test.go
	// This test documents the integration point
	t.Skip("Tested in deployment package")
}

// TestLoadProtectedFilesHelper tests the helper function for loading protected files
func TestLoadProtectedFilesHelper(t *testing.T) {
	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, ".ritual")

	// Create state with initial protected files
	state := &storage.State{
		RitualName:     "test",
		RitualVersion:  "1.0.0",
		ProtectedFiles: []string{"initial.yaml"},
	}

	if err := os.MkdirAll(ritualDir, 0750); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create protected.txt
	protectedTxt := filepath.Join(ritualDir, "protected.txt")
	content := "user1.conf\nuser2.env\n# Comment\n\nuser3.txt\n"
	if err := os.WriteFile(protectedTxt, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write protected.txt: %v", err)
	}

	// Helper function to load all protected files
	allProtected, err := loadAllProtectedFiles(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load protected files: %v", err)
	}

	// Should have initial + user files
	if len(allProtected) < 4 {
		t.Errorf("Expected at least 4 protected files, got %d", len(allProtected))
	}

	// Check presence of user files
	hasUser1 := false
	for _, f := range allProtected {
		if f == "user1.conf" {
			hasUser1 = true
		}
	}

	if !hasUser1 {
		t.Error("Expected user1.conf in protected files")
	}
}

// loadAllProtectedFiles is a helper that combines state and user protected files
// This should be moved to a shared utility
func loadAllProtectedFiles(projectPath string) ([]string, error) {
	state, err := storage.LoadState(projectPath)
	if err != nil {
		return nil, err
	}

	pm := storage.NewProtectedFileManager(state)

	// Load user-defined protected files
	_, err = pm.LoadUserProtectedFiles(projectPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return pm.GetAllProtectedFiles(), nil
}

// TestManifestProtectedFiles tests that ritual manifest can specify protected files
func TestManifestProtectedFiles(t *testing.T) {
	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test",
			Version: "1.0.0",
		},
		Files: ritual.FilesSection{
			Protected: []string{
				"config/*.yaml",
				".env*",
				"secrets/*",
			},
		},
	}

	// Verify manifest can specify protected files
	if len(manifest.Files.Protected) != 3 {
		t.Errorf("Expected 3 protected file patterns, got %d", len(manifest.Files.Protected))
	}
}

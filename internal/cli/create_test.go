package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		ritualPath  string
		answers     map[string]interface{}
		expectFiles []string
		expectError bool
	}{
		{
			name:       "create from minimal ritual",
			ritualPath: "../../rituals/minimal",
			answers: map[string]interface{}{
				"project_name": "test-project",
				"module_name":  "github.com/test/test-project",
				"port":         8080,
			},
			expectFiles: []string{
				"cmd/test-project/main.go",
				"go.mod",
				"README.md",
				".gitignore",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir, err := os.MkdirTemp("", "ritual-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create workflow
			workflow := NewCreateWorkflow()

			// Execute creation
			err = workflow.Execute(tt.ritualPath, tempDir, tt.answers, false)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify expected files exist
			if !tt.expectError {
				for _, expectedFile := range tt.expectFiles {
					filePath := filepath.Join(tempDir, expectedFile)
					if _, err := os.Stat(filePath); os.IsNotExist(err) {
						// List what was actually created for debugging
						t.Logf("Expected file not found: %s", expectedFile)
						t.Logf("Contents of %s:", tempDir)
						filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
							if err == nil && !info.IsDir() {
								relPath, _ := filepath.Rel(tempDir, path)
								t.Logf("  - %s", relPath)
							}
							return nil
						})
						t.Errorf("Expected file not found: %s", expectedFile)
					}
				}
			}
		})
	}
}

func TestCreateWorkflow_WithQuestionnaire(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ritual-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	workflow := NewCreateWorkflow()

	// Test with questionnaire mode (no answers provided)
	err = workflow.Execute("../../rituals/minimal", tempDir, nil, false)

	// This should fail in automated tests since we can't provide interactive input
	// In real usage, this would show the interactive questionnaire
	if err == nil {
		t.Error("Expected error when running questionnaire without answers")
	}
}

func TestCreateWorkflow_DryRun(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ritual-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	workflow := NewCreateWorkflow()

	answers := map[string]interface{}{
		"project_name": "test-project",
		"module_path":  "github.com/test/test-project",
	}

	// Execute in dry-run mode
	err = workflow.Execute("../../rituals/minimal", tempDir, answers, true)
	if err != nil {
		t.Errorf("Unexpected error in dry-run: %v", err)
	}

	// Verify no files were created
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	if len(entries) > 0 {
		t.Errorf("Expected no files in dry-run mode, but found %d", len(entries))
	}
}

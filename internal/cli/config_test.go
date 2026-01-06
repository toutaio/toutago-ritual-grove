package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAnswersFromFile_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create YAML config file
	configPath := filepath.Join(tmpDir, "answers.yaml")
	configContent := `project_name: my-blog
module_name: github.com/user/my-blog
port: 8080
database: postgres
enable_auth: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load answers
	answers, err := LoadAnswersFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load answers: %v", err)
	}

	// Verify values
	tests := []struct {
		key      string
		expected interface{}
	}{
		{"project_name", "my-blog"},
		{"module_name", "github.com/user/my-blog"},
		{"port", 8080},
		{"database", "postgres"},
		{"enable_auth", true},
	}

	for _, tt := range tests {
		got, ok := answers[tt.key]
		if !ok {
			t.Errorf("Expected key %q not found", tt.key)
			continue
		}
		
		// Handle numeric comparisons
		if expectedInt, ok := tt.expected.(int); ok {
			// YAML unmarshals integers as int
			if gotInt, ok := got.(int); ok {
				if gotInt != expectedInt {
					t.Errorf("Key %q = %v, want %v", tt.key, gotInt, expectedInt)
				}
			} else {
				t.Errorf("Key %q type mismatch: got %T, want int", tt.key, got)
			}
		} else if got != tt.expected {
			t.Errorf("Key %q = %v, want %v", tt.key, got, tt.expected)
		}
	}
}

func TestLoadAnswersFromFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create JSON config file
	configPath := filepath.Join(tmpDir, "answers.json")
	configContent := `{
  "project_name": "my-api",
  "module_name": "github.com/user/my-api",
  "port": 3000,
  "enable_auth": false
}`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load answers
	answers, err := LoadAnswersFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load answers: %v", err)
	}

	// Verify values
	if answers["project_name"] != "my-api" {
		t.Errorf("project_name = %v, want 'my-api'", answers["project_name"])
	}
	
	// JSON unmarshals numbers as float64
	if port, ok := answers["port"].(float64); !ok || port != 3000 {
		t.Errorf("port = %v (%T), want 3000", answers["port"], answers["port"])
	}
	
	if auth, ok := answers["enable_auth"].(bool); !ok || auth != false {
		t.Errorf("enable_auth = %v, want false", answers["enable_auth"])
	}
}

func TestLoadAnswersFromFile_InvalidPath(t *testing.T) {
	_, err := LoadAnswersFromFile("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoadAnswersFromFile_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create invalid file
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content: :"), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	_, err := LoadAnswersFromFile(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestLoadAnswersFromFile_UnsupportedExtension(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create file with unsupported extension
	configPath := filepath.Join(tmpDir, "answers.txt")
	if err := os.WriteFile(configPath, []byte("content"), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	_, err := LoadAnswersFromFile(configPath)
	if err == nil {
		t.Error("Expected error for unsupported file extension")
	}
}

func TestCreateWorkflow_WithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "project")
	
	// Create config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `project_name: test-project
module_name: github.com/test/test-project
port: 8080
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load answers
	answers, err := LoadAnswersFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load answers: %v", err)
	}

	// Execute workflow
	workflow := NewCreateWorkflow()
	err = workflow.ExecuteWithOptions(CreateOptions{
		RitualPath: "../../rituals/minimal",
		TargetPath: targetDir,
		Answers:    answers,
		DryRun:     false,
		InitGit:    false,
	})
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify project was created
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		t.Error("Expected project directory to be created")
	}
}

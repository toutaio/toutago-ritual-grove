package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateCommandInteractive(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test ritual
	ritualDir := filepath.Join(tmpDir, "test-ritual")
	templatesDir := filepath.Join(ritualDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	ritualYAML := `ritual:
  name: test-ritual
  version: 1.0.0
  description: Test ritual
  template_engine: go-template

questions:
  - name: app_name
    prompt: "Application name?"
    type: text
    default: "my-app"
    required: true

  - name: port
    prompt: "HTTP port?"
    type: number
    default: 8080

files:
  templates:
    - src: "test.txt.tmpl"
      dest: "test.txt"
`
	if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), []byte(ritualYAML), 0644); err != nil {
		t.Fatal(err)
	}
	
	template := "App: {{ .app_name }}, Port: {{ .port }}"
	if err := os.WriteFile(filepath.Join(templatesDir, "test.txt.tmpl"), []byte(template), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Mock answers (simulating user input)
	answers := map[string]interface{}{
		"app_name": "test-app",
		"port":     9000,
	}
	
	targetDir := filepath.Join(tmpDir, "output")
	
	// Create command handler
	handler := NewCreateHandler()
	
	err := handler.Execute(ritualDir, targetDir, answers, CreateOptions{
		SkipQuestionnaire: true, // Skip for testing
	})
	
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	
	// Verify output
	outputFile := filepath.Join(targetDir, "test.txt")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal("Output file should exist")
	}
	
	expected := "App: test-app, Port: 9000"
	if string(content) != expected {
		t.Errorf("Content = %s, want %s", string(content), expected)
	}
}

func TestCreateCommandWithDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	
	ritualDir := filepath.Join(tmpDir, "defaults-ritual")
	templatesDir := filepath.Join(ritualDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	ritualYAML := `ritual:
  name: defaults-ritual
  version: 1.0.0
  description: Test defaults
  template_engine: go-template

questions:
  - name: app_name
    prompt: "App name?"
    type: text
    default: "default-app"

  - name: enable_feature
    prompt: "Enable feature?"
    type: boolean
    default: true

files:
  templates:
    - src: "config.txt.tmpl"
      dest: "config.txt"
`
	if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), []byte(ritualYAML), 0644); err != nil {
		t.Fatal(err)
	}
	
	template := "{{ .app_name }}: {{ .enable_feature }}"
	if err := os.WriteFile(filepath.Join(templatesDir, "config.txt.tmpl"), []byte(template), 0644); err != nil {
		t.Fatal(err)
	}
	
	targetDir := filepath.Join(tmpDir, "output")
	
	handler := NewCreateHandler()
	
	// Use defaults (empty answers)
	err := handler.ExecuteWithDefaults(ritualDir, targetDir)
	if err != nil {
		t.Fatalf("ExecuteWithDefaults() error = %v", err)
	}
	
	// Verify defaults were used
	outputFile := filepath.Join(targetDir, "config.txt")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal("Output file should exist")
	}
	
	if !contains(string(content), "default-app") {
		t.Error("Should use default app name")
	}
	
	if !contains(string(content), "true") {
		t.Error("Should use default boolean value")
	}
}

func TestCreateCommandValidation(t *testing.T) {
	tmpDir := t.TempDir()
	
	handler := NewCreateHandler()
	
	t.Run("missing ritual path", func(t *testing.T) {
		err := handler.Execute("", tmpDir, nil, CreateOptions{})
		if err == nil {
			t.Error("Should error on empty ritual path")
		}
	})
	
	t.Run("missing target path", func(t *testing.T) {
		err := handler.Execute(tmpDir, "", nil, CreateOptions{})
		if err == nil {
			t.Error("Should error on empty target path")
		}
	})
	
	t.Run("invalid ritual path", func(t *testing.T) {
		err := handler.Execute("/nonexistent", tmpDir, nil, CreateOptions{})
		if err == nil {
			t.Error("Should error on nonexistent ritual")
		}
	})
}

func TestExtractDefaultAnswers(t *testing.T) {
	tmpDir := t.TempDir()
	
	ritualDir := filepath.Join(tmpDir, "extract-ritual")
	if err := os.MkdirAll(ritualDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	ritualYAML := `ritual:
  name: extract-ritual
  version: 1.0.0

questions:
  - name: question1
    type: text
    default: "value1"
  
  - name: question2
    type: number
    default: 42
  
  - name: question3
    type: boolean
    default: true
  
  - name: no_default
    type: text
`
	if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), []byte(ritualYAML), 0644); err != nil {
		t.Fatal(err)
	}
	
	handler := NewCreateHandler()
	defaults, err := handler.ExtractDefaultAnswers(ritualDir)
	if err != nil {
		t.Fatalf("ExtractDefaultAnswers() error = %v", err)
	}
	
	if defaults["question1"] != "value1" {
		t.Errorf("question1 = %v, want value1", defaults["question1"])
	}
	
	if defaults["question2"] != 42 {
		t.Errorf("question2 = %v, want 42", defaults["question2"])
	}
	
	if defaults["question3"] != true {
		t.Errorf("question3 = %v, want true", defaults["question3"])
	}
	
	if _, exists := defaults["no_default"]; exists {
		t.Error("no_default should not be in defaults")
	}
}

func TestMergeAnswersWithDefaults(t *testing.T) {
	handler := NewCreateHandler()
	
	defaults := map[string]interface{}{
		"key1": "default1",
		"key2": "default2",
		"key3": 100,
	}
	
	userAnswers := map[string]interface{}{
		"key1": "user1",
		"key4": "user4",
	}
	
	merged := handler.MergeAnswersWithDefaults(userAnswers, defaults)
	
	// User answers should override defaults
	if merged["key1"] != "user1" {
		t.Errorf("key1 = %v, want user1", merged["key1"])
	}
	
	// Defaults should be preserved
	if merged["key2"] != "default2" {
		t.Errorf("key2 = %v, want default2", merged["key2"])
	}
	
	if merged["key3"] != 100 {
		t.Errorf("key3 = %v, want 100", merged["key3"])
	}
	
	// User-only answers should be included
	if merged["key4"] != "user4" {
		t.Errorf("key4 = %v, want user4", merged["key4"])
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && anySubstringMatch(s, substr)
}

func anySubstringMatch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

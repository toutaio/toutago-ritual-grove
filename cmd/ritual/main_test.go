package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListCommand(t *testing.T) {
	// Create temp directory with test rituals
	tmpDir := t.TempDir()
	
	// Create a test ritual
	ritualDir := filepath.Join(tmpDir, "test-ritual")
	if err := os.MkdirAll(ritualDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	ritualYAML := `ritual:
  name: test-ritual
  version: 1.0.0
  description: A test ritual
  tags:
    - test
    - example
`
	if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), []byte(ritualYAML), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Run list command
	err := runListCommand([]string{tmpDir})
	
	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	
	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	// Verify
	if err != nil {
		t.Fatalf("runListCommand() error = %v", err)
	}
	
	if !strings.Contains(output, "test-ritual") {
		t.Error("Output should contain ritual name")
	}
	
	if !strings.Contains(output, "1.0.0") {
		t.Error("Output should contain version")
	}
	
	if !strings.Contains(output, "A test ritual") {
		t.Error("Output should contain description")
	}
}

func TestListCommandJSON(t *testing.T) {
	tmpDir := t.TempDir()
	
	ritualDir := filepath.Join(tmpDir, "json-ritual")
	if err := os.MkdirAll(ritualDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	ritualYAML := `ritual:
  name: json-ritual
  version: 2.0.0
  description: JSON output test
`
	if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), []byte(ritualYAML), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Run with JSON flag
	err := runListCommandJSON([]string{tmpDir})
	
	w.Close()
	os.Stdout = oldStdout
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	if err != nil {
		t.Fatalf("runListCommandJSON() error = %v", err)
	}
	
	if !strings.Contains(output, `"name"`) {
		t.Error("JSON output should contain name field")
	}
	
	if !strings.Contains(output, "json-ritual") {
		t.Error("JSON output should contain ritual name")
	}
}

func TestListCommandEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	err := runListCommand([]string{tmpDir})
	
	w.Close()
	os.Stdout = oldStdout
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	if err != nil {
		t.Fatalf("runListCommand() error = %v", err)
	}
	
	if !strings.Contains(output, "No rituals found") {
		t.Error("Should indicate no rituals found")
	}
}

func TestCreateCommand(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create source ritual
	ritualDir := filepath.Join(tmpDir, "source-ritual")
	templatesDir := filepath.Join(ritualDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	ritualYAML := `ritual:
  name: source-ritual
  version: 1.0.0
  description: Test ritual
  template_engine: go-template

files:
  templates:
    - src: "test.txt.tmpl"
      dest: "test.txt"
`
	if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), []byte(ritualYAML), 0644); err != nil {
		t.Fatal(err)
	}
	
	template := "Hello {{ .app_name }}!"
	if err := os.WriteFile(filepath.Join(templatesDir, "test.txt.tmpl"), []byte(template), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Target directory
	targetDir := filepath.Join(tmpDir, "my-project")
	
	// Mock answers
	answers := map[string]interface{}{
		"app_name":    "my-app",
		"module_name": "github.com/example/my-app",
	}
	
	// Run create command
	err := runCreateCommand(ritualDir, targetDir, answers)
	if err != nil {
		t.Fatalf("runCreateCommand() error = %v", err)
	}
	
	// Verify project was created
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		t.Error("Project directory should be created")
	}
	
	// Verify template was rendered
	testFile := filepath.Join(targetDir, "test.txt")
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Error("Template file should be created")
	} else if string(content) != "Hello my-app!" {
		t.Errorf("Template should be rendered, got: %s", string(content))
	}
}

func TestVersionCommand(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	runVersionCommand()
	
	w.Close()
	os.Stdout = oldStdout
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	if !strings.Contains(output, "Ritual Grove") {
		t.Error("Version output should contain 'Ritual Grove'")
	}
	
	if !strings.Contains(output, "v") {
		t.Error("Version output should contain version number")
	}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantJSON bool
		wantPath string
	}{
		{
			name:     "no flags",
			args:     []string{"list"},
			wantJSON: false,
			wantPath: "",
		},
		{
			name:     "json flag",
			args:     []string{"list", "--json"},
			wantJSON: true,
			wantPath: "",
		},
		{
			name:     "path flag",
			args:     []string{"list", "--path", "/custom/path"},
			wantJSON: false,
			wantPath: "/custom/path",
		},
		{
			name:     "both flags",
			args:     []string{"list", "--json", "--path", "/test"},
			wantJSON: true,
			wantPath: "/test",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := parseFlags(tt.args)
			
			if flags.JSON != tt.wantJSON {
				t.Errorf("JSON flag = %v, want %v", flags.JSON, tt.wantJSON)
			}
			
			if flags.Path != tt.wantPath {
				t.Errorf("Path flag = %v, want %v", flags.Path, tt.wantPath)
			}
		})
	}
}

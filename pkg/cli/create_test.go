package cli

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
	
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// TestCreateRitual_BasicStructure tests that basic ritual structure is created
func TestCreateRitual_BasicStructure(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	ritualName := "test-ritual"
	err := createRitual(ritualName)
	if err != nil {
		t.Fatalf("Failed to create ritual: %v", err)
	}

	// Check directories created
	expectedDirs := []string{
		ritualName,
		filepath.Join(ritualName, "templates"),
		filepath.Join(ritualName, "static"),
		filepath.Join(ritualName, "migrations"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}

	// Check ritual.yaml created
	ritualYAMLPath := filepath.Join(ritualName, "ritual.yaml")
	if _, err := os.Stat(ritualYAMLPath); os.IsNotExist(err) {
		t.Error("ritual.yaml not created")
	}

	// Check README.md created
	readmePath := filepath.Join(ritualName, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md not created")
	}
}

// TestCreateRitual_ValidYAML tests that generated ritual.yaml is valid
func TestCreateRitual_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	ritualName := "test-ritual"
	err := createRitual(ritualName)
	if err != nil {
		t.Fatalf("Failed to create ritual: %v", err)
	}

	// Load and parse ritual.yaml
	ritualYAMLPath := filepath.Join(ritualName, "ritual.yaml")
	data, err := os.ReadFile(ritualYAMLPath)
	if err != nil {
		t.Fatalf("Failed to read ritual.yaml: %v", err)
	}

	var manifest ritual.Manifest
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		t.Fatalf("Failed to parse ritual.yaml: %v", err)
	}

	// Verify required fields
	if manifest.Ritual.Name != ritualName {
		t.Errorf("Expected ritual name %s, got %s", ritualName, manifest.Ritual.Name)
	}

	if manifest.Ritual.Version == "" {
		t.Error("Ritual version is empty")
	}

	if manifest.Ritual.Description == "" {
		t.Error("Ritual description is empty")
	}

	if manifest.Compatibility.MinToutaVersion == "" {
		t.Error("Min Touta version is empty")
	}

	if len(manifest.Questions) == 0 {
		t.Error("No questions defined")
	}
}

// TestCreateRitual_TemplateExample tests template file example is created
func TestCreateRitual_TemplateExample(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	ritualName := "test-ritual"
	err := createRitualWithTemplates(ritualName, true)
	if err != nil {
		t.Fatalf("Failed to create ritual: %v", err)
	}

	// Check template example exists
	examplePath := filepath.Join(ritualName, "templates", "main.go.tmpl")
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Error("Template example not created")
	}

	// Verify template content is valid
	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "package main") {
		t.Error("Template doesn't contain package declaration")
	}

	if !contains(contentStr, "{{ .project_name }}") {
		t.Error("Template doesn't contain variable substitution")
	}
}

// TestCreateRitual_GitignoreIncluded tests .gitignore is created
func TestCreateRitual_GitignoreIncluded(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	ritualName := "test-ritual"
	err := createRitualWithTemplates(ritualName, true)
	if err != nil {
		t.Fatalf("Failed to create ritual: %v", err)
	}

	// Check .gitignore exists
	gitignorePath := filepath.Join(ritualName, "static", ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore not created")
	}

	// Verify gitignore content
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}

	contentStr := string(content)
	expectedEntries := []string{".ritual/", "*.log", "vendor/"}
	for _, entry := range expectedEntries {
		if !contains(contentStr, entry) {
			t.Errorf(".gitignore doesn't contain %s", entry)
		}
	}
}

// TestCreateRitual_MigrationExample tests migration example is created
func TestCreateRitual_MigrationExample(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	ritualName := "test-ritual"
	err := createRitualWithTemplates(ritualName, true)
	if err != nil {
		t.Fatalf("Failed to create ritual: %v", err)
	}

	// Check migration example exists
	examplePath := filepath.Join(ritualName, "migrations", "example.md")
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Error("Migration example not created")
	}

	// Verify migration example content
	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("Failed to read migration example: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "from_version") {
		t.Error("Migration example doesn't explain from_version")
	}

	if !contains(contentStr, "to_version") {
		t.Error("Migration example doesn't explain to_version")
	}
}

// TestCreateRitual_HooksDocumented tests hooks are documented
func TestCreateRitual_HooksDocumented(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	ritualName := "test-ritual"
	err := createRitualWithTemplates(ritualName, true)
	if err != nil {
		t.Fatalf("Failed to create ritual: %v", err)
	}

	// Check README documents hooks
	readmePath := filepath.Join(ritualName, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "Hooks") || !contains(contentStr, "hooks") {
		t.Error("README doesn't document hooks")
	}

	if !contains(contentStr, "pre_install") {
		t.Error("README doesn't mention pre_install hooks")
	}

	if !contains(contentStr, "post_install") {
		t.Error("README doesn't mention post_install hooks")
	}
}

// TestCreateRitual_CommentsInYAML tests ritual.yaml has helpful comments
func TestCreateRitual_CommentsInYAML(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	ritualName := "test-ritual"
	err := createRitualWithTemplates(ritualName, true)
	if err != nil {
		t.Fatalf("Failed to create ritual: %v", err)
	}

	// Read ritual.yaml
	ritualYAMLPath := filepath.Join(ritualName, "ritual.yaml")
	content, err := os.ReadFile(ritualYAMLPath)
	if err != nil {
		t.Fatalf("Failed to read ritual.yaml: %v", err)
	}

	contentStr := string(content)
	
	// Check for comments
	if !contains(contentStr, "#") {
		t.Error("ritual.yaml doesn't contain any comments")
	}

	// Should have comments for major sections
	expectedComments := []string{"questions", "files", "hooks"}
	commentCount := 0
	for _, section := range expectedComments {
		// Look for comment before section
		if contains(contentStr, "# "+section) || contains(contentStr, "# "+section) {
			commentCount++
		}
	}

	if commentCount < 2 {
		t.Error("ritual.yaml doesn't have enough helpful comments")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

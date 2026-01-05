package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocGenerator_GenerateREADME(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewDocGenerator()

	info := ProjectInfo{
		Name:        "my-app",
		Description: "A test application",
		Version:     "1.0.0",
		Author:      "Test Author",
		Database:    "postgres",
	}

	err := gen.GenerateREADME(tmpDir, info)
	if err != nil {
		t.Fatalf("GenerateREADME() error = %v", err)
	}

	// Verify README was created
	readmePath := filepath.Join(tmpDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md should be created")
	}

	// Verify content
	content, _ := os.ReadFile(readmePath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "my-app") {
		t.Error("README should contain project name")
	}

	if !strings.Contains(contentStr, "A test application") {
		t.Error("README should contain description")
	}

	if !strings.Contains(contentStr, "postgres") {
		t.Error("README should mention database")
	}
}

func TestDocGenerator_GenerateAPIDoc(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewDocGenerator()

	endpoints := []APIEndpoint{
		{
			Method:      "GET",
			Path:        "/api/users",
			Description: "List all users",
			Response:    "[]User",
		},
		{
			Method:      "POST",
			Path:        "/api/users",
			Description: "Create a user",
			Request:     "User",
			Response:    "User",
		},
	}

	err := gen.GenerateAPIDoc(tmpDir, endpoints)
	if err != nil {
		t.Fatalf("GenerateAPIDoc() error = %v", err)
	}

	// Verify API doc was created
	apiDocPath := filepath.Join(tmpDir, "docs", "API.md")
	if _, err := os.Stat(apiDocPath); os.IsNotExist(err) {
		t.Error("API.md should be created")
	}

	content, _ := os.ReadFile(apiDocPath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "GET /api/users") {
		t.Error("API doc should contain GET endpoint")
	}

	if !strings.Contains(contentStr, "POST /api/users") {
		t.Error("API doc should contain POST endpoint")
	}
}

func TestDocGenerator_GenerateChangelog(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewDocGenerator()

	version := "1.0.0"
	err := gen.GenerateChangelog(tmpDir, version)
	if err != nil {
		t.Fatalf("GenerateChangelog() error = %v", err)
	}

	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		t.Error("CHANGELOG.md should be created")
	}

	content, _ := os.ReadFile(changelogPath)
	if !strings.Contains(string(content), version) {
		t.Error("CHANGELOG should contain version")
	}
}

func TestDocGenerator_GenerateContributing(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewDocGenerator()

	err := gen.GenerateContributing(tmpDir)
	if err != nil {
		t.Fatalf("GenerateContributing() error = %v", err)
	}

	contributingPath := filepath.Join(tmpDir, "CONTRIBUTING.md")
	if _, err := os.Stat(contributingPath); os.IsNotExist(err) {
		t.Error("CONTRIBUTING.md should be created")
	}

	content, _ := os.ReadFile(contributingPath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "Contributing") {
		t.Error("CONTRIBUTING should have contributing guidelines")
	}
}

func TestDocGenerator_GenerateLicense(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewDocGenerator()

	tests := []struct {
		name        string
		licenseType string
		wantText    string
	}{
		{
			name:        "MIT",
			licenseType: "MIT",
			wantText:    "MIT License",
		},
		{
			name:        "Apache",
			licenseType: "Apache-2.0",
			wantText:    "Apache License",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subDir := filepath.Join(tmpDir, tt.name)
			os.MkdirAll(subDir, 0750)

			err := gen.GenerateLicense(subDir, tt.licenseType, "Test Author", "2026")
			if err != nil {
				t.Fatalf("GenerateLicense() error = %v", err)
			}

			licensePath := filepath.Join(subDir, "LICENSE")
			content, _ := os.ReadFile(licensePath)

			if !strings.Contains(string(content), tt.wantText) {
				t.Errorf("LICENSE should contain %s", tt.wantText)
			}

			if !strings.Contains(string(content), "Test Author") {
				t.Error("LICENSE should contain author name")
			}
		})
	}
}

func TestDocGenerator_GenerateArchitectureDoc(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewDocGenerator()

	components := []Component{
		{Name: "Router", Description: "HTTP routing"},
		{Name: "Database", Description: "Data persistence"},
		{Name: "Auth", Description: "Authentication"},
	}

	err := gen.GenerateArchitectureDoc(tmpDir, components)
	if err != nil {
		t.Fatalf("GenerateArchitectureDoc() error = %v", err)
	}

	archPath := filepath.Join(tmpDir, "docs", "ARCHITECTURE.md")
	if _, err := os.Stat(archPath); os.IsNotExist(err) {
		t.Error("ARCHITECTURE.md should be created")
	}

	content, _ := os.ReadFile(archPath)
	contentStr := string(content)

	for _, comp := range components {
		if !strings.Contains(contentStr, comp.Name) {
			t.Errorf("Architecture doc should contain %s", comp.Name)
		}
	}
}

func TestDocGenerator_GenerateDeploymentGuide(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewDocGenerator()

	config := DeploymentConfig{
		Platform:    "docker",
		Database:    "postgres",
		Environment: "production",
	}

	err := gen.GenerateDeploymentGuide(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateDeploymentGuide() error = %v", err)
	}

	deployPath := filepath.Join(tmpDir, "docs", "DEPLOYMENT.md")
	if _, err := os.Stat(deployPath); os.IsNotExist(err) {
		t.Error("DEPLOYMENT.md should be created")
	}

	content, _ := os.ReadFile(deployPath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "docker") {
		t.Error("Deployment guide should mention platform")
	}

	if !strings.Contains(contentStr, "postgres") {
		t.Error("Deployment guide should mention database")
	}
}

func TestDocGenerator_GenerateAllDocs(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewDocGenerator()

	config := DocConfig{
		ProjectInfo: ProjectInfo{
			Name:        "test-project",
			Description: "Test description",
			Version:     "1.0.0",
			Author:      "Test",
		},
		GenerateAPI:          true,
		GenerateArchitecture: true,
		GenerateDeployment:   true,
		GenerateChangelog:    true,
		GenerateContributing: true,
		License:              "MIT",
	}

	err := gen.GenerateAll(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateAll() error = %v", err)
	}

	// Verify all docs were created
	expectedFiles := []string{
		"README.md",
		"CHANGELOG.md",
		"CONTRIBUTING.md",
		"LICENSE",
		"docs/API.md",
		"docs/ARCHITECTURE.md",
		"docs/DEPLOYMENT.md",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("File %s should be created", file)
		}
	}
}

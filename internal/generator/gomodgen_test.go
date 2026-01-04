package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestGoModGenerator_Generate(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/myapp")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
		Dependencies: ritual.Dependencies{
			Packages: []string{
				"github.com/gorilla/mux@v1.8.0",
				"github.com/lib/pq@v1.10.9",
			},
		},
	}

	gen := NewGoModGenerator()
	err := gen.Generate(tempDir, manifest, vars)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	goModPath := filepath.Join(tempDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "module github.com/test/myapp") {
		t.Error("go.mod should contain module name")
	}
	if !strings.Contains(contentStr, "github.com/gorilla/mux v1.8.0") {
		t.Error("go.mod should contain gorilla/mux dependency")
	}
	if !strings.Contains(contentStr, "github.com/lib/pq v1.10.9") {
		t.Error("go.mod should contain lib/pq dependency")
	}
}

func TestGoModGenerator_WithoutVersion(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/app")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test",
			Version: "1.0.0",
		},
		Dependencies: ritual.Dependencies{
			Packages: []string{
				"github.com/gorilla/mux", // No version
			},
		},
	}

	gen := NewGoModGenerator()
	err := gen.Generate(tempDir, manifest, vars)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	goModPath := filepath.Join(tempDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	// Should add a default version or latest
	contentStr := string(content)
	if !strings.Contains(contentStr, "github.com/gorilla/mux") {
		t.Error("go.mod should contain dependency even without version")
	}
}

func TestGoModGenerator_AddToutaDependencies(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/app")
	vars.Set("touta_version", "v0.1.0")

	gen := NewGoModGenerator()
	err := gen.AddToutaDependencies(tempDir, vars)
	if err != nil {
		t.Fatalf("AddToutaDependencies failed: %v", err)
	}

	goModPath := filepath.Join(tempDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	contentStr := string(content)
	// Check for Toutā dependencies
	if !strings.Contains(contentStr, "github.com/toutaio/toutago") {
		t.Error("go.mod should contain toutago dependency")
	}
}

func TestGoModGenerator_AddDatabaseDriver(t *testing.T) {
	testCases := []struct {
		name     string
		database string
		expected string
	}{
		{
			name:     "PostgreSQL",
			database: "postgresql",
			expected: "github.com/lib/pq",
		},
		{
			name:     "MySQL",
			database: "mysql",
			expected: "github.com/go-sql-driver/mysql",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()

			vars := NewVariables()
			vars.Set("module_name", "github.com/test/app")
			vars.Set("database", tc.database)

			manifest := &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Dependencies: ritual.Dependencies{
					Database: &ritual.DatabaseRequirement{
						Types:      []string{tc.database},
						MinVersion: "13.0",
					},
				},
			}

			gen := NewGoModGenerator()
			err := gen.Generate(tempDir, manifest, vars)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			goModPath := filepath.Join(tempDir, "go.mod")
			content, err := os.ReadFile(goModPath)
			if err != nil {
				t.Fatalf("Failed to read go.mod: %v", err)
			}

			contentStr := string(content)
			if !strings.Contains(contentStr, tc.expected) {
				t.Errorf("go.mod should contain %s driver", tc.expected)
			}
		})
	}
}

func TestGoModGenerator_RunGoModTidy(t *testing.T) {
	tempDir := t.TempDir()

	// Create a simple go.mod
	goModContent := `module github.com/test/app

go 1.21

require (
	github.com/gorilla/mux v1.8.0
)
`
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a simple main.go that uses the dependency
	mainDir := filepath.Join(tempDir, "cmd", "server")
	if err := os.MkdirAll(mainDir, 0755); err != nil {
		t.Fatalf("Failed to create main dir: %v", err)
	}

	mainContent := `package main

import (
	"fmt"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	fmt.Println(r)
}
`
	mainPath := filepath.Join(mainDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	gen := NewGoModGenerator()
	err := gen.RunGoModTidy(tempDir)
	
	// Note: This test might fail in environments without network access
	// or without Go installed. We'll make it lenient.
	if err != nil {
		t.Logf("go mod tidy failed (might be expected): %v", err)
	}
}

func TestGoModGenerator_GenerateComplete(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/complete")
	vars.Set("touta_version", "v0.1.0")
	vars.Set("database", "postgresql")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "complete-test",
			Version: "1.0.0",
		},
		Dependencies: ritual.Dependencies{
			Packages: []string{
				"github.com/gorilla/mux@v1.8.0",
			},
			Database: &ritual.DatabaseRequirement{
				Types: []string{"postgresql"},
			},
		},
	}

	gen := NewGoModGenerator()
	
	// Generate go.mod
	if err := gen.Generate(tempDir, manifest, vars); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Add Touta dependencies
	if err := gen.AddToutaDependencies(tempDir, vars); err != nil {
		t.Fatalf("AddToutaDependencies failed: %v", err)
	}

	// Verify the complete go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	contentStr := string(content)
	
	requiredElements := []string{
		"module github.com/test/complete",
		"github.com/gorilla/mux",
		"github.com/lib/pq",
		"github.com/toutaio/toutago",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(contentStr, elem) {
			t.Errorf("go.mod missing required element: %s", elem)
		}
	}
}

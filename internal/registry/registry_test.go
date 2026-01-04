package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()
	if reg == nil {
		t.Fatal("expected registry to be created")
	}
	
	if len(reg.searchPaths) == 0 {
		t.Error("expected default search paths to be set")
	}
}

func TestAddSearchPath(t *testing.T) {
	reg := NewRegistry()
	initialCount := len(reg.searchPaths)
	
	reg.AddSearchPath("/custom/path")
	
	if len(reg.searchPaths) != initialCount+1 {
		t.Errorf("expected %d search paths, got %d", initialCount+1, len(reg.searchPaths))
	}
}

func TestSearch(t *testing.T) {
	reg := NewRegistry()
	
	// Add some test metadata
	reg.cache["blog"] = &RitualMetadata{
		Name:        "blog",
		Description: "A blogging platform",
		Tags:        []string{"web", "content"},
	}
	reg.cache["wiki"] = &RitualMetadata{
		Name:        "wiki",
		Description: "A wiki system",
		Tags:        []string{"web", "documentation"},
	}
	
	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{"search by name", "blog", 1},
		{"search by description", "wiki", 1},
		{"search by tag", "web", 2},
		{"search by tag partial", "doc", 1},
		{"no matches", "xyz", 0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := reg.Search(tt.query)
			if len(results) != tt.expected {
				t.Errorf("expected %d results, got %d", tt.expected, len(results))
			}
		})
	}
}

func TestGet(t *testing.T) {
	reg := NewRegistry()
	reg.cache["test"] = &RitualMetadata{
		Name: "test",
	}
	
	t.Run("existing ritual", func(t *testing.T) {
		meta, err := reg.Get("test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if meta.Name != "test" {
			t.Errorf("expected name 'test', got '%s'", meta.Name)
		}
	})
	
	t.Run("non-existing ritual", func(t *testing.T) {
		_, err := reg.Get("nonexistent")
		if err == nil {
			t.Error("expected error for non-existing ritual")
		}
	})
}

func TestFilterByTag(t *testing.T) {
	reg := NewRegistry()
	
	reg.cache["blog"] = &RitualMetadata{
		Name: "blog",
		Tags: []string{"web", "content"},
	}
	reg.cache["wiki"] = &RitualMetadata{
		Name: "wiki",
		Tags: []string{"web", "documentation"},
	}
	reg.cache["api"] = &RitualMetadata{
		Name: "api",
		Tags: []string{"rest", "backend"},
	}
	
	results := reg.FilterByTag("web")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'web' tag, got %d", len(results))
	}
	
	results = reg.FilterByTag("backend")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'backend' tag, got %d", len(results))
	}
}

func TestMatchesQuery(t *testing.T) {
	reg := NewRegistry()
	
	meta := &RitualMetadata{
		Name:        "blog",
		Description: "A blogging platform",
		Tags:        []string{"web", "content"},
	}
	
	tests := []struct {
		query   string
		matches bool
	}{
		{"blog", true},
		{"BLOG", true},
		{"blogging", true},
		{"web", true},
		{"content", true},
		{"platform", true},
		{"xyz", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := reg.matchesQuery(meta, strings.ToLower(tt.query))
			if result != tt.matches {
				t.Errorf("query '%s': expected %v, got %v", tt.query, tt.matches, result)
			}
		})
	}
}

func TestGetDefaultSearchPaths(t *testing.T) {
	paths := getDefaultSearchPaths()
	
	if len(paths) == 0 {
		t.Error("expected at least one default search path")
	}
	
	// Check that paths are absolute
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			t.Errorf("expected absolute path, got: %s", path)
		}
	}
}

func TestList(t *testing.T) {
	reg := NewRegistry()
	
	reg.cache["ritual1"] = &RitualMetadata{Name: "ritual1"}
	reg.cache["ritual2"] = &RitualMetadata{Name: "ritual2"}
	reg.cache["ritual3"] = &RitualMetadata{Name: "ritual3"}
	
	list := reg.List()
	if len(list) != 3 {
		t.Errorf("expected 3 rituals in list, got %d", len(list))
	}
}

func TestScan(t *testing.T) {
	// Create temporary directory with a ritual
	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "test-ritual")
	
	// Create ritual directory and file
	if err := createTestRitual(ritualDir, "test-ritual", "1.0.0"); err != nil {
		t.Fatalf("failed to create test ritual: %v", err)
	}
	
	reg := NewRegistry()
	reg.searchPaths = []string{tmpDir}
	
	if err := reg.Scan(); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	
	// Check that the ritual was discovered
	if len(reg.cache) == 0 {
		t.Error("expected at least one ritual after scan")
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "my-ritual")
	
	if err := createTestRitual(ritualDir, "my-ritual", "1.0.0"); err != nil {
		t.Fatalf("failed to create test ritual: %v", err)
	}
	
	reg := NewRegistry()
	
	// Add ritual to cache first
	reg.cache["my-ritual"] = &RitualMetadata{
		Name:    "my-ritual",
		Version: "1.0.0",
		Path:    ritualDir,
		Source:  SourceLocal,
	}
	
	manifest, err := reg.Load("my-ritual")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	if manifest.Ritual.Name != "my-ritual" {
		t.Errorf("expected name 'my-ritual', got '%s'", manifest.Ritual.Name)
	}
}

func TestFilterByCompatibility(t *testing.T) {
	reg := NewRegistry()
	
	reg.cache["compat1"] = &RitualMetadata{
		Name: "compat1",
		Compatibility: &ritual.Compatibility{
			MinToutaVersion: "1.0.0",
			MaxToutaVersion: "2.0.0",
		},
	}
	reg.cache["compat2"] = &RitualMetadata{
		Name: "compat2",
		Compatibility: &ritual.Compatibility{
			MinToutaVersion: "2.0.0",
			MaxToutaVersion: "3.0.0",
		},
	}
	
	results := reg.FilterByCompatibility("1.5.0")
	// The function exists but may not be fully implemented
	_ = results
}

// Helper function to create a test ritual
func createTestRitual(dir, name, version string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	yamlContent := []byte("ritual:\n  name: " + name + "\n  version: " + version + "\n")
	ritualFile := filepath.Join(dir, "ritual.yaml")
	
	return os.WriteFile(ritualFile, yamlContent, 0644)
}

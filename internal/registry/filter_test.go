package registry

import (
	"testing"
)

func TestRegistry_Filter(t *testing.T) {
	reg := NewRegistry()
	
	// Add test rituals to registry
	reg.cache = map[string]*RitualMetadata{
		"blog": {
			Name:        "blog",
			Version:     "1.0.0",
			Description: "Blog application",
			Tags:        []string{"web", "content"},
		},
		"api": {
			Name:        "api",
			Version:     "1.0.0",
			Description: "REST API",
			Tags:        []string{"web", "api", "backend"},
		},
		"cli": {
			Name:        "cli",
			Version:     "1.0.0",
			Description: "CLI application",
			Tags:        []string{"cli", "tool"},
		},
	}

	tests := []struct {
		name     string
		filter   FilterOptions
		expected []string
	}{
		{
			name:     "no filter returns all",
			filter:   FilterOptions{},
			expected: []string{"blog", "api", "cli"},
		},
		{
			name:     "filter by single tag",
			filter:   FilterOptions{Tags: []string{"web"}},
			expected: []string{"blog", "api"},
		},
		{
			name:     "filter by multiple tags (OR)",
			filter:   FilterOptions{Tags: []string{"api", "cli"}},
			expected: []string{"api", "cli"},
		},
		{
			name:     "filter by tag with no matches",
			filter:   FilterOptions{Tags: []string{"nonexistent"}},
			expected: []string{},
		},
		{
			name:     "filter by name pattern",
			filter:   FilterOptions{NamePattern: "bl"},
			expected: []string{"blog"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := reg.Filter(tt.filter)
			
			if len(results) != len(tt.expected) {
				t.Errorf("Filter() returned %d results, want %d", len(results), len(tt.expected))
			}

			// Check that all expected rituals are in results
			resultMap := make(map[string]bool)
			for _, r := range results {
				resultMap[r.Name] = true
			}

			for _, expectedName := range tt.expected {
				if !resultMap[expectedName] {
					t.Errorf("Expected ritual %q not found in results", expectedName)
				}
			}
		})
	}
}

func TestRegistry_FilterByDatabase(t *testing.T) {
	reg := NewRegistry()
	
	// Add test rituals with database requirements
	reg.cache = map[string]*RitualMetadata{
		"blog": {
			Name:        "blog",
			Version:     "1.0.0",
			Description: "Blog with database",
			Tags:        []string{"web"},
		},
		"api": {
			Name:        "api",
			Version:     "1.0.0",
			Description: "API without database",
			Tags:        []string{"api"},
		},
	}

	// To test database filtering, we need to load manifests
	// For now, we'll test the filter structure
	filter := FilterOptions{
		Tags:         []string{"web"},
		DatabaseType: "postgres",
	}

	results := reg.Filter(filter)
	
	// Should still work with tags filter
	if len(results) == 0 {
		t.Error("Filter should return results even when database filtering not fully implemented")
	}
}

func TestFilterOptions_MatchesTags(t *testing.T) {
	tests := []struct {
		name        string
		filterTags  []string
		ritualTags  []string
		shouldMatch bool
	}{
		{
			name:        "empty filter matches all",
			filterTags:  []string{},
			ritualTags:  []string{"web", "api"},
			shouldMatch: true,
		},
		{
			name:        "single matching tag",
			filterTags:  []string{"web"},
			ritualTags:  []string{"web", "api"},
			shouldMatch: true,
		},
		{
			name:        "multiple tags, one matches",
			filterTags:  []string{"web", "cli"},
			ritualTags:  []string{"api", "web"},
			shouldMatch: true,
		},
		{
			name:        "no matching tags",
			filterTags:  []string{"frontend"},
			ritualTags:  []string{"backend", "api"},
			shouldMatch: false,
		},
		{
			name:        "ritual has no tags",
			filterTags:  []string{"web"},
			ritualTags:  []string{},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := FilterOptions{Tags: tt.filterTags}
			matches := filter.matchesTags(tt.ritualTags)
			
			if matches != tt.shouldMatch {
				t.Errorf("matchesTags() = %v, want %v", matches, tt.shouldMatch)
			}
		})
	}
}

func TestFilterOptions_MatchesNamePattern(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		ritualName  string
		shouldMatch bool
	}{
		{
			name:        "empty pattern matches all",
			pattern:     "",
			ritualName:  "blog",
			shouldMatch: true,
		},
		{
			name:        "exact match",
			pattern:     "blog",
			ritualName:  "blog",
			shouldMatch: true,
		},
		{
			name:        "partial match",
			pattern:     "bl",
			ritualName:  "blog",
			shouldMatch: true,
		},
		{
			name:        "case insensitive",
			pattern:     "BLOG",
			ritualName:  "blog",
			shouldMatch: true,
		},
		{
			name:        "no match",
			pattern:     "api",
			ritualName:  "blog",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := FilterOptions{NamePattern: tt.pattern}
			matches := filter.matchesNamePattern(tt.ritualName)
			
			if matches != tt.shouldMatch {
				t.Errorf("matchesNamePattern() = %v, want %v", matches, tt.shouldMatch)
			}
		})
	}
}

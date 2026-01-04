package executor

import (
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestDependencyGraph_DetectCycles(t *testing.T) {
	tests := []struct {
		name      string
		nodes     map[string][]string
		wantError bool
	}{
		{
			name: "no cycles",
			nodes: map[string][]string{
				"A": {"B", "C"},
				"B": {"D"},
				"C": {"D"},
				"D": {},
			},
			wantError: false,
		},
		{
			name: "simple cycle",
			nodes: map[string][]string{
				"A": {"B"},
				"B": {"A"},
			},
			wantError: true,
		},
		{
			name: "complex cycle",
			nodes: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {"D"},
				"D": {"B"},
			},
			wantError: true,
		},
		{
			name: "self dependency",
			nodes: map[string][]string{
				"A": {"A"},
			},
			wantError: true,
		},
		{
			name: "no dependencies",
			nodes: map[string][]string{
				"A": {},
				"B": {},
				"C": {},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := NewDependencyGraph()
			for name, deps := range tt.nodes {
				graph.AddNode(name, deps)
			}

			err := graph.DetectCycles()
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDependencyGraph_TopologicalSort(t *testing.T) {
	tests := []struct {
		name      string
		nodes     map[string][]string
		wantError bool
		validate  func([]string) bool
	}{
		{
			name: "simple chain",
			nodes: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {},
			},
			wantError: false,
			validate: func(order []string) bool {
				// C should come before B, B before A
				cIdx, bIdx, aIdx := -1, -1, -1
				for i, v := range order {
					switch v {
					case "C":
						cIdx = i
					case "B":
						bIdx = i
					case "A":
						aIdx = i
					}
				}
				return cIdx < bIdx && bIdx < aIdx
			},
		},
		{
			name: "diamond dependency",
			nodes: map[string][]string{
				"A": {"B", "C"},
				"B": {"D"},
				"C": {"D"},
				"D": {},
			},
			wantError: false,
			validate: func(order []string) bool {
				// D should come before B and C, B and C before A
				dIdx, aIdx := -1, -1
				for i, v := range order {
					if v == "D" {
						dIdx = i
					}
					if v == "A" {
						aIdx = i
					}
				}
				return dIdx < aIdx && len(order) == 4
			},
		},
		{
			name: "cycle should error",
			nodes: map[string][]string{
				"A": {"B"},
				"B": {"A"},
			},
			wantError: true,
			validate:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := NewDependencyGraph()
			for name, deps := range tt.nodes {
				graph.AddNode(name, deps)
			}

			order, err := graph.TopologicalSort()
			
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validate != nil && !tt.validate(order) {
				t.Errorf("Invalid order: %v", order)
			}
		})
	}
}

func TestDependencyResolver_ResolveDependencies(t *testing.T) {
	manifest := &ritual.Manifest{
		Dependencies: ritual.Dependencies{
			Packages: []string{
				"github.com/lib/pq",
				"github.com/gorilla/mux",
			},
			Rituals: []string{
				"base-ritual",
			},
			Database: &ritual.DatabaseRequirement{
				Required: true,
				Types: []string{"postgres"},
				MinVersion: "13.0",
			},
		},
	}

	resolver := NewDependencyResolver()
	deps, err := resolver.ResolveDependencies(manifest)
	if err != nil {
		t.Fatalf("ResolveDependencies failed: %v", err)
	}

	// Should have 4 dependencies total
	if len(deps) != 4 {
		t.Errorf("Expected 4 dependencies, got %d", len(deps))
	}

	// Count by type
	typeCount := make(map[string]int)
	for _, dep := range deps {
		typeCount[dep.Type]++
	}

	if typeCount["package"] != 2 {
		t.Errorf("Expected 2 package dependencies, got %d", typeCount["package"])
	}

	if typeCount["ritual"] != 1 {
		t.Errorf("Expected 1 ritual dependency, got %d", typeCount["ritual"])
	}

	if typeCount["database"] != 1 {
		t.Errorf("Expected 1 database dependency, got %d", typeCount["database"])
	}
}

func TestDependencyResolver_ValidateDependencies(t *testing.T) {
	tests := []struct {
		name      string
		manifest  *ritual.Manifest
		wantError bool
	}{
		{
			name: "valid dependencies",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "my-ritual",
					Version: "1.0.0",
				},
				Dependencies: ritual.Dependencies{
					Rituals: []string{"base-ritual"},
				},
			},
			wantError: false,
		},
		{
			name: "no dependencies",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "standalone-ritual",
					Version: "1.0.0",
				},
				Dependencies: ritual.Dependencies{},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewDependencyResolver()
			err := resolver.ValidateDependencies(tt.manifest)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDependencyResolver_GetInstallationOrder(t *testing.T) {
	resolver := NewDependencyResolver()
	
	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-project",
			Version: "1.0.0",
		},
	}
	
	err := resolver.BuildGraph(manifest)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}
	
	order, err := resolver.GetInstallationOrder()
	if err != nil {
		t.Fatalf("GetInstallationOrder failed: %v", err)
	}
	
	// Should have the main ritual node
	if len(order) != 1 {
		t.Errorf("Expected 1 ritual, got %d", len(order))
	}
	
	if len(order) > 0 && order[0] != "test-project" {
		t.Errorf("Expected test-project, got %s", order[0])
	}
}

func TestDependencyResolver_ValidateDependenciesExtra(t *testing.T) {
	resolver := NewDependencyResolver()
	
	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-project",
			Version: "1.0.0",
		},
	}
	
	err := resolver.ValidateDependencies(manifest)
	if err != nil {
		t.Errorf("ValidateDependencies failed: %v", err)
	}
}

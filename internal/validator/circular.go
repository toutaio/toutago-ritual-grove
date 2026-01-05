package validator

import (
	"fmt"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// CircularDependencyDetector detects circular dependencies in ritual composition
type CircularDependencyDetector struct {
	manifests map[string]*ritual.Manifest
}

// NewCircularDependencyDetector creates a new detector
func NewCircularDependencyDetector(manifests map[string]*ritual.Manifest) *CircularDependencyDetector {
	return &CircularDependencyDetector{
		manifests: manifests,
	}
}

// DetectCycle detects circular dependencies starting from a ritual
// Returns the cycle path if found, or nil if no cycle exists
func (d *CircularDependencyDetector) DetectCycle(startID string) ([]string, error) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var path []string

	cycle := d.detectCycleUtil(startID, visited, recStack, &path)
	if cycle != nil {
		return cycle, fmt.Errorf("circular dependency detected: %s", strings.Join(cycle, " -> "))
	}

	return nil, nil
}

func (d *CircularDependencyDetector) detectCycleUtil(
	manifestID string,
	visited map[string]bool,
	recStack map[string]bool,
	path *[]string,
) []string {
	// Add to path
	*path = append(*path, manifestID)

	// Mark as visited and in recursion stack
	visited[manifestID] = true
	recStack[manifestID] = true

	// Get manifest
	m, exists := d.manifests[manifestID]
	if !exists {
		// Manifest not found - skip (will be caught by dependency validator)
		recStack[manifestID] = false
		*path = (*path)[:len(*path)-1]
		return nil
	}

	// Check dependencies
	for _, depID := range m.Dependencies.Rituals {
		// If dependency is in recursion stack, we found a cycle
		if recStack[depID] {
			// Find start of cycle in path
			cycleStart := -1
			for i, p := range *path {
				if p == depID {
					cycleStart = i
					break
				}
			}

			// Build cycle path
			var cycle []string
			if cycleStart >= 0 {
				cycle = append(cycle, (*path)[cycleStart:]...)
			}
			cycle = append(cycle, depID)
			return cycle
		}

		// If not visited, recurse
		if !visited[depID] {
			if cycle := d.detectCycleUtil(depID, visited, recStack, path); cycle != nil {
				return cycle
			}
		}
	}

	// Remove from recursion stack
	recStack[manifestID] = false

	// Remove from path
	*path = (*path)[:len(*path)-1]

	return nil
}

// ValidateCircularDependencies validates that a manifest has no circular dependencies
func (v *Validator) ValidateCircularDependencies(m *ritual.Manifest, context map[string]*ritual.Manifest) []error {
	var errs []error

	// If no dependencies, no cycles possible
	if len(m.Dependencies.Rituals) == 0 {
		return nil
	}

	// Build manifest map including the manifest being validated
	manifests := make(map[string]*ritual.Manifest)
	for k, val := range context {
		manifests[k] = val
	}
	manifests[m.Ritual.Name] = m

	// Create detector
	detector := NewCircularDependencyDetector(manifests)

	// Check for cycles starting from this manifest
	cycle, err := detector.DetectCycle(m.Ritual.Name)
	if err != nil {
		errs = append(errs, fmt.Errorf("circular dependency detected in ritual '%s': %s",
			m.Ritual.Name, strings.Join(cycle, " -> ")))
	}

	return errs
}

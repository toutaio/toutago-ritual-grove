package executor

import (
	"fmt"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// Dependency represents a dependency with version constraints
type Dependency struct {
	Name    string
	Version string
	Type    string // "package", "ritual", "database"
}

// DependencyGraph represents the dependency graph
type DependencyGraph struct {
	nodes map[string]*DependencyNode
}

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	Name         string
	Dependencies []string
	Visited      bool
	InProgress   bool
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*DependencyNode),
	}
}

// AddNode adds a node to the graph
func (g *DependencyGraph) AddNode(name string, dependencies []string) {
	g.nodes[name] = &DependencyNode{
		Name:         name,
		Dependencies: dependencies,
		Visited:      false,
		InProgress:   false,
	}
}

// DetectCycles detects circular dependencies
func (g *DependencyGraph) DetectCycles() error {
	// Reset visited flags
	for _, node := range g.nodes {
		node.Visited = false
		node.InProgress = false
	}

	// Check each node
	for name := range g.nodes {
		if err := g.detectCyclesRecursive(name, []string{}); err != nil {
			return err
		}
	}

	return nil
}

func (g *DependencyGraph) detectCyclesRecursive(name string, path []string) error {
	node, exists := g.nodes[name]
	if !exists {
		// Node doesn't exist, skip
		return nil
	}

	if node.InProgress {
		// Found a cycle
		cycle := append(path, name)
		return fmt.Errorf("circular dependency detected: %v", cycle)
	}

	if node.Visited {
		// Already checked this path
		return nil
	}

	node.InProgress = true
	path = append(path, name)

	for _, dep := range node.Dependencies {
		if err := g.detectCyclesRecursive(dep, path); err != nil {
			return err
		}
	}

	node.InProgress = false
	node.Visited = true

	return nil
}

// TopologicalSort returns the installation order
func (g *DependencyGraph) TopologicalSort() ([]string, error) {
	// Check for cycles first
	if err := g.DetectCycles(); err != nil {
		return nil, err
	}

	// Reset visited flags
	for _, node := range g.nodes {
		node.Visited = false
	}

	var result []string
	for name := range g.nodes {
		if err := g.topologicalSortRecursive(name, &result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (g *DependencyGraph) topologicalSortRecursive(name string, result *[]string) error {
	node, exists := g.nodes[name]
	if !exists {
		// Node doesn't exist, skip
		return nil
	}

	if node.Visited {
		return nil
	}

	node.Visited = true

	// Visit dependencies first
	for _, dep := range node.Dependencies {
		if err := g.topologicalSortRecursive(dep, result); err != nil {
			return err
		}
	}

	// Add this node to result
	*result = append(*result, name)

	return nil
}

// DependencyResolver resolves ritual dependencies
type DependencyResolver struct {
	graph *DependencyGraph
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver() *DependencyResolver {
	return &DependencyResolver{
		graph: NewDependencyGraph(),
	}
}

// ResolveDependencies resolves dependencies from a manifest
func (r *DependencyResolver) ResolveDependencies(manifest *ritual.Manifest) ([]Dependency, error) {
	var deps []Dependency

	// Add Go package dependencies
	for _, pkg := range manifest.Dependencies.Packages {
		deps = append(deps, Dependency{
			Name:    pkg,
			Version: "", // Version not specified in manifest
			Type:    "package",
		})
	}

	// Add ritual dependencies
	for _, ritualName := range manifest.Dependencies.Rituals {
		deps = append(deps, Dependency{
			Name:    ritualName,
			Version: "", // Version not specified in manifest
			Type:    "ritual",
		})
	}

	// Add database dependencies
	if manifest.Dependencies.Database != nil && manifest.Dependencies.Database.Required {
		for _, dbType := range manifest.Dependencies.Database.Types {
			deps = append(deps, Dependency{
				Name:    dbType,
				Version: manifest.Dependencies.Database.MinVersion,
				Type:    "database",
			})
		}
	}

	return deps, nil
}

// BuildGraph builds a dependency graph from manifest
func (r *DependencyResolver) BuildGraph(manifest *ritual.Manifest) error {
	// Add main ritual node
	var ritualDeps []string
	for _, dep := range manifest.Dependencies.Rituals {
		ritualDeps = append(ritualDeps, dep)
	}

	r.graph.AddNode(manifest.Ritual.Name, ritualDeps)
	return nil
}

// GetInstallationOrder returns the order in which rituals should be installed
func (r *DependencyResolver) GetInstallationOrder() ([]string, error) {
	return r.graph.TopologicalSort()
}

// ValidateDependencies validates that all dependencies are satisfied
func (r *DependencyResolver) ValidateDependencies(manifest *ritual.Manifest) error {
	// Build graph
	if err := r.BuildGraph(manifest); err != nil {
		return err
	}

	// Check for cycles
	if err := r.graph.DetectCycles(); err != nil {
		return err
	}

	return nil
}

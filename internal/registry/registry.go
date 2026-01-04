// Package registry provides ritual discovery and management capabilities.
package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// Source represents where a ritual can be loaded from
type Source string

const (
	SourceLocal   Source = "local"
	SourceGit     Source = "git"
	SourceTarball Source = "tarball"
)

// RitualMetadata contains metadata about an available ritual
type RitualMetadata struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Author       string   `json:"author"`
	Tags         []string `json:"tags"`
	Source       Source   `json:"source"`
	Path         string   `json:"path"`
	URL          string   `json:"url,omitempty"`
	Compatibility *ritual.Compatibility `json:"compatibility,omitempty"`
}

// Registry manages ritual discovery and loading
type Registry struct {
	searchPaths []string
	cache       map[string]*RitualMetadata
}

// NewRegistry creates a new ritual registry
func NewRegistry() *Registry {
	return &Registry{
		searchPaths: getDefaultSearchPaths(),
		cache:       make(map[string]*RitualMetadata),
	}
}

// AddSearchPath adds a directory to the registry search paths
func (r *Registry) AddSearchPath(path string) {
	r.searchPaths = append(r.searchPaths, path)
}

// Scan discovers all available rituals in search paths
func (r *Registry) Scan() error {
	r.cache = make(map[string]*RitualMetadata)
	
	for _, searchPath := range r.searchPaths {
		if err := r.scanDirectory(searchPath); err != nil {
			// Log error but continue scanning
			fmt.Fprintf(os.Stderr, "Warning: failed to scan %s: %v\n", searchPath, err)
		}
	}
	
	return nil
}

// scanDirectory scans a directory for rituals
func (r *Registry) scanDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, skip
	}
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		ritualPath := filepath.Join(dir, entry.Name())
		ritualFile := filepath.Join(ritualPath, "ritual.yaml")
		
		if _, err := os.Stat(ritualFile); err == nil {
			if err := r.loadRitual(ritualPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load ritual at %s: %v\n", ritualPath, err)
			}
		}
	}
	
	return nil
}

// loadRitual loads a ritual's metadata
func (r *Registry) loadRitual(path string) error {
	loader := ritual.NewLoader(path)
	manifest, err := loader.Load(path)
	if err != nil {
		return err
	}
	
	metadata := &RitualMetadata{
		Name:          manifest.Ritual.Name,
		Version:       manifest.Ritual.Version,
		Description:   manifest.Ritual.Description,
		Author:        manifest.Ritual.Author,
		Tags:          manifest.Ritual.Tags,
		Source:        SourceLocal,
		Path:          path,
		Compatibility: &manifest.Compatibility,
	}
	
	r.cache[manifest.Ritual.Name] = metadata
	return nil
}

// List returns all available rituals
func (r *Registry) List() []*RitualMetadata {
	var result []*RitualMetadata
	for _, meta := range r.cache {
		result = append(result, meta)
	}
	return result
}

// Search searches for rituals matching criteria
func (r *Registry) Search(query string) []*RitualMetadata {
	var result []*RitualMetadata
	query = strings.ToLower(query)
	
	for _, meta := range r.cache {
		if r.matchesQuery(meta, query) {
			result = append(result, meta)
		}
	}
	
	return result
}

// matchesQuery checks if metadata matches search query
func (r *Registry) matchesQuery(meta *RitualMetadata, query string) bool {
	query = strings.ToLower(query)
	
	// Check name
	if strings.Contains(strings.ToLower(meta.Name), query) {
		return true
	}
	
	// Check description
	if strings.Contains(strings.ToLower(meta.Description), query) {
		return true
	}
	
	// Check tags
	for _, tag := range meta.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	
	return false
}

// Get retrieves metadata for a specific ritual
func (r *Registry) Get(name string) (*RitualMetadata, error) {
	meta, exists := r.cache[name]
	if !exists {
		return nil, fmt.Errorf("ritual not found: %s", name)
	}
	return meta, nil
}

// Load loads a ritual by name
func (r *Registry) Load(name string) (*ritual.Manifest, error) {
	meta, err := r.Get(name)
	if err != nil {
		return nil, err
	}
	
	loader := ritual.NewLoader(meta.Path)
	return loader.Load(meta.Path)
}

// FilterByTag filters rituals by tag
func (r *Registry) FilterByTag(tag string) []*RitualMetadata {
	var result []*RitualMetadata
	tag = strings.ToLower(tag)
	
	for _, meta := range r.cache {
		for _, t := range meta.Tags {
			if strings.ToLower(t) == tag {
				result = append(result, meta)
				break
			}
		}
	}
	
	return result
}

// FilterByCompatibility filters rituals compatible with a specific Toutā version
func (r *Registry) FilterByCompatibility(version string) []*RitualMetadata {
	var result []*RitualMetadata
	
	for _, meta := range r.cache {
		if meta.Compatibility == nil {
			result = append(result, meta) // No restrictions
			continue
		}
		
		// TODO: Implement semantic version comparison
		// For now, just include all
		result = append(result, meta)
	}
	
	return result
}

// getDefaultSearchPaths returns default search paths for rituals
func getDefaultSearchPaths() []string {
	paths := []string{}
	
	// User's home directory
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".toutago", "rituals"))
	}
	
	// Current directory rituals folder
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, "rituals"))
	}
	
	return paths
}

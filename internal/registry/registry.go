package registry

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// Source represents where a ritual comes from
type Source string

const (
	SourceLocal   Source = "local"
	SourceGit     Source = "git"
	SourceTarball Source = "tarball"
	SourceBuiltin Source = "builtin"
)

// RitualMetadata contains information about an available ritual
type RitualMetadata struct {
	Name          string
	Version       string
	Description   string
	Author        string
	Tags          []string
	Path          string
	Source        Source
	Compatibility *ritual.Compatibility
}

// Registry manages ritual discovery and loading
type Registry struct {
	searchPaths []string
	cache       map[string]*RitualMetadata
	cacheDir    string
}

// NewRegistry creates a new ritual registry
func NewRegistry() *Registry {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".toutago", "ritual-cache")

	return &Registry{
		searchPaths: getDefaultSearchPaths(),
		cache:       make(map[string]*RitualMetadata),
		cacheDir:    cacheDir,
	}
}

// AddSearchPath adds a directory to search for rituals
func (r *Registry) AddSearchPath(path string) {
	r.searchPaths = append(r.searchPaths, path)
}

// Scan discovers all available rituals in search paths
func (r *Registry) Scan() error {
	// Ensure cache directory exists
	if err := os.MkdirAll(r.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	for _, searchPath := range r.searchPaths {
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue // Skip non-existent paths
		}

		// Scan for ritual directories
		entries, err := os.ReadDir(searchPath)
		if err != nil {
			continue // Skip directories we can't read
		}

		for _, entry := range entries {
			entryPath := filepath.Join(searchPath, entry.Name())

			if entry.IsDir() {
				// Check if this directory contains a ritual.yaml
				ritualFile := filepath.Join(entryPath, "ritual.yaml")
				if _, err := os.Stat(ritualFile); err == nil {
					if err := r.indexRitual(entryPath, SourceLocal); err != nil {
						// Log but don't fail on individual ritual errors
						continue
					}
				}
			} else if strings.HasSuffix(entry.Name(), ".tar.gz") || strings.HasSuffix(entry.Name(), ".tgz") {
				// Handle tarball
				if err := r.handleTarball(entryPath); err != nil {
					// Log but don't fail
					continue
				}
			}
		}
	}

	return nil
}

// handleTarball extracts and indexes a ritual tarball
func (r *Registry) handleTarball(tarballPath string) error {
	// Extract tarball
	extractedPath, err := r.extractTarball(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to extract tarball: %w", err)
	}

	// Index the extracted ritual
	return r.indexRitual(extractedPath, SourceTarball)
}

// extractTarball extracts a tarball to the cache directory
func (r *Registry) extractTarball(tarballPath string) (string, error) {
	// Open tarball file
	file, err := os.Open(tarballPath)
	if err != nil {
		return "", fmt.Errorf("failed to open tarball: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log but don't fail on close error
		}
	}()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if err := gzReader.Close(); err != nil {
			// Log but don't fail on close error
		}
	}()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Determine extraction directory
	baseName := filepath.Base(tarballPath)
	baseName = strings.TrimSuffix(baseName, ".tar.gz")
	baseName = strings.TrimSuffix(baseName, ".tgz")
	extractDir := filepath.Join(r.cacheDir, baseName)

	// Check if already extracted
	if _, err := os.Stat(extractDir); err == nil {
		// Already extracted, check if it has ritual.yaml
		ritualFile := filepath.Join(extractDir, "ritual.yaml")
		if _, err := os.Stat(ritualFile); err == nil {
			return extractDir, nil
		}
	}

	// Create extraction directory
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create extraction directory: %w", err)
	}

	// Extract all files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}

		// Build target path
		target := filepath.Join(extractDir, header.Name)

		// Ensure target is within extractDir (prevent path traversal)
		if !strings.HasPrefix(target, extractDir) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", fmt.Errorf("failed to create parent directory: %w", err)
			}

			// Create file
			outFile, err := os.Create(target)
			if err != nil {
				return "", fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				_ = outFile.Close()
				return "", fmt.Errorf("failed to write file: %w", err)
			}

			if err := outFile.Close(); err != nil {
				return "", fmt.Errorf("failed to close file: %w", err)
			}
		}
	}

	// Check if ritual.yaml is in a subdirectory (nested tarball)
	ritualFile := filepath.Join(extractDir, "ritual.yaml")
	if _, err := os.Stat(ritualFile); os.IsNotExist(err) {
		// Look for ritual.yaml in subdirectories (one level deep)
		entries, err := os.ReadDir(extractDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					nestedPath := filepath.Join(extractDir, entry.Name(), "ritual.yaml")
					if _, err := os.Stat(nestedPath); err == nil {
						// Found in nested directory, use that as extract dir
						return filepath.Join(extractDir, entry.Name()), nil
					}
				}
			}
		}
	}

	return extractDir, nil
}

// indexRitual loads and caches metadata for a ritual
func (r *Registry) indexRitual(path string, source Source) error {
	// Load ritual manifest
	loader := ritual.NewLoader(path)
	manifest, err := loader.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load ritual: %w", err)
	}

	// Create metadata
	meta := &RitualMetadata{
		Name:          manifest.Ritual.Name,
		Version:       manifest.Ritual.Version,
		Description:   manifest.Ritual.Description,
		Author:        manifest.Ritual.Author,
		Tags:          manifest.Ritual.Tags,
		Path:          path,
		Source:        source,
		Compatibility: &manifest.Compatibility,
	}

	// Add to cache
	r.cache[meta.Name] = meta

	return nil
}

// Get retrieves metadata for a specific ritual
func (r *Registry) Get(name string) (*RitualMetadata, error) {
	meta, exists := r.cache[name]
	if !exists {
		return nil, fmt.Errorf("ritual '%s' not found", name)
	}
	return meta, nil
}

// Load loads the full ritual manifest
func (r *Registry) Load(name string) (*ritual.Manifest, error) {
	meta, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	loader := ritual.NewLoader(meta.Path)
	return loader.Load(meta.Path)
}

// List returns all available rituals
func (r *Registry) List() []*RitualMetadata {
	result := make([]*RitualMetadata, 0, len(r.cache))
	for _, meta := range r.cache {
		result = append(result, meta)
	}
	return result
}

// Search finds rituals matching a query string
func (r *Registry) Search(query string) []*RitualMetadata {
	var results []*RitualMetadata
	queryLower := strings.ToLower(query)

	for _, meta := range r.cache {
		if r.matchesQuery(meta, queryLower) {
			results = append(results, meta)
		}
	}

	return results
}

// matchesQuery checks if metadata matches a search query
func (r *Registry) matchesQuery(meta *RitualMetadata, query string) bool {
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

// FilterByTag returns rituals with a specific tag
func (r *Registry) FilterByTag(tag string) []*RitualMetadata {
	var results []*RitualMetadata
	tagLower := strings.ToLower(tag)

	for _, meta := range r.cache {
		for _, metaTag := range meta.Tags {
			if strings.ToLower(metaTag) == tagLower {
				results = append(results, meta)
				break
			}
		}
	}

	return results
}

// FilterByCompatibility returns rituals compatible with a Toutā version
func (r *Registry) FilterByCompatibility(toutaVersion string) []*RitualMetadata {
	var results []*RitualMetadata

	for _, meta := range r.cache {
		if meta.Compatibility == nil {
			// No compatibility restrictions
			results = append(results, meta)
			continue
		}

		// TODO: Implement semantic version comparison
		// For now, include all
		results = append(results, meta)
	}

	return results
}

// getDefaultSearchPaths returns the default ritual search paths
func getDefaultSearchPaths() []string {
	var paths []string

	// Built-in rituals (relative to executable or in rituals/ directory)
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		builtinPath := filepath.Join(exeDir, "rituals")
		if _, err := os.Stat(builtinPath); err == nil {
			paths = append(paths, builtinPath)
		}
	}

	// Current directory .ritual/
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, ".ritual"))
		
		// Also check for rituals/ in current directory (for development)
		ritualsPath := filepath.Join(cwd, "rituals")
		if _, err := os.Stat(ritualsPath); err == nil {
			paths = append(paths, ritualsPath)
		}
	}

	// User home directory ~/.toutago/rituals/
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".toutago", "rituals"))
	}

	return paths
}

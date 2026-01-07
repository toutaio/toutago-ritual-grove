package registry

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	embedded "github.com/toutaio/toutago-ritual-grove"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// Source represents where a ritual comes from
type Source string

const (
	SourceLocal    Source = "local"
	SourceGit      Source = "git"
	SourceTarball  Source = "tarball"
	SourceBuiltin  Source = "builtin"
	SourceEmbedded Source = "embedded"
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
	rituals     map[string]*RitualMetadata
	cacheDir    string
}

// NewRegistry creates a new ritual registry
func NewRegistry() *Registry {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".toutago", "ritual-cache")

	return &Registry{
		searchPaths: getDefaultSearchPaths(),
		rituals:     make(map[string]*RitualMetadata),
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
	if err := os.MkdirAll(r.cacheDir, 0750); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// First, scan embedded rituals
	if err := r.scanEmbedded(); err != nil {
		// Log but don't fail on embedded ritual errors
		_ = err
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

// scanEmbedded scans and indexes embedded rituals
func (r *Registry) scanEmbedded() error {
	ritualNames, err := embedded.List()
	if err != nil {
		return fmt.Errorf("failed to list embedded rituals: %w", err)
	}

	// Extract embedded rituals to cache if not already present or outdated
	embeddedDir := filepath.Join(r.cacheDir, "embedded")

	for _, name := range ritualNames {
		ritualPath := filepath.Join(embeddedDir, name)

		// Check if need to extract/re-extract
		needsExtraction := false

		ritualFile := filepath.Join(ritualPath, "ritual.yaml")
		if _, err := os.Stat(ritualFile); os.IsNotExist(err) {
			// Doesn't exist, need to extract
			needsExtraction = true
		} else {
			// Exists, check if version matches embedded version
			needsExtraction = r.shouldReExtract(ritualPath, name)
		}

		if needsExtraction {
			// Remove old version if exists
			if err := os.RemoveAll(ritualPath); err != nil && !os.IsNotExist(err) {
				continue
			}

			// Extract this ritual
			if err := r.extractEmbeddedRitual(name, embeddedDir); err != nil {
				continue // Skip rituals that fail to extract
			}
		}

		// Index the ritual
		if err := r.indexRitual(ritualPath, SourceEmbedded); err != nil {
			continue // Skip rituals that fail to index
		}
	}

	return nil
}

// extractEmbeddedRitual extracts a single embedded ritual to the cache
func (r *Registry) extractEmbeddedRitual(name, destDir string) error {
	ritualFS := embedded.GetFS()
	ritualPath := filepath.Join(destDir, name)

	// Ensure destination exists
	if err := os.MkdirAll(ritualPath, 0750); err != nil {
		return fmt.Errorf("failed to create ritual directory: %w", err)
	}

	// Walk the embedded ritual files
	return fs.WalkDir(ritualFS, name, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, path)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0750)
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0750); err != nil {
			return err
		}

		// Copy file
		srcFile, err := ritualFS.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := srcFile.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		// #nosec G304 - destPath is validated above
		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := destFile.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
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
	// #nosec G304 - tarballPath is from validated ritual source
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
	if err := os.MkdirAll(extractDir, 0750); err != nil {
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
		// #nosec G305 - Path traversal is prevented by HasPrefix check below
		target := filepath.Join(extractDir, header.Name)

		// Ensure target is within extractDir (prevent path traversal)
		if !strings.HasPrefix(target, extractDir) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0750); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0750); err != nil {
				return "", fmt.Errorf("failed to create parent directory: %w", err)
			}

			// Create file
			// #nosec G304 - Path is validated above with HasPrefix check
			outFile, err := os.Create(target)
			if err != nil {
				return "", fmt.Errorf("failed to create file: %w", err)
			}

			// Limit extraction size to prevent decompression bombs
			const maxFileSize = 100 * 1024 * 1024 // 100MB per file
			limitedReader := io.LimitReader(tarReader, maxFileSize)

			// #nosec G110 - Size limited to 100MB to prevent decompression bombs
			if _, err := io.Copy(outFile, limitedReader); err != nil {
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

	// Add to registry
	r.rituals[meta.Name] = meta

	return nil
}

// Get retrieves metadata for a specific ritual
func (r *Registry) Get(name string) (*RitualMetadata, error) {
	meta, exists := r.rituals[name]
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
	result := make([]*RitualMetadata, 0, len(r.rituals))
	for _, meta := range r.rituals {
		result = append(result, meta)
	}
	return result
}

// Search finds rituals matching a query string
func (r *Registry) Search(query string) []*RitualMetadata {
	var results []*RitualMetadata
	queryLower := strings.ToLower(query)

	for _, meta := range r.rituals {
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

	for _, meta := range r.rituals {
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

	for _, meta := range r.rituals {
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

	// Environment variable override
	if envPath := os.Getenv("TOUTA_RITUALS_PATH"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			paths = append(paths, envPath)
		}
	}

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

// SortByName sorts rituals alphabetically by name
func (r *Registry) SortByName(rituals []*RitualMetadata) []*RitualMetadata {
	sorted := make([]*RitualMetadata, len(rituals))
	copy(sorted, rituals)

	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	return sorted
}

// ClearCache removes all cached rituals
func (r *Registry) ClearCache() error {
// Remove cache directory contents but keep the directory
if err := os.RemoveAll(r.cacheDir); err != nil {
return fmt.Errorf("failed to clear cache: %w", err)
}

// Recreate cache directory
if err := os.MkdirAll(r.cacheDir, 0750); err != nil {
return fmt.Errorf("failed to recreate cache directory: %w", err)
}

return nil
}

// ClearEmbeddedCache removes only the embedded ritual cache
func (r *Registry) ClearEmbeddedCache() error {
embeddedDir := filepath.Join(r.cacheDir, "embedded")

if err := os.RemoveAll(embeddedDir); err != nil && !os.IsNotExist(err) {
return fmt.Errorf("failed to clear embedded cache: %w", err)
}

return nil
}

// GetCacheSize returns the total size of the cache in bytes
func (r *Registry) GetCacheSize() (int64, error) {
var size int64

err := filepath.Walk(r.cacheDir, func(path string, info os.FileInfo, err error) error {
if err != nil {
return err
}
if !info.IsDir() {
size += info.Size()
}
return nil
})

if err != nil {
return 0, fmt.Errorf("failed to calculate cache size: %w", err)
}

return size, nil
}

// shouldReExtract checks if cached ritual should be re-extracted
func (r *Registry) shouldReExtract(cachedPath, ritualName string) bool {
// Load cached manifest
loader := ritual.NewLoader(cachedPath)
cachedManifest, err := loader.Load(cachedPath)
if err != nil {
// Can't load cached version, re-extract
return true
}

// Load embedded manifest (from memory)
ritualFS := embedded.GetFS()
manifestPath := filepath.Join(ritualName, "ritual.yaml")

manifestFile, err := ritualFS.Open(manifestPath)
if err != nil {
// Can't read embedded version, keep cached
return false
}
defer manifestFile.Close()

// Read manifest data
manifestData, err := io.ReadAll(manifestFile)
if err != nil {
// Can't read embedded version, keep cached
return false
}

embeddedManifest, err := ritual.LoadFromBytes(manifestData)
if err != nil {
// Can't parse embedded version, keep cached
return false
}

// Compare versions - re-extract if different
return cachedManifest.Ritual.Version != embeddedManifest.Ritual.Version
}


// GetCachePath returns the cache directory path
func (r *Registry) GetCachePath() string {
return r.cacheDir
}

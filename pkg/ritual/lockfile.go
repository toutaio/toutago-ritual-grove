package ritual

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LockFile represents a ritual.lock file for dependency resolution
type LockFile struct {
	Ritual       RitualLock             `yaml:"ritual"`
	Dependencies []DependencyLock       `yaml:"dependencies,omitempty"`
	Rituals      []RitualDependencyLock `yaml:"rituals,omitempty"`
}

// RitualLock contains ritual metadata in lock file
type RitualLock struct {
	Name       string    `yaml:"name"`
	Version    string    `yaml:"version"`
	ResolvedAt time.Time `yaml:"resolved_at"`
}

// DependencyLock represents a resolved Go package dependency
type DependencyLock struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	Resolved string `yaml:"resolved"`
	Checksum string `yaml:"checksum"`
}

// RitualDependencyLock represents a resolved ritual dependency
type RitualDependencyLock struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	Source   string `yaml:"source"`
	Checksum string `yaml:"checksum"`
}

// LoadLockFile loads and parses a ritual.lock file
func LoadLockFile(path string) (*LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lock LockFile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	return &lock, nil
}

// Save writes the lock file to disk
func (l *LockFile) Save(path string) error {
	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}

// Verify checks if the lock file matches the ritual definition
func (l *LockFile) Verify(manifest *Manifest) error {
	if l.Ritual.Name != manifest.Ritual.Name {
		return fmt.Errorf("lock file name mismatch: expected %s, got %s",
			manifest.Ritual.Name, l.Ritual.Name)
	}

	if l.Ritual.Version != manifest.Ritual.Version {
		return fmt.Errorf("lock file version mismatch: expected %s, got %s",
			manifest.Ritual.Version, l.Ritual.Version)
	}

	return nil
}

// GetDependency retrieves a dependency by name
func (l *LockFile) GetDependency(name string) (DependencyLock, bool) {
	for _, dep := range l.Dependencies {
		if dep.Name == name {
			return dep, true
		}
	}
	return DependencyLock{}, false
}

// GetRitualDependency retrieves a ritual dependency by name
func (l *LockFile) GetRitualDependency(name string) (RitualDependencyLock, bool) {
	for _, ritual := range l.Rituals {
		if ritual.Name == name {
			return ritual, true
		}
	}
	return RitualDependencyLock{}, false
}

// NewLockFile creates a new lock file from a manifest
func NewLockFile(manifest *Manifest) *LockFile {
	lock := &LockFile{
		Ritual: RitualLock{
			Name:       manifest.Ritual.Name,
			Version:    manifest.Ritual.Version,
			ResolvedAt: time.Now().UTC(),
		},
		Dependencies: make([]DependencyLock, 0),
		Rituals:      make([]RitualDependencyLock, 0),
	}

	// Add package dependencies
	for _, pkg := range manifest.Dependencies.Packages {
		lock.Dependencies = append(lock.Dependencies, DependencyLock{
			Name:     pkg,
			Version:  "latest", // Will be resolved
			Resolved: "",
		})
	}

	return lock
}

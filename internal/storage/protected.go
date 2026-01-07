package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProtectedFileManager manages protected files that should not be overwritten during updates.
type ProtectedFileManager struct {
	state *State
}

// NewProtectedFileManager creates a new protected file manager.
func NewProtectedFileManager(state *State) *ProtectedFileManager {
	return &ProtectedFileManager{
		state: state,
	}
}

// IsProtected checks if a file path is protected.
// Supports both exact matches and glob patterns.
func (pm *ProtectedFileManager) IsProtected(filePath string) bool {
	for _, protected := range pm.state.ProtectedFiles {
		// Exact match
		if protected == filePath {
			return true
		}

		// Pattern match
		matched, err := filepath.Match(protected, filePath)
		if err == nil && matched {
			return true
		}

		// Also try matching basename for patterns like *.env
		matched, err = filepath.Match(protected, filepath.Base(filePath))
		if err == nil && matched {
			return true
		}
	}

	return false
}

// AddProtectedFile adds a file to the protected list if not already present.
func (pm *ProtectedFileManager) AddProtectedFile(filePath string) {
	// Check if already protected
	for _, f := range pm.state.ProtectedFiles {
		if f == filePath {
			return // Already protected
		}
	}

	pm.state.ProtectedFiles = append(pm.state.ProtectedFiles, filePath)
}

// RemoveProtectedFile removes a file from the protected list.
func (pm *ProtectedFileManager) RemoveProtectedFile(filePath string) {
	var updated []string
	for _, f := range pm.state.ProtectedFiles {
		if f != filePath {
			updated = append(updated, f)
		}
	}
	pm.state.ProtectedFiles = updated
}

// GetAllProtectedFiles returns all protected file paths.
func (pm *ProtectedFileManager) GetAllProtectedFiles() []string {
	return pm.state.ProtectedFiles
}

// LoadUserProtectedFiles loads additional protected files from .ritual/protected.txt.
// Returns the list of user-defined protected files.
func (pm *ProtectedFileManager) LoadUserProtectedFiles(projectPath string) ([]string, error) {
	protectedFile := filepath.Join(projectPath, ".ritual", "protected.txt")

	// #nosec G304 - protectedFile is constructed from validated components
	file, err := os.Open(protectedFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // No user protected files
		}
		return nil, fmt.Errorf("failed to open protected file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var userProtected []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		userProtected = append(userProtected, line)

		// Add to state if not already present
		pm.AddProtectedFile(line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read protected file: %w", err)
	}

	return userProtected, nil
}

// SaveProtectedList saves the protected file list to .ritual/protected.txt.
func (pm *ProtectedFileManager) SaveProtectedList(projectPath string) error {
	ritualDir := filepath.Join(projectPath, ".ritual")
	if err := os.MkdirAll(ritualDir, 0750); err != nil {
		return fmt.Errorf("failed to create ritual directory: %w", err)
	}

	protectedFile := filepath.Join(ritualDir, "protected.txt")
	var content strings.Builder
	content.WriteString("# Protected files - do not overwrite during updates\n")
	content.WriteString("# Supports glob patterns like *.env or config/*.yaml\n\n")

	for _, file := range pm.state.ProtectedFiles {
		content.WriteString(file)
		content.WriteString("\n")
	}

	if err := os.WriteFile(protectedFile, []byte(content.String()), 0600); err != nil {
		return fmt.Errorf("failed to write protected file: %w", err)
	}

	return nil
}

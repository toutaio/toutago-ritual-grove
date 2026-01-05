package embedded

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:rituals
var ritualsFS embed.FS

// ExtractRituals extracts embedded rituals to a destination directory
func ExtractRituals(destDir string) error {
	return fs.WalkDir(ritualsFS, "rituals", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel("rituals", path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0750)
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0750); err != nil {
			return err
		}

		// Copy file
		srcFile, err := ritualsFS.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := srcFile.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		// #nosec G304 - destPath is constructed from validated embedded resources
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

// GetFS returns the embedded filesystem containing rituals
func GetFS() fs.FS {
	sub, err := fs.Sub(ritualsFS, "rituals")
	if err != nil {
		// Should never happen with valid embed
		panic(fmt.Sprintf("failed to create sub filesystem: %v", err))
	}
	return sub
}

// HasRitual checks if a ritual exists in embedded files
func HasRitual(name string) bool {
	ritualPath := filepath.Join("rituals", name, "ritual.yaml")
	_, err := ritualsFS.Open(ritualPath)
	return err == nil
}

// List returns all embedded ritual names
func List() ([]string, error) {
	entries, err := fs.ReadDir(ritualsFS, "rituals")
	if err != nil {
		return nil, err
	}

	var rituals []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Verify it has a ritual.yaml
			ritualPath := filepath.Join("rituals", entry.Name(), "ritual.yaml")
			if _, err := ritualsFS.Open(ritualPath); err == nil {
				rituals = append(rituals, entry.Name())
			}
		}
	}

	return rituals, nil
}

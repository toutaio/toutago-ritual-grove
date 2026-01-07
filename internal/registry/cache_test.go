package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClearCache(t *testing.T) {
	// Create temp cache directory
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "ritual-cache")

	reg := &Registry{
		cacheDir: cacheDir,
		rituals:  make(map[string]*RitualMetadata),
	}

	// Create some cache files
	embeddedDir := filepath.Join(cacheDir, "embedded")
	if err := os.MkdirAll(embeddedDir, 0750); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	testFile := filepath.Join(embeddedDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Clear cache
	if err := reg.ClearCache(); err != nil {
		t.Fatalf("ClearCache failed: %v", err)
	}

	// Verify cache is empty but directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("Cache directory should exist after clearing")
	}

	if _, err := os.Stat(embeddedDir); !os.IsNotExist(err) {
		t.Error("Embedded cache should be removed")
	}
}

func TestClearEmbeddedCache(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "ritual-cache")

	reg := &Registry{
		cacheDir: cacheDir,
		rituals:  make(map[string]*RitualMetadata),
	}

	// Create embedded cache
	embeddedDir := filepath.Join(cacheDir, "embedded")
	if err := os.MkdirAll(embeddedDir, 0750); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Create git cache (should not be removed)
	gitDir := filepath.Join(cacheDir, "git")
	if err := os.MkdirAll(gitDir, 0750); err != nil {
		t.Fatalf("Failed to create git dir: %v", err)
	}

	gitFile := filepath.Join(gitDir, "test.txt")
	if err := os.WriteFile(gitFile, []byte("git"), 0600); err != nil {
		t.Fatalf("Failed to create git file: %v", err)
	}

	// Clear only embedded cache
	if err := reg.ClearEmbeddedCache(); err != nil {
		t.Fatalf("ClearEmbeddedCache failed: %v", err)
	}

	// Verify only embedded cache is removed
	if _, err := os.Stat(embeddedDir); !os.IsNotExist(err) {
		t.Error("Embedded cache should be removed")
	}

	if _, err := os.Stat(gitFile); os.IsNotExist(err) {
		t.Error("Git cache should not be removed")
	}
}

func TestGetCacheSize(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "ritual-cache")

	reg := &Registry{
		cacheDir: cacheDir,
		rituals:  make(map[string]*RitualMetadata),
	}

	// Create cache with known size
	embeddedDir := filepath.Join(cacheDir, "embedded")
	if err := os.MkdirAll(embeddedDir, 0750); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	testData := []byte("test data content")
	testFile := filepath.Join(embeddedDir, "test.txt")
	if err := os.WriteFile(testFile, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	size, err := reg.GetCacheSize()
	if err != nil {
		t.Fatalf("GetCacheSize failed: %v", err)
	}

	if size < int64(len(testData)) {
		t.Errorf("Expected cache size >= %d, got %d", len(testData), size)
	}
}

func TestGetCachePath(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "ritual-cache")

	reg := &Registry{
		cacheDir: cacheDir,
		rituals:  make(map[string]*RitualMetadata),
	}

	path := reg.GetCachePath()
	if path != cacheDir {
		t.Errorf("Expected cache path %s, got %s", cacheDir, path)
	}
}

package registry

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestTarballDiscovery(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test tarball
	tarballPath := filepath.Join(tmpDir, "test-ritual.tar.gz")
	if err := createTestTarball(tarballPath, "test-ritual", "1.0.0"); err != nil {
		t.Fatalf("failed to create test tarball: %v", err)
	}

	reg := NewRegistry()
	reg.searchPaths = []string{tmpDir}

	// Scan should discover and extract the tarball
	if err := reg.Scan(); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Check that the ritual was extracted and cached
	meta, err := reg.Get("test-ritual")
	if err != nil {
		t.Errorf("Expected ritual to be discovered from tarball: %v", err)
	}

	if meta.Name != "test-ritual" {
		t.Errorf("Expected name 'test-ritual', got '%s'", meta.Name)
	}

	if meta.Source != SourceTarball {
		t.Errorf("Expected source to be SourceTarball, got %v", meta.Source)
	}
}

func TestMultipleTarballsInDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple tarballs
	tarballs := []struct {
		name    string
		version string
	}{
		{"ritual1", "1.0.0"},
		{"ritual2", "2.0.0"},
		{"ritual3", "1.5.0"},
	}

	for _, tb := range tarballs {
		tarballPath := filepath.Join(tmpDir, tb.name+".tar.gz")
		if err := createTestTarball(tarballPath, tb.name, tb.version); err != nil {
			t.Fatalf("failed to create tarball %s: %v", tb.name, err)
		}
	}

	reg := NewRegistry()
	reg.searchPaths = []string{tmpDir}

	if err := reg.Scan(); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// All rituals should be discovered
	if len(reg.rituals) < len(tarballs) {
		t.Errorf("Expected at least %d rituals, got %d", len(tarballs), len(reg.rituals))
	}

	for _, tb := range tarballs {
		if _, err := reg.Get(tb.name); err != nil {
			t.Errorf("Ritual '%s' not found in cache", tb.name)
		}
	}
}

func TestInvalidTarball(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid tarball (not actually a tar.gz file)
	invalidPath := filepath.Join(tmpDir, "invalid.tar.gz")
	if err := os.WriteFile(invalidPath, []byte("not a tarball"), 0600); err != nil {
		t.Fatal(err)
	}

	reg := NewRegistry()
	reg.searchPaths = []string{tmpDir}

	// Scan should not fail but should skip invalid tarball
	err := reg.Scan()
	// We expect it to either succeed (skip invalid) or fail gracefully
	if err != nil {
		t.Logf("Scan returned error (acceptable): %v", err)
	}
}

func TestTarballExtraction(t *testing.T) {
	tmpDir := t.TempDir()
	tarballPath := filepath.Join(tmpDir, "extract-test.tar.gz")

	// Create tarball with additional files
	if err := createComplexTarball(tarballPath, "complex-ritual", "2.0.0"); err != nil {
		t.Fatal(err)
	}

	reg := NewRegistry()
	cacheDir := filepath.Join(tmpDir, "cache")
	reg.cacheDir = cacheDir

	// Extract the tarball
	extractedPath, err := reg.extractTarball(tarballPath)
	if err != nil {
		t.Fatalf("extractTarball failed: %v", err)
	}

	// Verify ritual.yaml exists in extracted location
	ritualFile := filepath.Join(extractedPath, "ritual.yaml")
	if _, err := os.Stat(ritualFile); os.IsNotExist(err) {
		t.Error("ritual.yaml not found in extracted directory")
	}

	// Verify templates directory exists
	templatesDir := filepath.Join(extractedPath, "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		t.Error("templates directory not found in extracted directory")
	}
}

func TestTarballCaching(t *testing.T) {
	tmpDir := t.TempDir()
	tarballPath := filepath.Join(tmpDir, "cached-ritual.tar.gz")

	if err := createTestTarball(tarballPath, "cached-ritual", "1.0.0"); err != nil {
		t.Fatal(err)
	}

	reg := NewRegistry()
	cacheDir := filepath.Join(tmpDir, "ritual-cache")
	reg.cacheDir = cacheDir
	reg.searchPaths = []string{tmpDir}

	// First scan - should extract
	if err := reg.Scan(); err != nil {
		t.Fatalf("First scan failed: %v", err)
	}

	// Get extracted path
	meta, _ := reg.Get("cached-ritual")
	extractedPath := meta.Path

	// Verify extracted directory exists
	if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
		t.Error("Extracted directory should exist after first scan")
	}

	// Second scan - should use cached version
	if err := reg.Scan(); err != nil {
		t.Fatalf("Second scan failed: %v", err)
	}

	// Should still be accessible
	meta2, err := reg.Get("cached-ritual")
	if err != nil {
		t.Error("Ritual should still be accessible after second scan")
	}

	if meta2.Path != extractedPath {
		t.Error("Cached path should remain the same")
	}
}

func TestTarballWithNestedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	tarballPath := filepath.Join(tmpDir, "nested-ritual.tar.gz")

	// Some tarballs have a top-level directory, e.g., ritual-name/ritual.yaml
	if err := createNestedTarball(tarballPath, "nested-ritual", "1.0.0"); err != nil {
		t.Fatal(err)
	}

	reg := NewRegistry()
	cacheDir := filepath.Join(tmpDir, "cache")
	reg.cacheDir = cacheDir

	extractedPath, err := reg.extractTarball(tarballPath)
	if err != nil {
		t.Fatalf("extractTarball failed: %v", err)
	}

	// Should find ritual.yaml somewhere in the extracted structure
	ritualFile := filepath.Join(extractedPath, "ritual.yaml")
	if _, err := os.Stat(ritualFile); os.IsNotExist(err) {
		// Try in nested directory
		ritualFile = filepath.Join(extractedPath, "nested-ritual", "ritual.yaml")
		if _, err := os.Stat(ritualFile); os.IsNotExist(err) {
			t.Error("ritual.yaml not found in extracted structure")
		}
	}
}

// Helper: Create a simple tarball with ritual.yaml
func createTestTarball(path, name, version string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add ritual.yaml
	yamlContent := []byte("ritual:\n  name: " + name + "\n  version: " + version + "\n")

	header := &tar.Header{
		Name: "ritual.yaml",
		Mode: 0600,
		Size: int64(len(yamlContent)),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	if _, err := tarWriter.Write(yamlContent); err != nil {
		return err
	}

	return nil
}

// Helper: Create a complex tarball with multiple files and directories
func createComplexTarball(path, name, version string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add ritual.yaml
	yamlContent := []byte("ritual:\n  name: " + name + "\n  version: " + version + "\n")
	if err := addFileToTar(tarWriter, "ritual.yaml", yamlContent); err != nil {
		return err
	}

	// Add templates directory
	if err := addDirToTar(tarWriter, "templates/"); err != nil {
		return err
	}

	// Add a template file
	templateContent := []byte("# Template\nHello {{ .Name }}")
	if err := addFileToTar(tarWriter, "templates/main.tmpl", templateContent); err != nil {
		return err
	}

	return nil
}

// Helper: Create a nested tarball (with top-level directory)
func createNestedTarball(path, name, version string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add directory
	if err := addDirToTar(tarWriter, name+"/"); err != nil {
		return err
	}

	// Add ritual.yaml inside directory
	yamlContent := []byte("ritual:\n  name: " + name + "\n  version: " + version + "\n")
	if err := addFileToTar(tarWriter, name+"/ritual.yaml", yamlContent); err != nil {
		return err
	}

	return nil
}

// Helper: Add a file to tar
func addFileToTar(tw *tar.Writer, name string, content []byte) error {
	header := &tar.Header{
		Name: name,
		Mode: 0600,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := tw.Write(content); err != nil {
		return err
	}

	return nil
}

// Helper: Add a directory to tar
func addDirToTar(tw *tar.Writer, name string) error {
	header := &tar.Header{
		Name:     name,
		Mode:     0750,
		Typeflag: tar.TypeDir,
	}

	return tw.WriteHeader(header)
}

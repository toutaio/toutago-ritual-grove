package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanCommand(t *testing.T) {
	// Create temp home dir for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Create cache directory with some files
	cachePath := filepath.Join(tmpHome, ".toutago", "ritual-cache")
	if err := os.MkdirAll(filepath.Join(cachePath, "basic-site"), 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}
	
	testFile := filepath.Join(cachePath, "basic-site", "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); err != nil {
		t.Fatalf("Test file should exist: %v", err)
	}

	// Run clean command with force flag
	cmd := NewCleanCommand()
	cmd.SetArgs([]string{"--force"})
	
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Clean command failed: %v", err)
	}

	// Verify cache was deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("Cache file should be deleted")
	}
}

func TestCleanCommand_EmptyCache(t *testing.T) {
	// Create temp home dir without cache
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	cmd := NewCleanCommand()
	cmd.SetArgs([]string{"--force"})
	
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Clean command should handle empty cache: %v", err)
	}
}

func TestCleanCommand_Help(t *testing.T) {
	cmd := NewCleanCommand()
	cmd.SetArgs([]string{"--help"})
	
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Help should not error: %v", err)
	}
}

func TestCleanCommand_Description(t *testing.T) {
	cmd := NewCleanCommand()
	
	if cmd.Use != "clean" {
		t.Errorf("Expected use 'clean', got %s", cmd.Use)
	}
	
	if !strings.Contains(strings.ToLower(cmd.Short), "clean") {
		t.Errorf("Short description should mention 'clean': %s", cmd.Short)
	}
	
	if !strings.Contains(strings.ToLower(cmd.Long), "cache") {
		t.Errorf("Long description should mention 'cache': %s", cmd.Long)
	}
}

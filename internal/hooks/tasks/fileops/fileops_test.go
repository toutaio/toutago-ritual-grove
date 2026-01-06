package fileops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

func TestMkdirTask(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test", "nested", "dir")

	task := &MkdirTask{
		Path: testDir,
		Perm: 0755,
	}

	ctx := tasks.NewTaskContext()
	ctx.SetWorkingDir(tmpDir)

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify directory was created.
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected directory")
	}
}

func TestMkdirTaskValidation(t *testing.T) {
	// Test missing path.
	task := &MkdirTask{}
	if err := task.Validate(); err == nil {
		t.Error("Expected validation error for missing path")
	}

	// Test invalid permissions.
	task = &MkdirTask{Path: "/tmp/test", Perm: 0}
	if err := task.Validate(); err == nil {
		t.Error("Expected validation error for invalid perm")
	}

	// Test valid task.
	task = &MkdirTask{Path: "/tmp/test", Perm: 0755}
	if err := task.Validate(); err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

func TestCopyTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file.
	srcFile := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(srcFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dstFile := filepath.Join(tmpDir, "dest.txt")

	task := &CopyTask{
		Src:  srcFile,
		Dest: dstFile,
	}

	ctx := tasks.NewTaskContext()

	err = task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify file was copied.
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Destination file not created: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("Expected 'test content', got '%s'", string(content))
	}
}

func TestCopyTaskDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source directory with files.
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0644)

	dstDir := filepath.Join(tmpDir, "dst")

	task := &CopyTask{
		Src:  srcDir,
		Dest: dstDir,
	}

	ctx := tasks.NewTaskContext()

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify directory was copied.
	content1, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
	if err != nil {
		t.Fatalf("File1 not copied: %v", err)
	}
	if string(content1) != "content1" {
		t.Errorf("Expected 'content1', got '%s'", string(content1))
	}

	content2, err := os.ReadFile(filepath.Join(dstDir, "subdir", "file2.txt"))
	if err != nil {
		t.Fatalf("File2 not copied: %v", err)
	}
	if string(content2) != "content2" {
		t.Errorf("Expected 'content2', got '%s'", string(content2))
	}
}

func TestMoveTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file.
	srcFile := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(srcFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dstFile := filepath.Join(tmpDir, "dest.txt")

	task := &MoveTask{
		Src:  srcFile,
		Dest: dstFile,
	}

	ctx := tasks.NewTaskContext()

	err = task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify file was moved.
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Destination file not created: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("Expected 'test content', got '%s'", string(content))
	}

	// Verify source file was removed.
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("Source file should have been removed")
	}
}

func TestRemoveTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file.
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	task := &RemoveTask{
		Path: testFile,
	}

	ctx := tasks.NewTaskContext()

	err = task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify file was removed.
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File should have been removed")
	}
}

func TestChmodTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file.
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	task := &ChmodTask{
		Path: testFile,
		Perm: 0755,
	}

	ctx := tasks.NewTaskContext()

	err = task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify permissions changed.
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Mode().Perm() != 0755 {
		t.Errorf("Expected 0755, got %04o", info.Mode().Perm())
	}
}

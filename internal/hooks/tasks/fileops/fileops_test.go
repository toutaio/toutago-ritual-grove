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

func TestTemplateRenderTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create template file.
	templatePath := filepath.Join(tmpDir, "test.tmpl")
	templateContent := "Hello, {{ .Name }}!"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	destPath := filepath.Join(tmpDir, "output.txt")

	task := &TemplateRenderTask{
		Template: templatePath,
		Dest:     destPath,
		Data:     map[string]interface{}{"Name": "World"},
	}

	ctx := tasks.NewTaskContext()
	ctx.SetWorkingDir(tmpDir)

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify output file exists and has correct content.
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	expected := "Hello, World!"
	if string(content) != expected {
		t.Errorf("Expected %q, got %q", expected, string(content))
	}
}

func TestTemplateRenderTaskValidation(t *testing.T) {
	// Test missing template.
	task := &TemplateRenderTask{Dest: "/tmp/out"}
	if err := task.Validate(); err == nil {
		t.Error("Expected validation error for missing template")
	}

	// Test missing dest.
	task = &TemplateRenderTask{Template: "/tmp/tmpl"}
	if err := task.Validate(); err == nil {
		t.Error("Expected validation error for missing dest")
	}
}

func TestValidateFilesTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files.
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	task := &ValidateFilesTask{
		Files: []string{file1, file2},
	}

	ctx := tasks.NewTaskContext()
	ctx.SetWorkingDir(tmpDir)

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test with missing file.
	task = &ValidateFilesTask{
		Files: []string{file1, filepath.Join(tmpDir, "missing.txt")},
	}

	err = task.Execute(context.Background(), ctx)
	if err == nil {
		t.Error("Expected error for missing file")
	}
}

func TestValidateFilesTaskValidation(t *testing.T) {
	// Test empty files list.
	task := &ValidateFilesTask{}
	if err := task.Validate(); err == nil {
		t.Error("Expected validation error for empty files list")
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

// Test all Name() methods
func TestTaskNames(t *testing.T) {
	tests := []struct {
		task interface{ Name() string }
		want string
	}{
		{&MkdirTask{}, "mkdir"},
		{&CopyTask{}, "copy"},
		{&MoveTask{}, "move"},
		{&RemoveTask{}, "remove"},
		{&ChmodTask{}, "chmod"},
		{&TemplateRenderTask{}, "template-render"},
		{&ValidateFilesTask{}, "validate-files"},
	}

	for _, tt := range tests {
		if got := tt.task.Name(); got != tt.want {
			t.Errorf("Name() = %v, want %v", got, tt.want)
		}
	}
}

// Test all Validate() methods
func TestCopyTaskValidation(t *testing.T) {
	tests := []struct {
		name    string
		task    *CopyTask
		wantErr bool
	}{
		{"valid", &CopyTask{Src: "/tmp/src", Dest: "/tmp/dest"}, false},
		{"missing src", &CopyTask{Dest: "/tmp/dest"}, true},
		{"missing dest", &CopyTask{Src: "/tmp/src"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMoveTaskValidation(t *testing.T) {
	tests := []struct {
		name    string
		task    *MoveTask
		wantErr bool
	}{
		{"valid", &MoveTask{Src: "/tmp/src", Dest: "/tmp/dest"}, false},
		{"missing src", &MoveTask{Dest: "/tmp/dest"}, true},
		{"missing dest", &MoveTask{Src: "/tmp/src"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveTaskValidation(t *testing.T) {
	tests := []struct {
		name    string
		task    *RemoveTask
		wantErr bool
	}{
		{"valid", &RemoveTask{Path: "/tmp/file"}, false},
		{"missing path", &RemoveTask{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChmodTaskValidation(t *testing.T) {
	tests := []struct {
		name    string
		task    *ChmodTask
		wantErr bool
	}{
		{"valid", &ChmodTask{Path: "/tmp/file", Perm: 0755}, false},
		{"missing path", &ChmodTask{Perm: 0755}, true},
		{"zero perm", &ChmodTask{Path: "/tmp/file", Perm: 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

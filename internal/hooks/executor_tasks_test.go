package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	// Import task packages to register them
	_ "github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks/dbops"
	_ "github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks/envops"
	_ "github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks/fileops"
	_ "github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks/goops"
	_ "github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks/httpops"
	_ "github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks/sysops"
	_ "github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks/validationops"
)

func TestHookExecutor_ExecuteTaskObject(t *testing.T) {
	tmpDir := t.TempDir()
	executor := NewHookExecutor(tmpDir)

	// Test task object execution (mkdir task)
	taskObj := map[string]interface{}{
		"type": "mkdir",
		"path": filepath.Join(tmpDir, "testdir"),
		"perm": float64(0755),
	}

	taskJSON, err := json.Marshal(taskObj)
	if err != nil {
		t.Fatalf("Failed to marshal task: %v", err)
	}

	// Execute task as hook
	err = executor.ExecutePostInstall([]string{string(taskJSON)})
	if err != nil {
		t.Fatalf("Failed to execute task hook: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Join(tmpDir, "testdir")); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}

func TestHookExecutor_ExecuteMixedHooks(t *testing.T) {
	tmpDir := t.TempDir()
	executor := NewHookExecutor(tmpDir)

	// Test mixing shell commands and task objects
	mkdirTask := map[string]interface{}{
		"type": "mkdir",
		"path": filepath.Join(tmpDir, "taskdir"),
		"perm": float64(0755),
	}

	mkdirJSON, _ := json.Marshal(mkdirTask)

	hooks := []string{
		// Shell command
		"echo 'Hello from shell'",
		// Task object
		string(mkdirJSON),
		// Another shell command
		"touch " + filepath.Join(tmpDir, "test.txt"),
	}

	err := executor.ExecutePostInstall(hooks)
	if err != nil {
		t.Fatalf("Failed to execute mixed hooks: %v", err)
	}

	// Verify both succeeded
	if _, err := os.Stat(filepath.Join(tmpDir, "taskdir")); os.IsNotExist(err) {
		t.Error("Expected task directory to be created")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "test.txt")); os.IsNotExist(err) {
		t.Error("Expected shell command file to be created")
	}
}

func TestHookExecutor_ExecuteGoModTidyTask(t *testing.T) {
	tmpDir := t.TempDir()
	executor := NewHookExecutor(tmpDir)

	// Create a simple go.mod file
	goMod := `module testmodule

go 1.21
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Execute go-mod-tidy task
	task := map[string]interface{}{
		"type": "go-mod-tidy",
	}
	taskJSON, _ := json.Marshal(task)

	err = executor.ExecutePostInstall([]string{string(taskJSON)})
	if err != nil {
		t.Fatalf("Failed to execute go-mod-tidy task: %v", err)
	}

	// go.mod should still exist and be valid
	content, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	if err != nil {
		t.Error("Expected go.mod to exist after tidy")
	}
	if len(content) == 0 {
		t.Error("Expected go.mod to have content")
	}
}

func TestHookExecutor_ExecuteFileOperationTasks(t *testing.T) {
	tmpDir := t.TempDir()
	executor := NewHookExecutor(tmpDir)

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(srcFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Execute copy task
	copyTask := map[string]interface{}{
		"type": "copy",
		"src":  srcFile,
		"dest": filepath.Join(tmpDir, "dest.txt"),
	}
	copyJSON, _ := json.Marshal(copyTask)

	err = executor.ExecutePostInstall([]string{string(copyJSON)})
	if err != nil {
		t.Fatalf("Failed to execute copy task: %v", err)
	}

	// Verify file was copied
	content, err := os.ReadFile(filepath.Join(tmpDir, "dest.txt"))
	if err != nil {
		t.Error("Expected destination file to exist")
	}
	if string(content) != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", string(content))
	}
}

func TestHookExecutor_TaskDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	executor := NewHookExecutor(tmpDir)
	executor.SetDryRun(true)

	// Execute mkdir task in dry run
	task := map[string]interface{}{
		"type": "mkdir",
		"path": filepath.Join(tmpDir, "dryrundir"),
		"perm": float64(0755),
	}
	taskJSON, _ := json.Marshal(task)

	err := executor.ExecutePostInstall([]string{string(taskJSON)})
	if err != nil {
		t.Fatalf("Failed to execute task in dry run: %v", err)
	}

	// Verify directory was NOT created (dry run)
	if _, err := os.Stat(filepath.Join(tmpDir, "dryrundir")); !os.IsNotExist(err) {
		t.Error("Expected directory to NOT be created in dry run mode")
	}

	// Verify output mentions dry run
	output := executor.GetOutput()
	if output == "" {
		t.Error("Expected dry run output")
	}
}

func TestHookExecutor_InvalidTaskType(t *testing.T) {
	tmpDir := t.TempDir()
	executor := NewHookExecutor(tmpDir)

	// Try to execute invalid task type
	task := map[string]interface{}{
		"type": "invalid-task-type",
	}
	taskJSON, _ := json.Marshal(task)

	err := executor.ExecutePostInstall([]string{string(taskJSON)})
	if err == nil {
		t.Error("Expected error for invalid task type")
	}
}

func TestHookExecutor_TaskExecutionContext(t *testing.T) {
	tmpDir := t.TempDir()
	executor := NewHookExecutor(tmpDir)
	
	// Set environment variable in process for test
	os.Setenv("TEST_VAR_FOR_HOOK", "test_value")
	defer os.Unsetenv("TEST_VAR_FOR_HOOK")

	// Execute env-check task to verify environment variables
	task := map[string]interface{}{
		"type":     "env-check",
		"required": []string{"TEST_VAR_FOR_HOOK"},
	}
	taskJSON, _ := json.Marshal(task)

	err := executor.ExecutePostInstall([]string{string(taskJSON)})
	if err != nil {
		t.Fatalf("Failed to execute env-check task: %v", err)
	}
}

func TestIsTaskObject(t *testing.T) {
	tests := []struct {
		name     string
		hook     string
		expected bool
	}{
		{
			name:     "JSON task object",
			hook:     `{"type": "mkdir", "path": "/tmp/test"}`,
			expected: true,
		},
		{
			name:     "Shell command",
			hook:     "go mod tidy",
			expected: false,
		},
		{
			name:     "Shell command with JSON-like content",
			hook:     `echo '{"hello": "world"}'`,
			expected: false,
		},
		{
			name:     "Empty string",
			hook:     "",
			expected: false,
		},
		{
			name:     "Whitespace",
			hook:     "   ",
			expected: false,
		},
		{
			name:     "Array JSON (not a task)",
			hook:     `["not", "a", "task"]`,
			expected: false,
		},
		{
			name:     "Task without type field",
			hook:     `{"path": "/tmp/test"}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTaskObject(tt.hook)
			if result != tt.expected {
				t.Errorf("isTaskObject(%q) = %v, expected %v", tt.hook, result, tt.expected)
			}
		})
	}
}

func TestParseTaskObject(t *testing.T) {
	tests := []struct {
		name      string
		hook      string
		wantType  string
		wantError bool
	}{
		{
			name:      "Valid mkdir task",
			hook:      `{"type": "mkdir", "path": "/tmp/test", "perm": 493}`,
			wantType:  "mkdir",
			wantError: false,
		},
		{
			name:      "Valid go-mod-tidy task",
			hook:      `{"type": "go-mod-tidy"}`,
			wantType:  "go-mod-tidy",
			wantError: false,
		},
		{
			name:      "Invalid JSON",
			hook:      `{"type": "mkdir", invalid}`,
			wantType:  "",
			wantError: true,
		},
		{
			name:      "Missing type field",
			hook:      `{"path": "/tmp/test"}`,
			wantType:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskData, err := parseTaskObject(tt.hook)
			if (err != nil) != tt.wantError {
				t.Errorf("parseTaskObject() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if taskType, ok := taskData["type"].(string); !ok || taskType != tt.wantType {
					t.Errorf("parseTaskObject() type = %v, want %v", taskType, tt.wantType)
				}
			}
		})
	}
}

func TestCreateTaskFromData(t *testing.T) {
	tests := []struct {
		name      string
		taskData  map[string]interface{}
		wantError bool
	}{
		{
			name: "Valid mkdir task",
			taskData: map[string]interface{}{
				"type": "mkdir",
				"path": "/tmp/test",
				"perm": float64(0755),
			},
			wantError: false,
		},
		{
			name: "Invalid task type",
			taskData: map[string]interface{}{
				"type": "nonexistent-task",
			},
			wantError: true,
		},
		{
			name: "Missing type field",
			taskData: map[string]interface{}{
				"path": "/tmp/test",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := createTaskFromData(tt.taskData)
			if (err != nil) != tt.wantError {
				t.Errorf("createTaskFromData() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && task == nil {
				t.Error("createTaskFromData() returned nil task without error")
			}
		})
	}
}

func TestHookExecutor_MultipleTaskTypes(t *testing.T) {
	tmpDir := t.TempDir()
	executor := NewHookExecutor(tmpDir)

	// Create a sequence of different task types
	tasks := []map[string]interface{}{
		{
			"type": "mkdir",
			"path": filepath.Join(tmpDir, "dir1"),
			"perm": float64(0755),
		},
		{
			"type": "mkdir",
			"path": filepath.Join(tmpDir, "dir2"),
			"perm": float64(0755),
		},
		{
			"type": "copy",
			"src":  "/dev/null",
			"dest": filepath.Join(tmpDir, "empty.txt"),
		},
	}

	hooks := make([]string, len(tasks))
	for i, task := range tasks {
		taskJSON, _ := json.Marshal(task)
		hooks[i] = string(taskJSON)
	}

	err := executor.ExecutePostInstall(hooks)
	if err != nil {
		t.Fatalf("Failed to execute multiple task types: %v", err)
	}

	// Verify all tasks executed
	if _, err := os.Stat(filepath.Join(tmpDir, "dir1")); os.IsNotExist(err) {
		t.Error("Expected dir1 to be created")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "dir2")); os.IsNotExist(err) {
		t.Error("Expected dir2 to be created")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "empty.txt")); os.IsNotExist(err) {
		t.Error("Expected empty.txt to be created")
	}
}

func init() {
	// Tasks are auto-registered via init() in their packages
	// Imports above ensure registration happens
}

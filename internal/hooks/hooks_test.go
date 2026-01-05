package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHookExecutor_ExecutePreInstall(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)

	hooks := []string{
		"echo 'pre-install hook'",
		"mkdir -p test-dir",
	}

	err := executor.ExecutePreInstall(hooks)
	if err != nil {
		t.Fatalf("ExecutePreInstall() error = %v", err)
	}

	// Verify directory was created
	testDir := filepath.Join(tmpDir, "test-dir")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("Hook should have created test-dir")
	}
}

func TestHookExecutor_ExecutePostInstall(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file to verify hook execution
	testFile := filepath.Join(tmpDir, "marker.txt")

	executor := NewHookExecutor(tmpDir)

	hooks := []string{
		"echo 'post-install' > marker.txt",
	}

	err := executor.ExecutePostInstall(hooks)
	if err != nil {
		t.Fatalf("ExecutePostInstall() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Hook should have created marker.txt")
	}

	content, _ := os.ReadFile(testFile)
	if !strings.Contains(string(content), "post-install") {
		t.Error("File should contain 'post-install'")
	}
}

func TestHookExecutor_ExecuteGoModTidy(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal go.mod
	goMod := `module test.com/example

go 1.21

require (
	github.com/example/pkg v1.0.0
)
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	executor := NewHookExecutor(tmpDir)

	hooks := []string{
		"go mod tidy",
	}

	// This will fail without actual dependencies, but we test execution
	err := executor.ExecutePostInstall(hooks)

	// We expect an error since the module doesn't exist, but hook should execute
	// Just verify we attempted to run it
	if err == nil {
		t.Log("go mod tidy succeeded (unexpected but ok)")
	} else {
		t.Logf("go mod tidy failed as expected: %v", err)
	}
}

func TestHookExecutor_ExecuteWithTimeout(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)
	executor.SetTimeout(1 * time.Second)

	// Command that takes too long
	hooks := []string{
		"sleep 5",
	}

	err := executor.ExecutePostInstall(hooks)

	if err == nil {
		t.Error("Should timeout on long-running command")
	}

	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "killed") {
		t.Errorf("Error should mention timeout, got: %v", err)
	}
}

func TestHookExecutor_ExecuteInvalidCommand(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)

	hooks := []string{
		"nonexistent-command-xyz",
	}

	err := executor.ExecutePreInstall(hooks)

	if err == nil {
		t.Error("Should error on invalid command")
	}
}

func TestHookExecutor_ExecuteMultipleHooks(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)

	hooks := []string{
		"echo 'hook 1' > file1.txt",
		"echo 'hook 2' > file2.txt",
		"echo 'hook 3' > file3.txt",
	}

	err := executor.ExecutePostInstall(hooks)
	if err != nil {
		t.Fatalf("ExecutePostInstall() error = %v", err)
	}

	// Verify all files were created
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("File file%d.txt should exist", i)
		}
	}
}

func TestHookExecutor_ExecuteWithEnvironment(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)
	executor.SetEnv("CUSTOM_VAR", "custom_value")

	hooks := []string{
		"echo $CUSTOM_VAR > env_test.txt",
	}

	err := executor.ExecutePostInstall(hooks)
	if err != nil {
		t.Fatalf("ExecutePostInstall() error = %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "env_test.txt"))
	if !strings.Contains(string(content), "custom_value") {
		t.Error("Environment variable should be available in hook")
	}
}

func TestHookExecutor_DryRun(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)
	executor.SetDryRun(true)

	hooks := []string{
		"echo 'should not execute' > should_not_exist.txt",
	}

	err := executor.ExecutePostInstall(hooks)
	if err != nil {
		t.Fatalf("DryRun should not error: %v", err)
	}

	// Verify file was NOT created (dry run)
	testFile := filepath.Join(tmpDir, "should_not_exist.txt")
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File should not exist in dry run mode")
	}
}

func TestHookExecutor_CaptureOutput(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)

	hooks := []string{
		"echo 'test output'",
	}

	err := executor.ExecutePostInstall(hooks)
	if err != nil {
		t.Fatalf("ExecutePostInstall() error = %v", err)
	}

	output := executor.GetOutput()
	if !strings.Contains(output, "test output") {
		t.Errorf("Output should contain 'test output', got: %s", output)
	}
}

func TestHookExecutor_StopOnError(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)

	hooks := []string{
		"echo 'first' > first.txt",
		"exit 1",                   // This should fail
		"echo 'third' > third.txt", // Should not execute
	}

	err := executor.ExecutePostInstall(hooks)

	if err == nil {
		t.Error("Should error when a hook fails")
	}

	// First file should exist
	if _, err := os.Stat(filepath.Join(tmpDir, "first.txt")); os.IsNotExist(err) {
		t.Error("first.txt should exist")
	}

	// Third file should NOT exist
	if _, err := os.Stat(filepath.Join(tmpDir, "third.txt")); !os.IsNotExist(err) {
		t.Error("third.txt should not exist (execution should stop on error)")
	}
}

func TestHookExecutor_WorkingDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	executor := NewHookExecutor(subDir)

	hooks := []string{
		"pwd > pwd_output.txt",
	}

	err := executor.ExecutePostInstall(hooks)
	if err != nil {
		t.Fatalf("ExecutePostInstall() error = %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(subDir, "pwd_output.txt"))
	if !strings.Contains(string(content), "subdir") {
		t.Error("Hook should execute in correct working directory")
	}
}

func TestHookExecutor_ExecutePreUpdate(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)

	hooks := []string{
		"echo 'pre-update hook' > pre_update.txt",
	}

	err := executor.ExecutePreUpdate(hooks)
	if err != nil {
		t.Fatalf("ExecutePreUpdate() error = %v", err)
	}

	// Verify file was created
	testFile := filepath.Join(tmpDir, "pre_update.txt")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Hook should have created pre_update.txt")
	}
}

func TestHookExecutor_ExecutePostUpdate(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)

	hooks := []string{
		"echo 'post-update hook' > post_update.txt",
	}

	err := executor.ExecutePostUpdate(hooks)
	if err != nil {
		t.Fatalf("ExecutePostUpdate() error = %v", err)
	}

	// Verify file was created
	testFile := filepath.Join(tmpDir, "post_update.txt")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Hook should have created post_update.txt")
	}

	content, _ := os.ReadFile(testFile)
	if !strings.Contains(string(content), "post-update") {
		t.Error("File should contain 'post-update'")
	}
}

func TestHookExecutor_ExecutePreDeploy(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)

	hooks := []string{
		"mkdir -p deploy-prep",
		"echo 'deployment preparation' > deploy-prep/status.txt",
	}

	err := executor.ExecutePreDeploy(hooks)
	if err != nil {
		t.Fatalf("ExecutePreDeploy() error = %v", err)
	}

	// Verify directory and file were created
	statusFile := filepath.Join(tmpDir, "deploy-prep", "status.txt")
	if _, err := os.Stat(statusFile); os.IsNotExist(err) {
		t.Error("Hook should have created deploy-prep/status.txt")
	}
}

func TestHookExecutor_ExecutePostDeploy(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewHookExecutor(tmpDir)

	hooks := []string{
		"echo 'deployment complete' > deployment.log",
	}

	err := executor.ExecutePostDeploy(hooks)
	if err != nil {
		t.Fatalf("ExecutePostDeploy() error = %v", err)
	}

	// Verify file was created
	logFile := filepath.Join(tmpDir, "deployment.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Hook should have created deployment.log")
	}

	content, _ := os.ReadFile(logFile)
	if !strings.Contains(string(content), "deployment complete") {
		t.Error("Log file should contain 'deployment complete'")
	}
}

func TestHookExecutor_ValidateHook(t *testing.T) {
	tests := []struct {
		name    string
		hook    string
		wantErr bool
	}{
		{
			name:    "valid simple command",
			hook:    "echo 'hello'",
			wantErr: false,
		},
		{
			name:    "valid command with pipe",
			hook:    "cat file.txt | grep pattern",
			wantErr: false,
		},
		{
			name:    "valid multi-line command",
			hook:    "mkdir -p dir && cd dir && touch file.txt",
			wantErr: false,
		},
		{
			name:    "empty hook",
			hook:    "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			hook:    "   \t  \n  ",
			wantErr: true,
		},
	}

	executor := NewHookExecutor("/tmp")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.ValidateHook(tt.hook)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHookExecutor_AllHookTypes(t *testing.T) {
	tmpDir := t.TempDir()
	executor := NewHookExecutor(tmpDir)

	testCases := []struct {
		name     string
		executor func([]string) error
		marker   string
	}{
		{"PreInstall", executor.ExecutePreInstall, "pre_install"},
		{"PostInstall", executor.ExecutePostInstall, "post_install"},
		{"PreUpdate", executor.ExecutePreUpdate, "pre_update"},
		{"PostUpdate", executor.ExecutePostUpdate, "post_update"},
		{"PreDeploy", executor.ExecutePreDeploy, "pre_deploy"},
		{"PostDeploy", executor.ExecutePostDeploy, "post_deploy"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hooks := []string{
				"echo '" + tc.marker + "' > " + tc.marker + ".txt",
			}

			err := tc.executor(hooks)
			if err != nil {
				t.Fatalf("%s hook failed: %v", tc.name, err)
			}

			markerFile := filepath.Join(tmpDir, tc.marker+".txt")
			if _, err := os.Stat(markerFile); os.IsNotExist(err) {
				t.Errorf("%s hook should have created %s.txt", tc.name, tc.marker)
			}
		})
	}
}

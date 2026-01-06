package validationops

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

func TestValidateGoVersionTask(t *testing.T) {
	tests := []struct {
		name       string
		minVersion string
		wantErr    bool
	}{
		{
			name:       "valid minimum version",
			minVersion: "1.18",
			wantErr:    false,
		},
		{
			name:       "missing version",
			minVersion: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &ValidateGoVersionTask{MinVersion: tt.minVersion}
			
			err := task.Validate()
			if tt.name == "missing version" {
				if err == nil {
					t.Error("expected validation error")
				}
				return
			}

			if err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			ctx := context.Background()
			taskCtx := &tasks.TaskContext{}
			err = task.Execute(ctx, taskCtx)

			// We can't predict if execution will fail since it depends on actual Go version.
			if err != nil && !strings.Contains(err.Error(), "version") {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateDependenciesTask(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []string
		wantErr      bool
	}{
		{
			name:         "go command exists",
			dependencies: []string{"go"},
			wantErr:      false,
		},
		{
			name:         "nonexistent command",
			dependencies: []string{"this-command-should-not-exist-12345"},
			wantErr:      true,
		},
		{
			name:         "empty dependencies",
			dependencies: []string{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &ValidateDependenciesTask{Dependencies: tt.dependencies}

			err := task.Validate()
			if tt.name == "empty dependencies" {
				if err == nil {
					t.Error("expected validation error")
				}
				return
			}

			if err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			ctx := context.Background()
			taskCtx := &tasks.TaskContext{}
			err = task.Execute(ctx, taskCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfigTask(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a valid config file.
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte("key: value\n"), 0600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{
			name:    "valid config file",
			file:    configPath,
			wantErr: false,
		},
		{
			name:    "missing config file",
			file:    filepath.Join(tmpDir, "nonexistent.yaml"),
			wantErr: true,
		},
		{
			name:    "missing file parameter",
			file:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &ValidateConfigTask{File: tt.file}

			err := task.Validate()
			if tt.name == "missing file parameter" {
				if err == nil {
					t.Error("expected validation error")
				}
				return
			}

			if err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			ctx := context.Background()
			taskCtx := &tasks.TaskContext{}
			err = task.Execute(ctx, taskCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvCheckTask(t *testing.T) {
	// Set a test environment variable.
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	tests := []struct {
		name     string
		required []string
		wantErr  bool
	}{
		{
			name:     "existing environment variable",
			required: []string{"TEST_ENV_VAR"},
			wantErr:  false,
		},
		{
			name:     "missing environment variable",
			required: []string{"NONEXISTENT_ENV_VAR_12345"},
			wantErr:  true,
		},
		{
			name:     "empty required list",
			required: []string{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &EnvCheckTask{Required: tt.required}

			err := task.Validate()
			if tt.name == "empty required list" {
				if err == nil {
					t.Error("expected validation error")
				}
				return
			}

			if err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			ctx := context.Background()
			taskCtx := &tasks.TaskContext{}
			err = task.Execute(ctx, taskCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPortCheckTask(t *testing.T) {
	// Skip on Windows as network tests can be flaky.
	if runtime.GOOS == "windows" {
		t.Skip("Skipping port test on Windows")
	}

	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{
			name:    "high unused port",
			port:    45678,
			wantErr: false,
		},
		{
			name:    "invalid port",
			port:    0,
			wantErr: true,
		},
		{
			name:    "negative port",
			port:    -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &PortCheckTask{Port: tt.port}

			err := task.Validate()
			if tt.port <= 0 || tt.port > 65535 {
				if err == nil {
					t.Error("expected validation error for invalid port")
				}
				return
			}

			if err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			ctx := context.Background()
			taskCtx := &tasks.TaskContext{}
			err = task.Execute(ctx, taskCtx)

			// We expect success for unused ports.
			if err != nil && !tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationTasksValidation(t *testing.T) {
	tests := []struct {
		name    string
		task    tasks.Task
		wantErr bool
	}{
		{
			name:    "ValidateGoVersionTask missing version",
			task:    &ValidateGoVersionTask{},
			wantErr: true,
		},
		{
			name:    "ValidateDependenciesTask empty list",
			task:    &ValidateDependenciesTask{},
			wantErr: true,
		},
		{
			name:    "ValidateConfigTask missing file",
			task:    &ValidateConfigTask{},
			wantErr: true,
		},
		{
			name:    "EnvCheckTask empty list",
			task:    &EnvCheckTask{},
			wantErr: true,
		},
		{
			name:    "PortCheckTask invalid port",
			task:    &PortCheckTask{Port: -1},
			wantErr: true,
		},
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

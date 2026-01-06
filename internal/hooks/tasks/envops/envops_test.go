package envops_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks/envops"
)

func TestEnvSetTask(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	tests := []struct {
		name        string
		task        *envops.EnvSetTask
		existingEnv string
		wantEnv     string
		wantErr     bool
	}{
		{
			name: "create new env file",
			task: &envops.EnvSetTask{
				File:  envFile,
				Key:   "DATABASE_URL",
				Value: "postgres://localhost/mydb",
			},
			existingEnv: "",
			wantEnv:     "DATABASE_URL=postgres://localhost/mydb\n",
			wantErr:     false,
		},
		{
			name: "add to existing env file",
			task: &envops.EnvSetTask{
				File:  envFile,
				Key:   "API_KEY",
				Value: "secret123",
			},
			existingEnv: "DATABASE_URL=postgres://localhost/mydb\n",
			wantEnv:     "DATABASE_URL=postgres://localhost/mydb\nAPI_KEY=secret123\n",
			wantErr:     false,
		},
		{
			name: "update existing key",
			task: &envops.EnvSetTask{
				File:  envFile,
				Key:   "DATABASE_URL",
				Value: "postgres://localhost/newdb",
			},
			existingEnv: "DATABASE_URL=postgres://localhost/mydb\nAPI_KEY=secret123\n",
			wantEnv:     "DATABASE_URL=postgres://localhost/newdb\nAPI_KEY=secret123\n",
			wantErr:     false,
		},
		{
			name: "preserve comments",
			task: &envops.EnvSetTask{
				File:  envFile,
				Key:   "NEW_VAR",
				Value: "value",
			},
			existingEnv: "# Database config\nDATABASE_URL=postgres://localhost/mydb\n",
			wantEnv:     "# Database config\nDATABASE_URL=postgres://localhost/mydb\nNEW_VAR=value\n",
			wantErr:     false,
		},
		{
			name: "handle empty value",
			task: &envops.EnvSetTask{
				File:  envFile,
				Key:   "EMPTY_VAR",
				Value: "",
			},
			existingEnv: "",
			wantEnv:     "EMPTY_VAR=\n",
			wantErr:     false,
		},
		{
			name: "handle quoted values",
			task: &envops.EnvSetTask{
				File:  envFile,
				Key:   "QUOTED",
				Value: "value with spaces",
			},
			existingEnv: "",
			wantEnv:     "QUOTED=\"value with spaces\"\n",
			wantErr:     false,
		},
		{
			name: "validation error - no key",
			task: &envops.EnvSetTask{
				File:  envFile,
				Key:   "",
				Value: "value",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup existing env file if specified.
			if tt.existingEnv != "" {
				if err := os.WriteFile(envFile, []byte(tt.existingEnv), 0600); err != nil {
					t.Fatalf("failed to create test env file: %v", err)
				}
			} else {
				os.Remove(envFile)
			}

			// Validate task.
			err := tt.task.Validate()
			if tt.wantErr && err != nil {
				return // Expected validation error.
			}
			if err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			// Execute task.
			taskCtx := tasks.NewTaskContext()
			taskCtx.SetWorkingDir(tmpDir)
			err = tt.task.Execute(context.Background(), taskCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check result.
			got, err := os.ReadFile(envFile)
			if err != nil {
				t.Fatalf("failed to read result: %v", err)
			}

			if string(got) != tt.wantEnv {
				t.Errorf("env file content:\ngot:  %q\nwant: %q", string(got), tt.wantEnv)
			}
		})
	}
}

func TestEnvSetTask_Name(t *testing.T) {
	task := &envops.EnvSetTask{}
	if got := task.Name(); got != "env-set" {
		t.Errorf("Name() = %v, want %v", got, "env-set")
	}
}

func TestEnvSetTask_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := ".env"

	task := &envops.EnvSetTask{
		File:  envFile,
		Key:   "TEST_VAR",
		Value: "test_value",
	}

	if err := task.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	taskCtx := tasks.NewTaskContext()
	taskCtx.SetWorkingDir(tmpDir)
	err := task.Execute(context.Background(), taskCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Check file was created in tmpDir.
	fullPath := filepath.Join(tmpDir, envFile)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("env file not created at %s", fullPath)
	}
}

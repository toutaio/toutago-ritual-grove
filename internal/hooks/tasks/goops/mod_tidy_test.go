package goops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

func TestGoModTidy(t *testing.T) {
	tests := []struct {
		name    string
		task    *GoModTidyTask
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "success - tidy go.mod",
			task: &GoModTidyTask{},
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create a simple go.mod.
				modContent := `module test

go 1.21
`
				if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(modContent), 0600); err != nil {
					t.Fatal(err)
				}
				// Create a simple main.go.
				mainContent := `package main

func main() {}
`
				if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainContent), 0600); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantErr: false,
		},
		{
			name: "success - with custom directory",
			task: &GoModTidyTask{Dir: "subdir"},
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				subdir := filepath.Join(tmpDir, "subdir")
				if err := os.MkdirAll(subdir, 0755); err != nil {
					t.Fatal(err)
				}
				// Create go.mod in subdir.
				modContent := `module test

go 1.21
`
				if err := os.WriteFile(filepath.Join(subdir, "go.mod"), []byte(modContent), 0600); err != nil {
					t.Fatal(err)
				}
				// Create main.go in subdir.
				mainContent := `package main

func main() {}
`
				if err := os.WriteFile(filepath.Join(subdir, "main.go"), []byte(mainContent), 0600); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantErr: false,
		},
		{
			name: "error - no go.mod",
			task: &GoModTidyTask{},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := tt.setup(t)

			taskCtx := tasks.NewTaskContext()
			taskCtx.SetWorkingDir(baseDir)

			err := tt.task.Execute(context.Background(), taskCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

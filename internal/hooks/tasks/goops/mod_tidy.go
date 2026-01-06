package goops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

// GoModTidyTask runs go mod tidy.
type GoModTidyTask struct {
	// Dir is the directory containing go.mod (optional, defaults to working dir).
	Dir string
}

// Name returns the task name.
func (t *GoModTidyTask) Name() string {
	return "go-mod-tidy"
}

// Validate validates the task parameters.
func (t *GoModTidyTask) Validate() error {
	return nil
}

// Execute runs go mod tidy.
func (t *GoModTidyTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	dir := taskCtx.WorkingDir()
	if t.Dir != "" {
		if filepath.IsAbs(t.Dir) {
			dir = t.Dir
		} else {
			dir = filepath.Join(taskCtx.WorkingDir(), t.Dir)
		}
	}

	// Check if go.mod exists.
	goModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		return fmt.Errorf("go.mod not found in %s", dir)
	}

	// Run go mod tidy.
	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	return nil
}

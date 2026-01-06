// Package fileops provides file operation tasks for hooks.
package fileops

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

// MkdirTask creates a directory.
type MkdirTask struct {
	Path string
	Perm os.FileMode
}

func (t *MkdirTask) Name() string {
	return "mkdir"
}

func (t *MkdirTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	path := t.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(taskCtx.WorkingDir(), path)
	}

	return os.MkdirAll(path, t.Perm)
}

func (t *MkdirTask) Validate() error {
	if t.Path == "" {
		return errors.New("path is required")
	}
	if t.Perm == 0 {
		return errors.New("perm must be specified")
	}
	return nil
}

// CopyTask copies a file or directory.
type CopyTask struct {
	Src  string
	Dest string
}

func (t *CopyTask) Name() string {
	return "copy"
}

func (t *CopyTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	src, dest := t.Src, t.Dest

	if !filepath.IsAbs(src) {
		src = filepath.Join(taskCtx.WorkingDir(), src)
	}
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(taskCtx.WorkingDir(), dest)
	}

	// Check if source exists.
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source error: %w", err)
	}

	if srcInfo.IsDir() {
		return copyDir(src, dest)
	}
	return copyFile(src, dest)
}

func (t *CopyTask) Validate() error {
	if t.Src == "" {
		return errors.New("src is required")
	}
	if t.Dest == "" {
		return errors.New("dest is required")
	}
	return nil
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination directory if needed.
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	// Copy permissions.
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dest, srcInfo.Mode())
}

func copyDir(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory.
	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// MoveTask moves a file or directory.
type MoveTask struct {
	Src  string
	Dest string
}

func (t *MoveTask) Name() string {
	return "move"
}

func (t *MoveTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	src, dest := t.Src, t.Dest

	if !filepath.IsAbs(src) {
		src = filepath.Join(taskCtx.WorkingDir(), src)
	}
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(taskCtx.WorkingDir(), dest)
	}

	return os.Rename(src, dest)
}

func (t *MoveTask) Validate() error {
	if t.Src == "" {
		return errors.New("src is required")
	}
	if t.Dest == "" {
		return errors.New("dest is required")
	}
	return nil
}

// RemoveTask removes a file or directory.
type RemoveTask struct {
	Path string
}

func (t *RemoveTask) Name() string {
	return "remove"
}

func (t *RemoveTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	path := t.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(taskCtx.WorkingDir(), path)
	}

	return os.RemoveAll(path)
}

func (t *RemoveTask) Validate() error {
	if t.Path == "" {
		return errors.New("path is required")
	}
	return nil
}

// ChmodTask changes file permissions.
type ChmodTask struct {
	Path string
	Perm os.FileMode
}

func (t *ChmodTask) Name() string {
	return "chmod"
}

func (t *ChmodTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	path := t.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(taskCtx.WorkingDir(), path)
	}

	return os.Chmod(path, t.Perm)
}

func (t *ChmodTask) Validate() error {
	if t.Path == "" {
		return errors.New("path is required")
	}
	if t.Perm == 0 {
		return errors.New("perm must be specified")
	}
	return nil
}

// Register all file operation tasks.
func init() {
	tasks.Register("mkdir", func(config map[string]interface{}) (tasks.Task, error) {
		path, _ := config["path"].(string)
		permFloat, _ := config["perm"].(float64)
		perm := os.FileMode(permFloat)
		if perm == 0 {
			perm = 0755 // Default.
		}
		return &MkdirTask{Path: path, Perm: perm}, nil
	})

	tasks.Register("copy", func(config map[string]interface{}) (tasks.Task, error) {
		src, _ := config["src"].(string)
		dest, _ := config["dest"].(string)
		return &CopyTask{Src: src, Dest: dest}, nil
	})

	tasks.Register("move", func(config map[string]interface{}) (tasks.Task, error) {
		src, _ := config["src"].(string)
		dest, _ := config["dest"].(string)
		return &MoveTask{Src: src, Dest: dest}, nil
	})

	tasks.Register("remove", func(config map[string]interface{}) (tasks.Task, error) {
		path, _ := config["path"].(string)
		return &RemoveTask{Path: path}, nil
	})

	tasks.Register("chmod", func(config map[string]interface{}) (tasks.Task, error) {
		path, _ := config["path"].(string)
		permFloat, _ := config["perm"].(float64)
		perm := os.FileMode(permFloat)
		return &ChmodTask{Path: path, Perm: perm}, nil
	})
}

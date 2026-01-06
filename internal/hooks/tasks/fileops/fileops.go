// Package fileops provides file operation tasks for hooks.
package fileops

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

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

// TemplateRenderTask renders a template to a file.
type TemplateRenderTask struct {
	Template string
	Dest     string
	Data     map[string]interface{}
}

func (t *TemplateRenderTask) Name() string {
	return "template-render"
}

func (t *TemplateRenderTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	templatePath := t.Template
	destPath := t.Dest

	if !filepath.IsAbs(templatePath) {
		templatePath = filepath.Join(taskCtx.WorkingDir(), templatePath)
	}
	if !filepath.IsAbs(destPath) {
		destPath = filepath.Join(taskCtx.WorkingDir(), destPath)
	}

	// Read template file.
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	// Parse template.
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	// Create destination directory if needed.
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	// Render to destination file.
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create dest file: %w", err)
	}
	defer destFile.Close()

	data := t.Data
	if data == nil {
		data = make(map[string]interface{})
	}

	// Merge with context data.
	for k, v := range taskCtx.Data() {
		if _, exists := data[k]; !exists {
			data[k] = v
		}
	}

	if err := tmpl.Execute(destFile, data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	return nil
}

func (t *TemplateRenderTask) Validate() error {
	if t.Template == "" {
		return errors.New("template is required")
	}
	if t.Dest == "" {
		return errors.New("dest is required")
	}
	return nil
}

// ValidateFilesTask validates that files exist.
type ValidateFilesTask struct {
	Files []string
}

func (t *ValidateFilesTask) Name() string {
	return "validate-files"
}

func (t *ValidateFilesTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	for _, file := range t.Files {
		path := file
		if !filepath.IsAbs(path) {
			path = filepath.Join(taskCtx.WorkingDir(), path)
		}

		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", file)
			}
			return fmt.Errorf("check file %s: %w", file, err)
		}
	}
	return nil
}

func (t *ValidateFilesTask) Validate() error {
	if len(t.Files) == 0 {
		return errors.New("files list is required")
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

	tasks.Register("template-render", func(config map[string]interface{}) (tasks.Task, error) {
		template, _ := config["template"].(string)
		dest, _ := config["dest"].(string)
		data, _ := config["data"].(map[string]interface{})
		return &TemplateRenderTask{Template: template, Dest: dest, Data: data}, nil
	})

	tasks.Register("validate-files", func(config map[string]interface{}) (tasks.Task, error) {
		filesRaw, _ := config["files"].([]interface{})
		files := make([]string, 0, len(filesRaw))
		for _, f := range filesRaw {
			if str, ok := f.(string); ok {
				files = append(files, str)
			}
		}
		return &ValidateFilesTask{Files: files}, nil
	})
}

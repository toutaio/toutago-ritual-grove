package goops

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

// GoModDownloadTask runs go mod download.
type GoModDownloadTask struct {
	Dir string
}

func (t *GoModDownloadTask) Name() string {
	return "go-mod-download"
}

func (t *GoModDownloadTask) Validate() error {
	return nil
}

func (t *GoModDownloadTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
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

	cmd := exec.CommandContext(ctx, "go", "mod", "download")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod download failed: %w", err)
	}

	return nil
}

// GoBuildTask runs go build.
type GoBuildTask struct {
	Dir    string
	Output string
	Args   []string
}

func (t *GoBuildTask) Name() string {
	return "go-build"
}

func (t *GoBuildTask) Validate() error {
	return nil
}

func (t *GoBuildTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	dir := taskCtx.WorkingDir()
	if t.Dir != "" {
		if filepath.IsAbs(t.Dir) {
			dir = t.Dir
		} else {
			dir = filepath.Join(taskCtx.WorkingDir(), t.Dir)
		}
	}

	args := []string{"build"}
	if t.Output != "" {
		outputPath := t.Output
		if !filepath.IsAbs(outputPath) {
			outputPath = filepath.Join(dir, outputPath)
		}
		args = append(args, "-o", outputPath)
	}
	args = append(args, t.Args...)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	return nil
}

// GoTestTask runs go test.
type GoTestTask struct {
	Dir  string
	Args []string
}

func (t *GoTestTask) Name() string {
	return "go-test"
}

func (t *GoTestTask) Validate() error {
	return nil
}

func (t *GoTestTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	dir := taskCtx.WorkingDir()
	if t.Dir != "" {
		if filepath.IsAbs(t.Dir) {
			dir = t.Dir
		} else {
			dir = filepath.Join(taskCtx.WorkingDir(), t.Dir)
		}
	}

	args := []string{"test"}
	args = append(args, t.Args...)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go test failed: %w", err)
	}

	return nil
}

// GoFmtTask runs go fmt.
type GoFmtTask struct {
	Dir string
}

func (t *GoFmtTask) Name() string {
	return "go-fmt"
}

func (t *GoFmtTask) Validate() error {
	return nil
}

func (t *GoFmtTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	dir := taskCtx.WorkingDir()
	if t.Dir != "" {
		if filepath.IsAbs(t.Dir) {
			dir = t.Dir
		} else {
			dir = filepath.Join(taskCtx.WorkingDir(), t.Dir)
		}
	}

	cmd := exec.CommandContext(ctx, "go", "fmt", "./...")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go fmt failed: %w", err)
	}

	return nil
}

// GoRunTask runs go run.
type GoRunTask struct {
	File string
	Args []string
	Dir  string
	Env  map[string]string
}

func (t *GoRunTask) Name() string {
	return "go-run"
}

func (t *GoRunTask) Validate() error {
	if t.File == "" {
		return errors.New("file is required")
	}
	return nil
}

func (t *GoRunTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	dir := taskCtx.WorkingDir()
	if t.Dir != "" {
		if filepath.IsAbs(t.Dir) {
			dir = t.Dir
		} else {
			dir = filepath.Join(taskCtx.WorkingDir(), t.Dir)
		}
	}

	args := []string{"run", t.File}
	args = append(args, t.Args...)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment variables.
	if len(t.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range t.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go run failed: %w", err)
	}

	return nil
}

// ExecGoTask runs an arbitrary go command.
type ExecGoTask struct {
	Command []string
	Dir     string
}

func (t *ExecGoTask) Name() string {
	return "exec-go"
}

func (t *ExecGoTask) Validate() error {
	if len(t.Command) == 0 {
		return errors.New("command is required")
	}
	return nil
}

func (t *ExecGoTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	dir := taskCtx.WorkingDir()
	if t.Dir != "" {
		if filepath.IsAbs(t.Dir) {
			dir = t.Dir
		} else {
			dir = filepath.Join(taskCtx.WorkingDir(), t.Dir)
		}
	}

	cmd := exec.CommandContext(ctx, "go", t.Command...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go %v failed: %w", t.Command, err)
	}

	return nil
}

// Register all Go operation tasks.
func init() {
	tasks.Register("go-mod-tidy", func(config map[string]interface{}) (tasks.Task, error) {
		dir, _ := config["dir"].(string)
		return &GoModTidyTask{Dir: dir}, nil
	})

	tasks.Register("go-mod-download", func(config map[string]interface{}) (tasks.Task, error) {
		dir, _ := config["dir"].(string)
		return &GoModDownloadTask{Dir: dir}, nil
	})

	tasks.Register("go-build", func(config map[string]interface{}) (tasks.Task, error) {
		dir, _ := config["dir"].(string)
		output, _ := config["output"].(string)
		argsRaw, _ := config["args"].([]interface{})
		args := make([]string, 0, len(argsRaw))
		for _, a := range argsRaw {
			if str, ok := a.(string); ok {
				args = append(args, str)
			}
		}
		return &GoBuildTask{Dir: dir, Output: output, Args: args}, nil
	})

	tasks.Register("go-test", func(config map[string]interface{}) (tasks.Task, error) {
		dir, _ := config["dir"].(string)
		argsRaw, _ := config["args"].([]interface{})
		args := make([]string, 0, len(argsRaw))
		for _, a := range argsRaw {
			if str, ok := a.(string); ok {
				args = append(args, str)
			}
		}
		return &GoTestTask{Dir: dir, Args: args}, nil
	})

	tasks.Register("go-fmt", func(config map[string]interface{}) (tasks.Task, error) {
		dir, _ := config["dir"].(string)
		return &GoFmtTask{Dir: dir}, nil
	})

	tasks.Register("go-run", func(config map[string]interface{}) (tasks.Task, error) {
		file, _ := config["file"].(string)
		dir, _ := config["dir"].(string)

		argsRaw, _ := config["args"].([]interface{})
		args := make([]string, 0, len(argsRaw))
		for _, a := range argsRaw {
			if str, ok := a.(string); ok {
				args = append(args, str)
			}
		}

		envMap := make(map[string]string)
		if envRaw, ok := config["env"].(map[string]interface{}); ok {
			for k, v := range envRaw {
				if str, ok := v.(string); ok {
					envMap[k] = str
				}
			}
		}

		return &GoRunTask{File: file, Args: args, Dir: dir, Env: envMap}, nil
	})

	tasks.Register("exec-go", func(config map[string]interface{}) (tasks.Task, error) {
		dir, _ := config["dir"].(string)
		cmdRaw, _ := config["command"].([]interface{})
		cmd := make([]string, 0, len(cmdRaw))
		for _, c := range cmdRaw {
			if str, ok := c.(string); ok {
				cmd = append(cmd, str)
			}
		}
		return &ExecGoTask{Command: cmd, Dir: dir}, nil
	})
}

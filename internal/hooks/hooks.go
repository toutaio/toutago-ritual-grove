package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

// HookExecutor executes lifecycle hooks
type HookExecutor struct {
	workDir string
	timeout time.Duration
	dryRun  bool
	env     map[string]string
	output  bytes.Buffer
}

// NewHookExecutor creates a new hook executor
func NewHookExecutor(workDir string) *HookExecutor {
	return &HookExecutor{
		workDir: workDir,
		timeout: 5 * time.Minute, // Default timeout
		env:     make(map[string]string),
	}
}

// SetTimeout sets the execution timeout for hooks
func (e *HookExecutor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// SetDryRun enables/disables dry run mode
func (e *HookExecutor) SetDryRun(dryRun bool) {
	e.dryRun = dryRun
}

// SetEnv sets an environment variable for hook execution
func (e *HookExecutor) SetEnv(key, value string) {
	e.env[key] = value
}

// GetOutput returns the captured output from hooks
func (e *HookExecutor) GetOutput() string {
	return e.output.String()
}

// ExecutePreInstall executes pre-install hooks
func (e *HookExecutor) ExecutePreInstall(hooks []string) error {
	return e.executeHooks(hooks, "pre-install")
}

// ExecutePostInstall executes post-install hooks
func (e *HookExecutor) ExecutePostInstall(hooks []string) error {
	return e.executeHooks(hooks, "post-install")
}

// ExecutePreUpdate executes pre-update hooks
func (e *HookExecutor) ExecutePreUpdate(hooks []string) error {
	return e.executeHooks(hooks, "pre-update")
}

// ExecutePostUpdate executes post-update hooks
func (e *HookExecutor) ExecutePostUpdate(hooks []string) error {
	return e.executeHooks(hooks, "post-update")
}

// ExecutePreDeploy executes pre-deploy hooks
func (e *HookExecutor) ExecutePreDeploy(hooks []string) error {
	return e.executeHooks(hooks, "pre-deploy")
}

// ExecutePostDeploy executes post-deploy hooks
func (e *HookExecutor) ExecutePostDeploy(hooks []string) error {
	return e.executeHooks(hooks, "post-deploy")
}

// executeHooks executes a list of hook commands
func (e *HookExecutor) executeHooks(hooks []string, phase string) error {
	if len(hooks) == 0 {
		return nil
	}

	for i, hook := range hooks {
		if err := e.executeHook(hook, phase, i+1, len(hooks)); err != nil {
			return fmt.Errorf("hook %d/%d failed in %s phase: %w", i+1, len(hooks), phase, err)
		}
	}

	return nil
}

// executeHook executes a single hook command or task object
func (e *HookExecutor) executeHook(command, phase string, index, total int) error {
	// Check if this is a task object (JSON) or shell command
	if isTaskObject(command) {
		return e.executeTask(command, phase, index, total)
	}
	return e.executeShellCommand(command, phase, index, total)
}

// executeTask executes a declarative task object
func (e *HookExecutor) executeTask(taskJSON, phase string, index, total int) error {
	// Parse task object
	taskData, err := parseTaskObject(taskJSON)
	if err != nil {
		return fmt.Errorf("failed to parse task object: %w", err)
	}

	taskType, ok := taskData["type"].(string)
	if !ok {
		return fmt.Errorf("task object missing 'type' field")
	}

	if e.dryRun {
		e.output.WriteString(fmt.Sprintf("[DRY RUN] %s task %d/%d: %s\n", phase, index, total, taskType))
		return nil
	}

	// Create task from registry
	task, err := createTaskFromData(taskData)
	if err != nil {
		return fmt.Errorf("failed to create task '%s': %w", taskType, err)
	}

	// Create task execution context
	ctx := tasks.NewTaskContext()
	ctx.SetWorkingDir(e.workDir)
	for k, v := range e.env {
		ctx.SetEnv(k, v)
	}

	// Execute task
	e.output.WriteString(fmt.Sprintf("[%s %d/%d] Task: %s\n", phase, index, total, taskType))
	execCtx := context.Background()
	if err := task.Execute(execCtx, ctx); err != nil {
		return fmt.Errorf("task '%s' failed: %w", taskType, err)
	}

	return nil
}

// executeShellCommand executes a shell command string
func (e *HookExecutor) executeShellCommand(command, phase string, index, total int) error {
	if e.dryRun {
		e.output.WriteString(fmt.Sprintf("[DRY RUN] %s hook %d/%d: %s\n", phase, index, total, command))
		return nil
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	// Use shell to execute command
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = e.workDir

	// Set environment variables
	cmd.Env = append(cmd.Environ(), e.envSlice()...)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()

	// Store output
	e.output.WriteString(fmt.Sprintf("[%s %d/%d] %s\n", phase, index, total, command))
	if stdout.Len() > 0 {
		e.output.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		e.output.WriteString(stderr.String())
	}

	// Check for errors
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command timeout after %v: %s", e.timeout, command)
		}
		return fmt.Errorf("command failed: %w\nStderr: %s", err, stderr.String())
	}

	return nil
}

// envSlice converts env map to slice of KEY=VALUE strings
func (e *HookExecutor) envSlice() []string {
	result := make([]string, 0, len(e.env))
	for key, value := range e.env {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}
	return result
}

// ValidateHook checks if a hook command is safe to execute
func (e *HookExecutor) ValidateHook(command string) error {
	// Basic validation - can be extended
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("empty hook command")
	}

	// If it's a task object, validate the JSON
	if isTaskObject(command) {
		_, err := parseTaskObject(command)
		return err
	}

	// Warn about potentially dangerous commands
	dangerous := []string{"rm -rf /", "dd if=", "mkfs", "> /dev/"}
	for _, pattern := range dangerous {
		if strings.Contains(command, pattern) {
			return fmt.Errorf("potentially dangerous command detected: %s", pattern)
		}
	}

	return nil
}

// isTaskObject checks if a hook string is a JSON task object
func isTaskObject(hook string) bool {
	trimmed := strings.TrimSpace(hook)
	if len(trimmed) == 0 {
		return false
	}

	// Must start with { and end with }
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		return false
	}

	// Try to parse as JSON and check for "type" field
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &data); err != nil {
		return false
	}

	// Must have a "type" field with string value
	_, ok := data["type"].(string)
	return ok
}

// parseTaskObject parses a JSON task object string
func parseTaskObject(taskJSON string) (map[string]interface{}, error) {
	var taskData map[string]interface{}
	if err := json.Unmarshal([]byte(taskJSON), &taskData); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Validate required "type" field
	if _, ok := taskData["type"].(string); !ok {
		return nil, fmt.Errorf("task object must have a 'type' field")
	}

	return taskData, nil
}

// createTaskFromData creates a task instance from parsed task data
func createTaskFromData(taskData map[string]interface{}) (tasks.Task, error) {
	taskType, ok := taskData["type"].(string)
	if !ok {
		return nil, fmt.Errorf("task must have a 'type' field")
	}

	// Use the global task registry to create task
	task, err := tasks.Get(taskType, taskData)
	if err != nil {
		return nil, fmt.Errorf("failed to create task '%s': %w", taskType, err)
	}

	return task, nil
}

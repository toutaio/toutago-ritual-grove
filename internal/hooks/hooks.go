package hooks

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
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

// executeHook executes a single hook command
func (e *HookExecutor) executeHook(command, phase string, index, total int) error {
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

	// Warn about potentially dangerous commands
	dangerous := []string{"rm -rf /", "dd if=", "mkfs", "> /dev/"}
	for _, pattern := range dangerous {
		if strings.Contains(command, pattern) {
			return fmt.Errorf("potentially dangerous command detected: %s", pattern)
		}
	}

	return nil
}

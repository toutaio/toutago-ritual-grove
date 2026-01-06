// Package tasks provides a declarative, cross-platform task system for ritual hooks.
package tasks

import (
	"context"
	"os"
	"sync"
)

// Task represents a single executable task in a hook.
type Task interface {
	// Name returns the task identifier.
	Name() string

	// Execute runs the task with the given context.
	Execute(ctx context.Context, taskCtx *TaskContext) error

	// Validate checks if the task configuration is valid.
	Validate() error
}

// TaskContext provides context and shared state for task execution.
type TaskContext struct {
	mu         sync.RWMutex
	data       map[string]interface{}
	workingDir string
	env        map[string]string
}

// NewTaskContext creates a new task context.
func NewTaskContext() *TaskContext {
	wd, _ := os.Getwd()
	return &TaskContext{
		data:       make(map[string]interface{}),
		workingDir: wd,
		env:        make(map[string]string),
	}
}

// Set stores a value in the context.
func (tc *TaskContext) Set(key string, value interface{}) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.data[key] = value
}

// Get retrieves a value from the context.
func (tc *TaskContext) Get(key string) (interface{}, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	val, ok := tc.data[key]
	return val, ok
}

// WorkingDir returns the current working directory for tasks.
func (tc *TaskContext) WorkingDir() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.workingDir
}

// SetWorkingDir sets the working directory for tasks.
func (tc *TaskContext) SetWorkingDir(dir string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.workingDir = dir
}

// Env returns an environment variable.
func (tc *TaskContext) Env(key string) string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	if val, ok := tc.env[key]; ok {
		return val
	}
	return os.Getenv(key)
}

// SetEnv sets an environment variable for task execution.
func (tc *TaskContext) SetEnv(key, value string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.env[key] = value
}

// AllEnv returns all environment variables including those set in context.
func (tc *TaskContext) AllEnv() map[string]string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	result := make(map[string]string)
	for _, e := range os.Environ() {
		// Parse KEY=VALUE.
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				result[e[:i]] = e[i+1:]
				break
			}
		}
	}

	// Override with context env.
	for k, v := range tc.env {
		result[k] = v
	}

	return result
}

// Data returns all data stored in the context.
func (tc *TaskContext) Data() map[string]interface{} {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	result := make(map[string]interface{}, len(tc.data))
	for k, v := range tc.data {
		result[k] = v
	}
	return result
}

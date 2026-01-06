package tasks

import (
	"context"
	"testing"
)

func TestTaskContext(t *testing.T) {
	ctx := NewTaskContext()
	
	// Test setting and getting values.
	ctx.Set("key1", "value1")
	ctx.Set("key2", 42)
	
	val1, ok := ctx.Get("key1")
	if !ok || val1 != "value1" {
		t.Errorf("Expected key1=value1, got %v", val1)
	}
	
	val2, ok := ctx.Get("key2")
	if !ok || val2 != 42 {
		t.Errorf("Expected key2=42, got %v", val2)
	}
	
	// Test non-existent key.
	_, ok = ctx.Get("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}
}

func TestTaskContextWorkingDir(t *testing.T) {
	ctx := NewTaskContext()
	
	// Test default working dir.
	wd := ctx.WorkingDir()
	if wd == "" {
		t.Error("Expected non-empty working dir")
	}
	
	// Test setting working dir.
	ctx.SetWorkingDir("/tmp")
	if ctx.WorkingDir() != "/tmp" {
		t.Errorf("Expected /tmp, got %s", ctx.WorkingDir())
	}
}

func TestTaskContextEnv(t *testing.T) {
	ctx := NewTaskContext()
	
	// Test setting env vars.
	ctx.SetEnv("VAR1", "value1")
	ctx.SetEnv("VAR2", "value2")
	
	val1 := ctx.Env("VAR1")
	if val1 != "value1" {
		t.Errorf("Expected value1, got %s", val1)
	}
	
	// Test non-existent var.
	val3 := ctx.Env("NONEXISTENT")
	if val3 != "" {
		t.Errorf("Expected empty string, got %s", val3)
	}
}

func TestTaskInterface(t *testing.T) {
	// Create a simple test task.
	testTask := &testTaskImpl{
		name: "test-task",
	}
	
	if testTask.Name() != "test-task" {
		t.Errorf("Expected test-task, got %s", testTask.Name())
	}
	
	ctx := NewTaskContext()
	ctx.Set("test", "value")
	
	err := testTask.Execute(context.Background(), ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Verify task executed.
	executed, _ := ctx.Get("executed")
	if executed != true {
		t.Error("Task was not executed")
	}
}

// Test task implementation.
type testTaskImpl struct {
	name string
}

func (t *testTaskImpl) Name() string {
	return t.name
}

func (t *testTaskImpl) Execute(ctx context.Context, taskCtx *TaskContext) error {
	taskCtx.Set("executed", true)
	return nil
}

func (t *testTaskImpl) Validate() error {
	return nil
}

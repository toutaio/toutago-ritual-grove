package tasks

import (
	"errors"
	"testing"
)

func TestRegistryRegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	// Register a task factory.
	registry.Register("test-task", func(config map[string]interface{}) (Task, error) {
		return &testTaskImpl{name: "test-task"}, nil
	})

	// Get the task.
	task, err := registry.Get("test-task", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if task.Name() != "test-task" {
		t.Errorf("Expected test-task, got %s", task.Name())
	}
}

func TestRegistryGetNonExistent(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("nonexistent", nil)
	if err == nil {
		t.Error("Expected error for non-existent task")
	}
}

func TestRegistryList(t *testing.T) {
	registry := NewRegistry()

	registry.Register("task1", func(config map[string]interface{}) (Task, error) {
		return &testTaskImpl{name: "task1"}, nil
	})
	registry.Register("task2", func(config map[string]interface{}) (Task, error) {
		return &testTaskImpl{name: "task2"}, nil
	})

	names := registry.List()
	if len(names) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(names))
	}

	found1, found2 := false, false
	for _, name := range names {
		if name == "task1" {
			found1 = true
		}
		if name == "task2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Expected to find both task1 and task2")
	}
}

func TestRegistryFactoryError(t *testing.T) {
	registry := NewRegistry()

	registry.Register("error-task", func(config map[string]interface{}) (Task, error) {
		return nil, errors.New("factory error")
	})

	_, err := registry.Get("error-task", nil)
	if err == nil {
		t.Error("Expected error from factory")
	}

	if err.Error() != "factory error" {
		t.Errorf("Expected 'factory error', got %s", err.Error())
	}
}

func TestRegistryWithConfig(t *testing.T) {
	registry := NewRegistry()

	registry.Register("config-task", func(config map[string]interface{}) (Task, error) {
		name, ok := config["name"].(string)
		if !ok {
			return nil, errors.New("missing name")
		}
		return &testTaskImpl{name: name}, nil
	})

	config := map[string]interface{}{
		"name": "custom-name",
	}

	task, err := registry.Get("config-task", config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if task.Name() != "custom-name" {
		t.Errorf("Expected custom-name, got %s", task.Name())
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Test that global registry exists and works.
	Register("global-task", func(config map[string]interface{}) (Task, error) {
		return &testTaskImpl{name: "global-task"}, nil
	})

	task, err := Get("global-task", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if task.Name() != "global-task" {
		t.Errorf("Expected global-task, got %s", task.Name())
	}

	names := List()
	found := false
	for _, name := range names {
		if name == "global-task" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find global-task in global registry")
	}
}

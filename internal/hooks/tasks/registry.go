package tasks

import (
	"fmt"
	"sync"
)

// TaskFactory creates a task from configuration.
type TaskFactory func(config map[string]interface{}) (Task, error)

// Registry manages task factories.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]TaskFactory
}

// NewRegistry creates a new task registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]TaskFactory),
	}
}

// Register adds a task factory to the registry.
func (r *Registry) Register(name string, factory TaskFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// Get creates a task instance from the registry.
func (r *Registry) Get(name string, config map[string]interface{}) (Task, error) {
	r.mu.RLock()
	factory, ok := r.factories[name]
	r.mu.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("task not found: %s", name)
	}
	
	return factory(config)
}

// List returns all registered task names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// Global registry for tasks.
var globalRegistry = NewRegistry()

// Register adds a task factory to the global registry.
func Register(name string, factory TaskFactory) {
	globalRegistry.Register(name, factory)
}

// Get creates a task instance from the global registry.
func Get(name string, config map[string]interface{}) (Task, error) {
	return globalRegistry.Get(name, config)
}

// List returns all registered task names from the global registry.
func List() []string {
	return globalRegistry.List()
}

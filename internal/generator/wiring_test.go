package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestNewComponentWiring(t *testing.T) {
	wiring := NewComponentWiring()
	if wiring == nil {
		t.Fatal("Expected wiring to be created")
	}
	if wiring.generator == nil {
		t.Fatal("Expected generator to be initialized")
	}
}

func TestWireComponents(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/myapp")
	vars.Set("app_name", "MyApp")
	vars.Set("port", "8080")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:        "test-ritual",
			Version:     "1.0.0",
			Description: "Test ritual",
		},
	}

	wiring := NewComponentWiring()
	err := wiring.WireComponents(tempDir, manifest, vars)
	if err != nil {
		t.Fatalf("WireComponents failed: %v", err)
	}

	// Check that DI container was created
	containerPath := filepath.Join(tempDir, "internal", "container", "container.go")
	if _, err := os.Stat(containerPath); os.IsNotExist(err) {
		t.Errorf("Container file was not created")
	}

	// Check that router was created
	routerPath := filepath.Join(tempDir, "internal", "router", "router.go")
	if _, err := os.Stat(routerPath); os.IsNotExist(err) {
		t.Errorf("Router file was not created")
	}

	// Check that middleware was created
	middlewarePath := filepath.Join(tempDir, "internal", "middleware", "middleware.go")
	if _, err := os.Stat(middlewarePath); os.IsNotExist(err) {
		t.Errorf("Middleware file was not created")
	}
}

func TestGenerateDIContainer(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/myapp")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
	}

	wiring := NewComponentWiring()
	err := wiring.generateDIContainer(tempDir, manifest, vars)
	if err != nil {
		t.Fatalf("generateDIContainer failed: %v", err)
	}

	containerPath := filepath.Join(tempDir, "internal", "container", "container.go")
	content, err := os.ReadFile(containerPath)
	if err != nil {
		t.Fatalf("Failed to read container file: %v", err)
	}

	// Check that the file contains expected content
	contentStr := string(content)
	if !strings.Contains(contentStr, "package container") {
		t.Error("Container file should have package container")
	}
	if !strings.Contains(contentStr, "type Container struct") {
		t.Error("Container file should define Container struct")
	}
	if !strings.Contains(contentStr, "func NewContainer") {
		t.Error("Container file should have NewContainer function")
	}
	if !strings.Contains(contentStr, "github.com/test/myapp") {
		t.Error("Container file should use correct module name")
	}
}

func TestGenerateRouterSetup(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/myapp")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
	}

	wiring := NewComponentWiring()
	err := wiring.generateRouterSetup(tempDir, manifest, vars)
	if err != nil {
		t.Fatalf("generateRouterSetup failed: %v", err)
	}

	routerPath := filepath.Join(tempDir, "internal", "router", "router.go")
	content, err := os.ReadFile(routerPath)
	if err != nil {
		t.Fatalf("Failed to read router file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package router") {
		t.Error("Router file should have package router")
	}
	if !strings.Contains(contentStr, "func Setup") {
		t.Error("Router file should have Setup function")
	}
	if !strings.Contains(contentStr, "/health") {
		t.Error("Router file should have health check route")
	}
}

func TestGenerateMiddlewareChain(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()

	wiring := NewComponentWiring()
	err := wiring.generateMiddlewareChain(tempDir, vars)
	if err != nil {
		t.Fatalf("generateMiddlewareChain failed: %v", err)
	}

	middlewarePath := filepath.Join(tempDir, "internal", "middleware", "middleware.go")
	content, err := os.ReadFile(middlewarePath)
	if err != nil {
		t.Fatalf("Failed to read middleware file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package middleware") {
		t.Error("Middleware file should have package middleware")
	}
	if !strings.Contains(contentStr, "func Chain") {
		t.Error("Middleware file should have Chain function")
	}
	if !strings.Contains(contentStr, "func Logger") {
		t.Error("Middleware file should have Logger middleware")
	}
	if !strings.Contains(contentStr, "func Recovery") {
		t.Error("Middleware file should have Recovery middleware")
	}
	if !strings.Contains(contentStr, "func CORS") {
		t.Error("Middleware file should have CORS middleware")
	}
}

func TestUpdateMainWithWiring(t *testing.T) {
	tempDir := t.TempDir()

	// Create cmd/server directory
	mainDir := filepath.Join(tempDir, "cmd", "server")
	if err := os.MkdirAll(mainDir, 0755); err != nil {
		t.Fatalf("Failed to create main directory: %v", err)
	}

	// Create initial main.go
	initialMain := `package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting server...")
	http.ListenAndServe(":8080", nil)
}
`
	mainPath := filepath.Join(mainDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(initialMain), 0644); err != nil {
		t.Fatalf("Failed to write initial main.go: %v", err)
	}

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/myapp")
	vars.Set("app_name", "MyApp")
	vars.Set("port", "8080")

	wiring := NewComponentWiring()
	err := wiring.updateMainWithWiring(tempDir, vars)
	if err != nil {
		t.Fatalf("updateMainWithWiring failed: %v", err)
	}

	// Read updated main.go
	content, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("Failed to read updated main.go: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "container.NewContainer") {
		t.Error("Updated main.go should use container.NewContainer")
	}
	if !strings.Contains(contentStr, "router.Setup") {
		t.Error("Updated main.go should use router.Setup")
	}
	if !strings.Contains(contentStr, "github.com/test/myapp/internal/container") {
		t.Error("Updated main.go should import container package")
	}
}

func TestUpdateMainWithWiring_NoMainFile(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/myapp")
	vars.Set("app_name", "MyApp")
	vars.Set("port", "8080")

	wiring := NewComponentWiring()
	err := wiring.updateMainWithWiring(tempDir, vars)
	// Should not error when main.go doesn't exist
	if err != nil {
		t.Fatalf("updateMainWithWiring should not fail when main.go doesn't exist: %v", err)
	}
}

func TestUpdateMainWithWiring_AlreadyWired(t *testing.T) {
	tempDir := t.TempDir()

	// Create cmd/server directory
	mainDir := filepath.Join(tempDir, "cmd", "server")
	if err := os.MkdirAll(mainDir, 0755); err != nil {
		t.Fatalf("Failed to create main directory: %v", err)
	}

	// Create main.go that's already wired
	wiredMain := `package main

import (
	"github.com/test/myapp/internal/container"
)

func main() {
	c := container.NewContainer(nil)
	defer c.Close()
}
`
	mainPath := filepath.Join(mainDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(wiredMain), 0644); err != nil {
		t.Fatalf("Failed to write wired main.go: %v", err)
	}

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/myapp")

	wiring := NewComponentWiring()
	err := wiring.updateMainWithWiring(tempDir, vars)
	if err != nil {
		t.Fatalf("updateMainWithWiring failed: %v", err)
	}

	// Read main.go - should be unchanged
	content, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	if string(content) != wiredMain {
		t.Error("Main.go should not be changed if already wired")
	}
}

func TestWireComponents_Integration(t *testing.T) {
	tempDir := t.TempDir()

	// Create a complete project setup
	vars := NewVariables()
	vars.Set("module_name", "github.com/test/integration")
	vars.Set("app_name", "IntegrationApp")
	vars.Set("port", "3000")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:        "integration-ritual",
			Version:     "1.0.0",
			Description: "Integration test ritual",
		},
	}

	wiring := NewComponentWiring()
	err := wiring.WireComponents(tempDir, manifest, vars)
	if err != nil {
		t.Fatalf("WireComponents failed: %v", err)
	}

	// Verify all components were created
	expectedFiles := []string{
		"internal/container/container.go",
		"internal/router/router.go",
		"internal/middleware/middleware.go",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tempDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}

	// Verify container file has proper structure
	containerPath := filepath.Join(tempDir, "internal", "container", "container.go")
	content, _ := os.ReadFile(containerPath)
	contentStr := string(content)

	requiredElements := []string{
		"package container",
		"type Container struct",
		"func NewContainer",
		"func (c *Container) Close()",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(contentStr, elem) {
			t.Errorf("Container file missing required element: %s", elem)
		}
	}
}

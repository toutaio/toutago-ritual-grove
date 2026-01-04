package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestNewMiddlewareGenerator(t *testing.T) {
	gen := NewMiddlewareGenerator()
	if gen == nil {
		t.Fatal("Expected middleware generator to be created")
	}
}

func TestMiddlewareGenerator_GenerateAuthMiddleware(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/myapp")

	gen := NewMiddlewareGenerator()
	err := gen.GenerateAuthMiddleware(tempDir, vars)
	if err != nil {
		t.Fatalf("GenerateAuthMiddleware failed: %v", err)
	}

	authPath := filepath.Join(tempDir, "internal", "middleware", "auth.go")
	content, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("Failed to read auth middleware: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package middleware") {
		t.Error("Auth middleware should have package middleware")
	}
	if !strings.Contains(contentStr, "func Auth") {
		t.Error("Auth middleware should have Auth function")
	}
	if !strings.Contains(contentStr, "Authorization") {
		t.Error("Auth middleware should check Authorization header")
	}
}

func TestMiddlewareGenerator_GenerateLoggingMiddleware(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()

	gen := NewMiddlewareGenerator()
	err := gen.GenerateLoggingMiddleware(tempDir, vars)
	if err != nil {
		t.Fatalf("GenerateLoggingMiddleware failed: %v", err)
	}

	loggingPath := filepath.Join(tempDir, "internal", "middleware", "logging.go")
	content, err := os.ReadFile(loggingPath)
	if err != nil {
		t.Fatalf("Failed to read logging middleware: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package middleware") {
		t.Error("Logging middleware should have package middleware")
	}
	if !strings.Contains(contentStr, "func RequestLogger") {
		t.Error("Logging middleware should have RequestLogger function")
	}
}

func TestMiddlewareGenerator_GenerateCORSMiddleware(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("cors_origins", "*")

	gen := NewMiddlewareGenerator()
	err := gen.GenerateCORSMiddleware(tempDir, vars)
	if err != nil {
		t.Fatalf("GenerateCORSMiddleware failed: %v", err)
	}

	corsPath := filepath.Join(tempDir, "internal", "middleware", "cors.go")
	content, err := os.ReadFile(corsPath)
	if err != nil {
		t.Fatalf("Failed to read CORS middleware: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package middleware") {
		t.Error("CORS middleware should have package middleware")
	}
	if !strings.Contains(contentStr, "Access-Control-Allow-Origin") {
		t.Error("CORS middleware should set CORS headers")
	}
}

func TestMiddlewareGenerator_GenerateCustomMiddleware(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/myapp")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
	}

	middlewareSpec := MiddlewareSpec{
		Name:        "RateLimit",
		Description: "Rate limiting middleware",
		Logic:       "// Custom rate limiting logic here",
	}

	gen := NewMiddlewareGenerator()
	err := gen.GenerateCustomMiddleware(tempDir, manifest, vars, middlewareSpec)
	if err != nil {
		t.Fatalf("GenerateCustomMiddleware failed: %v", err)
	}

	middlewarePath := filepath.Join(tempDir, "internal", "middleware", "ratelimit.go")
	content, err := os.ReadFile(middlewarePath)
	if err != nil {
		t.Fatalf("Failed to read custom middleware: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package middleware") {
		t.Error("Custom middleware should have package middleware")
	}
	if !strings.Contains(contentStr, "func RateLimit") {
		t.Error("Custom middleware should have function with correct name")
	}
}

func TestMiddlewareGenerator_GenerateAll(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("module_name", "github.com/test/myapp")
	vars.Set("cors_origins", "*")

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test-ritual",
			Version: "1.0.0",
		},
	}

	gen := NewMiddlewareGenerator()
	err := gen.GenerateAll(tempDir, manifest, vars)
	if err != nil {
		t.Fatalf("GenerateAll failed: %v", err)
	}

	// Check that all standard middleware were created
	expectedFiles := []string{
		"internal/middleware/auth.go",
		"internal/middleware/logging.go",
		"internal/middleware/cors.go",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tempDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}
}

func TestMiddlewareGenerator_WithTemplateVariable(t *testing.T) {
	tempDir := t.TempDir()

	vars := NewVariables()
	vars.Set("cors_origins", "https://example.com")

	gen := NewMiddlewareGenerator()
	err := gen.GenerateCORSMiddleware(tempDir, vars)
	if err != nil {
		t.Fatalf("GenerateCORSMiddleware failed: %v", err)
	}

	corsPath := filepath.Join(tempDir, "internal", "middleware", "cors.go")
	content, err := os.ReadFile(corsPath)
	if err != nil {
		t.Fatalf("Failed to read CORS middleware: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "https://example.com") {
		t.Error("CORS middleware should use cors_origins from variables")
	}
}

package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRouteGenerator_GenerateRoutes(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewRouteGenerator()
	
	config := RouteConfig{
		Package: "routes",
		Routes: []Route{
			{Method: "GET", Path: "/users", Handler: "userHandler.ListUsers"},
			{Method: "POST", Path: "/users", Handler: "userHandler.CreateUser"},
			{Method: "GET", Path: "/users/{id}", Handler: "userHandler.GetUser"},
		},
	}
	
	err := gen.GenerateRoutes(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateRoutes() error = %v", err)
	}
	
	routePath := filepath.Join(tmpDir, "internal", "routes", "routes.go")
	if _, err := os.Stat(routePath); os.IsNotExist(err) {
		t.Error("Route file should be created")
	}
	
	content, _ := os.ReadFile(routePath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "GET") {
		t.Error("Routes should contain GET method")
	}
	
	if !strings.Contains(contentStr, "POST") {
		t.Error("Routes should contain POST method")
	}
	
	if !strings.Contains(contentStr, "/users") {
		t.Error("Routes should contain /users path")
	}
}

func TestRouteGenerator_GenerateRESTfulRoutes(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewRouteGenerator()
	
	config := RouteConfig{
		Resource: "products",
		Handler:  "productHandler",
		RESTful:  true,
	}
	
	err := gen.GenerateRoutes(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateRoutes() error = %v", err)
	}
	
	routePath := filepath.Join(tmpDir, "internal", "routes", "routes.go")
	content, _ := os.ReadFile(routePath)
	contentStr := string(content)
	
	expectedRoutes := []string{
		"GET /products",
		"POST /products",
		"GET /products/{id}",
		"PUT /products/{id}",
		"DELETE /products/{id}",
	}
	
	for _, route := range expectedRoutes {
		parts := strings.Fields(route)
		for _, part := range parts {
			if !strings.Contains(contentStr, part) {
				t.Errorf("Routes should contain %s", route)
				break
			}
		}
	}
}

func TestRouteGenerator_GenerateWithGroups(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewRouteGenerator()
	
	config := RouteConfig{
		Groups: []RouteGroup{
			{
				Prefix: "/api/v1",
				Routes: []Route{
					{Method: "GET", Path: "/users", Handler: "userHandler.ListUsers"},
				},
			},
			{
				Prefix: "/admin",
				Routes: []Route{
					{Method: "GET", Path: "/dashboard", Handler: "adminHandler.Dashboard"},
				},
			},
		},
	}
	
	err := gen.GenerateRoutes(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateRoutes() error = %v", err)
	}
	
	routePath := filepath.Join(tmpDir, "internal", "routes", "routes.go")
	content, _ := os.ReadFile(routePath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "/api/v1") {
		t.Error("Routes should contain /api/v1 prefix")
	}
	
	if !strings.Contains(contentStr, "/admin") {
		t.Error("Routes should contain /admin prefix")
	}
}

func TestRouteGenerator_GenerateWithMiddleware(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewRouteGenerator()
	
	config := RouteConfig{
		Routes: []Route{
			{
				Method:      "POST",
				Path:        "/protected",
				Handler:     "handler.Protected",
				Middlewares: []string{"authMiddleware", "loggingMiddleware"},
			},
		},
	}
	
	err := gen.GenerateRoutes(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateRoutes() error = %v", err)
	}
	
	routePath := filepath.Join(tmpDir, "internal", "routes", "routes.go")
	content, _ := os.ReadFile(routePath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "authMiddleware") {
		t.Error("Routes should contain authMiddleware")
	}
	
	if !strings.Contains(contentStr, "loggingMiddleware") {
		t.Error("Routes should contain loggingMiddleware")
	}
}

func TestRouteGenerator_GenerateDocumentedRoutes(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewRouteGenerator()
	
	config := RouteConfig{
		Routes: []Route{
			{
				Method:      "GET",
				Path:        "/items",
				Handler:     "itemHandler.List",
				Description: "List all items",
			},
		},
		Documentation: true,
	}
	
	err := gen.GenerateRoutes(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateRoutes() error = %v", err)
	}
	
	routePath := filepath.Join(tmpDir, "internal", "routes", "routes.go")
	content, _ := os.ReadFile(routePath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "List all items") {
		t.Error("Routes should contain route documentation")
	}
}

func TestRouteGenerator_GenerateRouterSetup(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewRouteGenerator()
	
	config := RouteConfig{
		Package: "routes",
		Routes: []Route{
			{Method: "GET", Path: "/health", Handler: "healthHandler.Check"},
		},
	}
	
	err := gen.GenerateRoutes(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateRoutes() error = %v", err)
	}
	
	routePath := filepath.Join(tmpDir, "internal", "routes", "routes.go")
	content, _ := os.ReadFile(routePath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "SetupRoutes") {
		t.Error("Routes should contain SetupRoutes function")
	}
	
	if !strings.Contains(contentStr, "mux.NewRouter") {
		t.Error("Routes should initialize router")
	}
}

func TestRouteGenerator_GenerateWithCORS(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewRouteGenerator()
	
	config := RouteConfig{
		EnableCORS: true,
		Routes: []Route{
			{Method: "GET", Path: "/api", Handler: "apiHandler.Index"},
		},
	}
	
	err := gen.GenerateRoutes(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateRoutes() error = %v", err)
	}
	
	routePath := filepath.Join(tmpDir, "internal", "routes", "routes.go")
	content, _ := os.ReadFile(routePath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "CORS") {
		t.Error("Routes should configure CORS")
	}
}

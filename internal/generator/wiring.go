package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// ComponentWiring handles wiring up all project components
type ComponentWiring struct {
	generator *FileGenerator
}

// NewComponentWiring creates a new component wiring generator
func NewComponentWiring() *ComponentWiring {
	return &ComponentWiring{
		generator: NewFileGenerator("go-template"),
	}
}

// WireComponents generates code to wire up all components
func (w *ComponentWiring) WireComponents(projectPath string, manifest *ritual.Manifest, vars *Variables) error {
	// Generate DI container if needed
	if err := w.generateDIContainer(projectPath, manifest, vars); err != nil {
		return fmt.Errorf("failed to generate DI container: %w", err)
	}

	// Generate router setup
	if err := w.generateRouterSetup(projectPath, manifest, vars); err != nil {
		return fmt.Errorf("failed to generate router setup: %w", err)
	}

	// Generate middleware chain
	if err := w.generateMiddlewareChain(projectPath, vars); err != nil {
		return fmt.Errorf("failed to generate middleware chain: %w", err)
	}

	// Update main.go to use wired components
	if err := w.updateMainWithWiring(projectPath, vars); err != nil {
		return fmt.Errorf("failed to update main.go: %w", err)
	}

	return nil
}

// generateDIContainer creates a simple dependency injection container
func (w *ComponentWiring) generateDIContainer(projectPath string, manifest *ritual.Manifest, vars *Variables) error {
	template := `package container

import (
	"database/sql"
	"log"

	"[[ .module_name ]]/internal/handlers"
	"[[ .module_name ]]/internal/repositories"
	"[[ .module_name ]]/internal/services"
)

// Container holds all application dependencies
type Container struct {
	DB *sql.DB
	
	// Repositories
	[[ range .repositories ]][[ .Name ]]Repository repositories.[[ .Type ]]Repository
	[[ end ]]
	
	// Services
	[[ range .services ]][[ .Name ]]Service services.[[ .Type ]]Service
	[[ end ]]
	
	// Handlers
	[[ range .handlers ]][[ .Name ]]Handler *handlers.[[ .Type ]]Handler
	[[ end ]]
}

// NewContainer creates and wires up all dependencies
func NewContainer(db *sql.DB) *Container {
	c := &Container{
		DB: db,
	}
	
	// Initialize repositories
	[[ range .repositories ]]c.[[ .Name ]]Repository = repositories.New[[ .Type ]]Repository(db)
	[[ end ]]
	
	// Initialize services
	[[ range .services ]]c.[[ .Name ]]Service = services.New[[ .Type ]]Service(c.[[ .DependsOn ]]Repository)
	[[ end ]]
	
	// Initialize handlers
	[[ range .handlers ]]c.[[ .Name ]]Handler = handlers.New[[ .Type ]]Handler(c.[[ .DependsOn ]]Service)
	[[ end ]]
	
	log.Println("Dependency container initialized")
	return c
}

// Close cleans up resources
func (c *Container) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
`

	w.generator.SetVariables(vars)

	// Add placeholder data for repositories, services, and handlers
	varsMap := vars.All()
	if varsMap["repositories"] == nil {
		varsMap["repositories"] = []map[string]string{}
	}
	if varsMap["services"] == nil {
		varsMap["services"] = []map[string]string{}
	}
	if varsMap["handlers"] == nil {
		varsMap["handlers"] = []map[string]string{}
	}

	content, err := w.generator.engine.Render(template, varsMap)
	if err != nil {
		return err
	}

	containerDir := filepath.Join(projectPath, "internal", "container")
	if err := os.MkdirAll(containerDir, 0750); err != nil {
		return err
	}

	containerPath := filepath.Join(containerDir, "container.go")
	return os.WriteFile(containerPath, []byte(content), 0600)
}

// generateRouterSetup creates router configuration
func (w *ComponentWiring) generateRouterSetup(projectPath string, manifest *ritual.Manifest, vars *Variables) error {
	template := `package router

import (
	"net/http"

	"[[ .module_name ]]/internal/container"
	"[[ .module_name ]]/internal/middleware"
)

// Setup configures all application routes
func Setup(c *container.Container) *http.ServeMux {
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"status\":\"healthy\"}"))
	})
	
	// API routes
	[[ range .routes ]]mux.HandleFunc("[[ .Path ]]", middleware.Chain(
		c.[[ .Handler ]]Handler.[[ .Method ]],
		middleware.Logger,
		middleware.Recovery,
	))
	[[ end ]]
	
	return mux
}
`

	w.generator.SetVariables(vars)

	varsMap := vars.All()
	if varsMap["routes"] == nil {
		varsMap["routes"] = []map[string]string{}
	}

	content, err := w.generator.engine.Render(template, varsMap)
	if err != nil {
		return err
	}

	routerDir := filepath.Join(projectPath, "internal", "router")
	if err := os.MkdirAll(routerDir, 0750); err != nil {
		return err
	}

	routerPath := filepath.Join(routerDir, "router.go")
	return os.WriteFile(routerPath, []byte(content), 0600)
}

// generateMiddlewareChain creates middleware utilities
func (w *ComponentWiring) generateMiddlewareChain(projectPath string, vars *Variables) error {
	template := `package middleware

import (
	"log"
	"net/http"
	"time"
)

// Middleware is a function that wraps an http.HandlerFunc
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Chain applies middlewares to a handler
func Chain(handler http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// Logger logs HTTP requests
func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	}
}

// Recovery recovers from panics
func Recovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}

// CORS adds CORS headers
func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}
`

	middlewareDir := filepath.Join(projectPath, "internal", "middleware")
	if err := os.MkdirAll(middlewareDir, 0750); err != nil {
		return err
	}

	middlewarePath := filepath.Join(middlewareDir, "middleware.go")
	return os.WriteFile(middlewarePath, []byte(template), 0600)
}

// updateMainWithWiring updates main.go to use the wired components
func (w *ComponentWiring) updateMainWithWiring(projectPath string, vars *Variables) error {
	mainPath := filepath.Join(projectPath, "cmd", "server", "main.go")

	// Check if main.go exists
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		// If main.go doesn't exist yet, skip this step
		return nil
	}

	// Read existing main.go
	// #nosec G304 - mainPath is a validated project file path
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return err
	}

	// Check if it's already wired
	if strings.Contains(string(content), "container.NewContainer") {
		return nil // Already wired
	}

	// Generate new main.go with wiring
	template := `package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"[[ .module_name ]]/internal/container"
	"[[ .module_name ]]/internal/router"
	
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "[[ .port ]]"
	}
	
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	
	// Initialize database connection (if configured)
	var db *sql.DB
	var err error
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr != "" {
		db, err = sql.Open("postgres", dbConnStr)
		if err != nil {
			log.Printf("Warning: Could not connect to database: %v", err)
		} else {
			defer db.Close()
			log.Println("Database connection established")
		}
	}
	
	// Initialize dependency container
	c := container.NewContainer(db)
	defer c.Close()
	
	// Setup routes
	mux := router.Setup(c)
	
	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting [[ .app_name ]] on %s", addr)
	
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
`

	w.generator.SetVariables(vars)

	newContent, err := w.generator.engine.Render(template, vars.All())
	if err != nil {
		return err
	}

	return os.WriteFile(mainPath, []byte(newContent), 0600)
}

package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// MiddlewareGenerator generates middleware components
type MiddlewareGenerator struct {
	generator *FileGenerator
}

// MiddlewareSpec defines a custom middleware specification
type MiddlewareSpec struct {
	Name        string
	Description string
	Logic       string
}

// NewMiddlewareGenerator creates a new middleware generator
func NewMiddlewareGenerator() *MiddlewareGenerator {
	return &MiddlewareGenerator{
		generator: NewFileGenerator("go-template"),
	}
}

// GenerateAuthMiddleware generates authentication middleware
func (m *MiddlewareGenerator) GenerateAuthMiddleware(projectPath string, vars *Variables) error {
	template := `package middleware

import (
	"net/http"
	"strings"
)

// Auth validates authentication tokens
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized: Invalid token format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		
		// Validate token (implement your validation logic here)
		if !validateToken(token) {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Token is valid, proceed to next handler
		next(w, r)
	}
}

// validateToken validates the authentication token
func validateToken(token string) bool {
	// TODO: Implement proper token validation
	// This is a placeholder implementation
	return token != ""
}
`

	middlewareDir := filepath.Join(projectPath, "internal", "middleware")
	if err := os.MkdirAll(middlewareDir, 0750); err != nil {
		return err
	}

	authPath := filepath.Join(middlewareDir, "auth.go")
	return os.WriteFile(authPath, []byte(template), 0600)
}

// GenerateLoggingMiddleware generates request logging middleware
func (m *MiddlewareGenerator) GenerateLoggingMiddleware(projectPath string, vars *Variables) error {
	template := `package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// RequestLogger logs HTTP requests with timing and status information
func RequestLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next(wrapped, r)

		// Log the request
		duration := time.Since(start)
		log.Printf(
			"%s %s %d %d bytes %v",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			wrapped.bytes,
			duration,
		)
	}
}

// DetailedLogger logs requests with more details including headers
func DetailedLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next(wrapped, r)

		duration := time.Since(start)
		log.Printf(
			"%s %s [%s] %d %d bytes %v User-Agent: %s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			wrapped.statusCode,
			wrapped.bytes,
			duration,
			r.UserAgent(),
		)
	}
}
`

	middlewareDir := filepath.Join(projectPath, "internal", "middleware")
	if err := os.MkdirAll(middlewareDir, 0750); err != nil {
		return err
	}

	loggingPath := filepath.Join(middlewareDir, "logging.go")
	return os.WriteFile(loggingPath, []byte(template), 0600)
}

// GenerateCORSMiddleware generates CORS middleware
func (m *MiddlewareGenerator) GenerateCORSMiddleware(projectPath string, vars *Variables) error {
	corsOrigins := vars.GetString("cors_origins")
	if corsOrigins == "" {
		corsOrigins = "*"
	}

	template := `package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"[[ .cors_origins ]]"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Authorization", "X-Requested-With"},
		ExposedHeaders: []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           3600,
	}
}

// CORS creates a CORS middleware with the given configuration
func CORS(config *CORSConfig) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if isOriginAllowed(origin, config.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			// Set other CORS headers
			if len(config.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			}

			if len(config.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			}

			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}
}

// SimpleCORS creates a simple permissive CORS middleware
func SimpleCORS(next http.HandlerFunc) http.HandlerFunc {
	return CORS(DefaultCORSConfig())(next)
}

// isOriginAllowed checks if an origin is in the allowed list
func isOriginAllowed(origin string, allowed []string) bool {
	for _, o := range allowed {
		if o == origin || o == "*" {
			return true
		}
	}
	return false
}
`

	m.generator.SetVariables(vars)
	varsMap := vars.All()
	varsMap["cors_origins"] = corsOrigins

	content, err := m.generator.engine.Render(template, varsMap)
	if err != nil {
		return err
	}

	middlewareDir := filepath.Join(projectPath, "internal", "middleware")
	if err := os.MkdirAll(middlewareDir, 0750); err != nil {
		return err
	}

	// Add missing import
	if !strings.Contains(content, `"fmt"`) {
		content = strings.Replace(content, `import (
	"net/http"`, `import (
	"fmt"
	"net/http"`, 1)
	}

	corsPath := filepath.Join(middlewareDir, "cors.go")
	return os.WriteFile(corsPath, []byte(content), 0600)
}

// GenerateCustomMiddleware generates a custom middleware from specification
func (m *MiddlewareGenerator) GenerateCustomMiddleware(projectPath string, manifest *ritual.Manifest, vars *Variables, spec MiddlewareSpec) error {
	template := fmt.Sprintf(`package middleware

import (
	"net/http"
)

// %s %s
func %s(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		%s
		
		next(w, r)
	}
}
`, spec.Name, spec.Description, spec.Name, spec.Logic)

	middlewareDir := filepath.Join(projectPath, "internal", "middleware")
	if err := os.MkdirAll(middlewareDir, 0750); err != nil {
		return err
	}

	filename := strings.ToLower(spec.Name) + ".go"
	middlewarePath := filepath.Join(middlewareDir, filename)
	return os.WriteFile(middlewarePath, []byte(template), 0600)
}

// GenerateAll generates all standard middleware
func (m *MiddlewareGenerator) GenerateAll(projectPath string, manifest *ritual.Manifest, vars *Variables) error {
	// Generate auth middleware
	if err := m.GenerateAuthMiddleware(projectPath, vars); err != nil {
		return fmt.Errorf("failed to generate auth middleware: %w", err)
	}

	// Generate logging middleware
	if err := m.GenerateLoggingMiddleware(projectPath, vars); err != nil {
		return fmt.Errorf("failed to generate logging middleware: %w", err)
	}

	// Generate CORS middleware
	if err := m.GenerateCORSMiddleware(projectPath, vars); err != nil {
		return fmt.Errorf("failed to generate CORS middleware: %w", err)
	}

	return nil
}

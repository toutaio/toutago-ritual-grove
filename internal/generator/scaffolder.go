package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// ProjectScaffolder creates project structure and generates files
type ProjectScaffolder struct {
	generator *FileGenerator
}

// NewProjectScaffolder creates a new project scaffolder
func NewProjectScaffolder() *ProjectScaffolder {
	return &ProjectScaffolder{
		generator: NewFileGenerator("go-template"),
	}
}

// CreateStructure creates the standard project directory structure
func (s *ProjectScaffolder) CreateStructure(projectPath string) error {
	dirs := []string{
		"cmd/server",
		"internal/handlers",
		"internal/models",
		"internal/repositories",
		"internal/middleware",
		"internal/services",
		"pkg",
		"config",
		"docs",
		"test",
		"migrations",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GenerateMainGo generates the main.go entry point
func (s *ProjectScaffolder) GenerateMainGo(projectPath string, vars *Variables) error {
	template := `package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"{{ .module_name }}/internal/handlers"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "{{ .port }}"
	}

	// Initialize router
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/health", handlers.HealthCheck)

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting {{ .app_name }} on %s", addr)
	
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
`

	s.generator.SetVariables(vars)

	content, err := s.generator.engine.Render(template, vars.All())
	if err != nil {
		return fmt.Errorf("failed to render main.go: %w", err)
	}

	mainPath := filepath.Join(projectPath, "cmd", "server", "main.go")
	if err := os.WriteFile(mainPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	// Also generate a basic health check handler
	return s.generateHealthHandler(projectPath)
}

func (s *ProjectScaffolder) generateHealthHandler(projectPath string) error {
	template := `package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthCheck handles health check requests
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "healthy",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
`

	handlerPath := filepath.Join(projectPath, "internal", "handlers", "health.go")
	return os.WriteFile(handlerPath, []byte(template), 0644)
}

// GenerateGoMod generates the go.mod file with dependencies
func (s *ProjectScaffolder) GenerateGoMod(projectPath string, manifest *ritual.Manifest, vars *Variables) error {
	moduleName := vars.GetString("module_name")
	if moduleName == "" {
		moduleName = "example.com/app"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("module %s\n\n", moduleName))
	sb.WriteString("go 1.21\n\n")

	if len(manifest.Dependencies.Packages) > 0 {
		sb.WriteString("require (\n")
		for _, pkg := range manifest.Dependencies.Packages {
			// Add version if not specified (use latest for now)
			if !strings.Contains(pkg, "@") {
				pkg = pkg + " v1.0.0"
			}
			sb.WriteString(fmt.Sprintf("\t%s\n", pkg))
		}
		sb.WriteString(")\n")
	}

	goModPath := filepath.Join(projectPath, "go.mod")
	return os.WriteFile(goModPath, []byte(sb.String()), 0644)
}

// GenerateConfig generates configuration files (.env.example, config files)
func (s *ProjectScaffolder) GenerateConfig(projectPath string, vars *Variables) error {
	template := `# Application Configuration
APP_NAME={{ .app_name }}
PORT={{ .port }}
ENV=development

# Database (if applicable)
# DB_HOST=localhost
# DB_PORT=5432
# DB_NAME={{ .app_name }}
# DB_USER=postgres
# DB_PASSWORD=

# Logging
LOG_LEVEL=info
`

	s.generator.SetVariables(vars)

	content, err := s.generator.engine.Render(template, vars.All())
	if err != nil {
		return fmt.Errorf("failed to render .env.example: %w", err)
	}

	envPath := filepath.Join(projectPath, ".env.example")
	return os.WriteFile(envPath, []byte(content), 0644)
}

// GenerateREADME generates a README.md file
func (s *ProjectScaffolder) GenerateREADME(projectPath string, manifest *ritual.Manifest, vars *Variables) error {
	appName := vars.GetString("app_name")
	if appName == "" {
		appName = "Application"
	}

	content := fmt.Sprintf("# %s\n\n%s\n\n", appName, manifest.Ritual.Description)
	content += "## Getting Started\n\n"
	content += "### Prerequisites\n\n"
	content += "- Go 1.21 or higher\n"
	content += "- Git\n\n"
	content += "### Installation\n\n"
	content += "```bash\n"
	content += "# Clone the repository\n"
	content += "git clone <repository-url>\n"
	content += fmt.Sprintf("cd %s\n\n", appName)
	content += "# Install dependencies\n"
	content += "go mod download\n\n"
	content += "# Run the application\n"
	content += "go run cmd/server/main.go\n"
	content += "```\n\n"
	content += "### Configuration\n\n"
	content += "Copy `.env.example` to `.env` and configure your environment variables:\n\n"
	content += "```bash\n"
	content += "cp .env.example .env\n"
	content += "```\n\n"
	content += "### Running\n\n"
	content += "```bash\n"
	content += "go run cmd/server/main.go\n"
	content += "```\n\n"
	content += "The server will start and listen for requests.\n\n"
	content += "### Testing\n\n"
	content += "```bash\n"
	content += "go test ./...\n"
	content += "```\n\n"
	content += "### Building\n\n"
	content += "```bash\n"
	content += "go build -o bin/app cmd/server/main.go\n"
	content += "```\n\n"
	content += "## Project Structure\n\n"
	content += "- `cmd/server/` - Application entry point\n"
	content += "- `internal/` - Internal packages\n"
	content += "  - `handlers/` - HTTP handlers\n"
	content += "  - `models/` - Data models\n"
	content += "  - `repositories/` - Data access layer\n"
	content += "  - `middleware/` - HTTP middleware\n"
	content += "  - `services/` - Business logic\n"
	content += "- `pkg/` - Public packages\n"
	content += "- `config/` - Configuration files\n"
	content += "- `docs/` - Documentation\n"
	content += "- `test/` - Test files\n"
	content += "- `migrations/` - Database migrations\n\n"
	content += "## License\n\n"
	content += "MIT\n"

	readmePath := filepath.Join(projectPath, "README.md")
	return os.WriteFile(readmePath, []byte(content), 0644)
}

// GenerateGitignore generates a .gitignore file
func (s *ProjectScaffolder) GenerateGitignore(projectPath string) error {
	content := `# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Test coverage
*.out
coverage.html

# Dependencies
vendor/

# Environment
.env
.env.local
.env.*.local

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Build artifacts
dist/
build/

# Logs
*.log
logs/
`

	gitignorePath := filepath.Join(projectPath, ".gitignore")
	return os.WriteFile(gitignorePath, []byte(content), 0644)
}

// ApplyTemplateFiles applies template files from the ritual
func (s *ProjectScaffolder) ApplyTemplateFiles(projectPath, ritualPath string, manifest *ritual.Manifest, vars *Variables) error {
	s.generator.SetVariables(vars)

	templatesDir := filepath.Join(ritualPath, "templates")

	// Process template files
	for _, fileMapping := range manifest.Files.Templates {
		srcPath := filepath.Join(templatesDir, fileMapping.Source)

		// Render destination path (may contain template variables)
		destPathRendered, err := s.generator.engine.Render(fileMapping.Destination, vars.All())
		if err != nil {
			return fmt.Errorf("failed to render destination path %s: %w", fileMapping.Destination, err)
		}
		destPath := filepath.Join(projectPath, destPathRendered)

		// Check if source exists
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			if fileMapping.Optional {
				continue
			}
			return fmt.Errorf("template file not found: %s", srcPath)
		}

		// Generate file (templates are always rendered)
		if err := s.generator.GenerateFile(srcPath, destPath, true); err != nil {
			return fmt.Errorf("failed to generate %s: %w", destPathRendered, err)
		}
	}

	// Process static files
	for _, fileMapping := range manifest.Files.Static {
		srcPath := filepath.Join(templatesDir, fileMapping.Source)

		// Render destination path (may contain template variables)
		destPathRendered, err := s.generator.engine.Render(fileMapping.Destination, vars.All())
		if err != nil {
			return fmt.Errorf("failed to render destination path %s: %w", fileMapping.Destination, err)
		}
		destPath := filepath.Join(projectPath, destPathRendered)

		// Check if source exists
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			if fileMapping.Optional {
				continue
			}
			return fmt.Errorf("static file not found: %s", srcPath)
		}

		// Copy static file (no template rendering)
		if err := s.generator.GenerateFile(srcPath, destPath, false); err != nil {
			return fmt.Errorf("failed to copy %s: %w", destPathRendered, err)
		}
	}

	return nil
}

// GenerateFromRitual generates a complete project from a ritual
func (s *ProjectScaffolder) GenerateFromRitual(projectPath, ritualPath string, manifest *ritual.Manifest, vars *Variables) error {
	// Create project structure
	if err := s.CreateStructure(projectPath); err != nil {
		return err
	}

	// Generate go.mod
	if err := s.GenerateGoMod(projectPath, manifest, vars); err != nil {
		return err
	}

	// Generate main.go (if not provided by ritual)
	hasMainGo := false
	for _, tmpl := range manifest.Files.Templates {
		if strings.Contains(tmpl.Destination, "main.go") {
			hasMainGo = true
			break
		}
	}
	if !hasMainGo {
		if err := s.GenerateMainGo(projectPath, vars); err != nil {
			return err
		}
	}

	// Generate README (if not provided by ritual)
	hasREADME := false
	for _, tmpl := range manifest.Files.Templates {
		if strings.ToUpper(tmpl.Destination) == "README.MD" {
			hasREADME = true
			break
		}
	}
	if !hasREADME {
		if err := s.GenerateREADME(projectPath, manifest, vars); err != nil {
			return err
		}
	}

	// Generate .gitignore (if not provided by ritual)
	hasGitignore := false
	for _, tmpl := range manifest.Files.Templates {
		if tmpl.Destination == ".gitignore" {
			hasGitignore = true
			break
		}
	}
	if !hasGitignore {
		if err := s.GenerateGitignore(projectPath); err != nil {
			return err
		}
	}

	// Generate .env.example (if not provided by ritual)
	hasEnvExample := false
	for _, tmpl := range manifest.Files.Templates {
		if tmpl.Destination == ".env.example" {
			hasEnvExample = true
			break
		}
	}
	if !hasEnvExample {
		if err := s.GenerateConfig(projectPath, vars); err != nil {
			return err
		}
	}

	// Apply ritual template files
	if err := s.ApplyTemplateFiles(projectPath, ritualPath, manifest, vars); err != nil {
		return err
	}

	return nil
}

// ExecutePostGenerateHooks executes post-generation hooks
func (s *ProjectScaffolder) ExecutePostGenerateHooks(projectPath string, hookCommands []string) error {
	if len(hookCommands) == 0 {
		return nil
	}

	hookExecutor := hooks.NewHookExecutor(projectPath)
	return hookExecutor.ExecutePostInstall(hookCommands)
}

// GenerateFromRitualWithHooks generates a project and executes hooks
func (s *ProjectScaffolder) GenerateFromRitualWithHooks(projectPath, ritualPath string, manifest *ritual.Manifest, vars *Variables) error {
	// Execute pre-install hooks
	if len(manifest.Hooks.PreInstall) > 0 {
		hookExecutor := hooks.NewHookExecutor(projectPath)
		if err := hookExecutor.ExecutePreInstall(manifest.Hooks.PreInstall); err != nil {
			return fmt.Errorf("pre-install hooks failed: %w", err)
		}
	}

	// Generate project
	if err := s.GenerateFromRitual(projectPath, ritualPath, manifest, vars); err != nil {
		return err
	}

	// Execute post-install hooks
	if len(manifest.Hooks.PostInstall) > 0 {
		hookExecutor := hooks.NewHookExecutor(projectPath)
		if err := hookExecutor.ExecutePostInstall(manifest.Hooks.PostInstall); err != nil {
			return fmt.Errorf("post-install hooks failed: %w", err)
		}
	}

	return nil
}

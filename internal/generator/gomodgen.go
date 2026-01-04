package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// GoModGenerator generates and manages go.mod files
type GoModGenerator struct {
	generator *FileGenerator
}

// NewGoModGenerator creates a new go.mod generator
func NewGoModGenerator() *GoModGenerator {
	return &GoModGenerator{
		generator: NewFileGenerator("go-template"),
	}
}

// Generate creates a go.mod file with ritual dependencies
func (g *GoModGenerator) Generate(projectPath string, manifest *ritual.Manifest, vars *Variables) error {
	moduleName := vars.GetString("module_name")
	if moduleName == "" {
		moduleName = "example.com/app"
	}

	goVersion := vars.GetString("go_version")
	if goVersion == "" {
		goVersion = "1.21"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("module %s\n\n", moduleName))
	sb.WriteString(fmt.Sprintf("go %s\n\n", goVersion))

	// Collect all dependencies
	deps := make(map[string]string)

	// Add ritual-specified packages
	for _, pkg := range manifest.Dependencies.Packages {
		name, version := parsePackageVersion(pkg)
		deps[name] = version
	}

	// Add database driver if specified
	if manifest.Dependencies.Database != nil && len(manifest.Dependencies.Database.Types) > 0 {
		// Use the first database type specified
		driver := getDatabaseDriver(manifest.Dependencies.Database.Types[0])
		if driver != "" {
			deps[driver] = "latest"
		}
	}

	// Write require block if there are dependencies
	if len(deps) > 0 {
		sb.WriteString("require (\n")
		for name, version := range deps {
			if version == "latest" || version == "" {
				// For latest, we'll let go mod tidy figure it out
				sb.WriteString(fmt.Sprintf("\t%s\n", name))
			} else {
				sb.WriteString(fmt.Sprintf("\t%s %s\n", name, version))
			}
		}
		sb.WriteString(")\n")
	}

	goModPath := filepath.Join(projectPath, "go.mod")
	return os.WriteFile(goModPath, []byte(sb.String()), 0644)
}

// AddToutaDependencies adds Toutā framework dependencies to go.mod
func (g *GoModGenerator) AddToutaDependencies(projectPath string, vars *Variables) error {
	toutaVersion := vars.GetString("touta_version")
	if toutaVersion == "" {
		toutaVersion = "v0.1.0"
	}

	// Read existing go.mod
	goModPath := filepath.Join(projectPath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		// If go.mod doesn't exist, create it first
		moduleName := vars.GetString("module_name")
		if moduleName == "" {
			moduleName = "example.com/app"
		}
		content = []byte(fmt.Sprintf("module %s\n\ngo 1.21\n\n", moduleName))
	}

	contentStr := string(content)

	// Check if Touta dependencies are already added
	if strings.Contains(contentStr, "github.com/toutaio/toutago") {
		return nil // Already added
	}

	// Add require block if it doesn't exist
	if !strings.Contains(contentStr, "require (") {
		contentStr += "\nrequire (\n)\n"
	}

	// Add Touta dependencies
	toutaDeps := []string{
		fmt.Sprintf("github.com/toutaio/toutago %s", toutaVersion),
		fmt.Sprintf("github.com/toutaio/toutago-cosan-router %s", toutaVersion),
		fmt.Sprintf("github.com/toutaio/toutago-nasc-dependency-injector %s", toutaVersion),
	}

	// Insert dependencies before the closing parenthesis
	for _, dep := range toutaDeps {
		contentStr = strings.Replace(contentStr, "require (", fmt.Sprintf("require (\n\t%s", dep), 1)
	}

	return os.WriteFile(goModPath, []byte(contentStr), 0644)
}

// RunGoModTidy runs `go mod tidy` to update dependencies
func (g *GoModGenerator) RunGoModTidy(projectPath string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod tidy failed: %w\nOutput: %s", err, string(output))
	}
	
	return nil
}

// RunGoModDownload runs `go mod download` to download dependencies
func (g *GoModGenerator) RunGoModDownload(projectPath string) error {
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = projectPath
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod download failed: %w\nOutput: %s", err, string(output))
	}
	
	return nil
}

// parsePackageVersion parses a package string into name and version
// Examples:
//   - "github.com/pkg/errors@v0.9.1" -> ("github.com/pkg/errors", "v0.9.1")
//   - "github.com/pkg/errors" -> ("github.com/pkg/errors", "")
func parsePackageVersion(pkg string) (string, string) {
	if strings.Contains(pkg, "@") {
		parts := strings.SplitN(pkg, "@", 2)
		return parts[0], parts[1]
	}
	return pkg, ""
}

// getDatabaseDriver returns the Go package for a database driver
func getDatabaseDriver(dbType string) string {
	switch strings.ToLower(dbType) {
	case "postgresql", "postgres":
		return "github.com/lib/pq"
	case "mysql":
		return "github.com/go-sql-driver/mysql"
	case "sqlite", "sqlite3":
		return "github.com/mattn/go-sqlite3"
	case "sqlserver", "mssql":
		return "github.com/denisenkom/go-mssqldb"
	default:
		return ""
	}
}

// GenerateComplete generates a complete go.mod with all dependencies and runs tidy
func (g *GoModGenerator) GenerateComplete(projectPath string, manifest *ritual.Manifest, vars *Variables) error {
	// Generate base go.mod
	if err := g.Generate(projectPath, manifest, vars); err != nil {
		return fmt.Errorf("failed to generate go.mod: %w", err)
	}

	// Add Touta dependencies if needed
	if vars.GetBool("include_touta") {
		if err := g.AddToutaDependencies(projectPath, vars); err != nil {
			return fmt.Errorf("failed to add Touta dependencies: %w", err)
		}
	}

	// Run go mod tidy if requested
	if vars.GetBool("run_go_mod_tidy") {
		if err := g.RunGoModTidy(projectPath); err != nil {
			// Log warning but don't fail - environment might not have network
			fmt.Printf("Warning: go mod tidy failed: %v\n", err)
		}
	}

	return nil
}

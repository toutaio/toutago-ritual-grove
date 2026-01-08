package generator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// FileGenerator handles file generation from templates
type FileGenerator struct {
	engine          TemplateEngine
	variables       *Variables
	protected       map[string]bool
	ritualsBasePath string // Base path for rituals directory (for _shared access)
}

// NewFileGenerator creates a new file generator
func NewFileGenerator(engineType string) *FileGenerator {
	return &FileGenerator{
		engine:    NewTemplateEngine(engineType),
		variables: NewVariables(),
		protected: make(map[string]bool),
	}
}

// SetVariables sets the variables for template rendering
func (g *FileGenerator) SetVariables(vars *Variables) {
	g.variables = vars
}

// SetProtectedFiles sets files that should not be overwritten
func (g *FileGenerator) SetProtectedFiles(files []string) {
	g.protected = make(map[string]bool)
	for _, file := range files {
		g.protected[file] = true
	}
}

// SetRitualsBasePath sets the base path for the rituals directory
// This is needed to access shared templates in _shared/
func (g *FileGenerator) SetRitualsBasePath(path string) {
	g.ritualsBasePath = path
}

// GenerateFile generates a single file from a template
func (g *FileGenerator) GenerateFile(srcPath, destPath string, isTemplate bool) error {
	// Normalize destPath for comparison with protected files
	normalizedDest := filepath.ToSlash(destPath)

	// Check if destination is protected and exists
	for protectedPath := range g.protected {
		normalizedProtected := filepath.ToSlash(protectedPath)
		if normalizedDest == normalizedProtected || filepath.Base(normalizedDest) == normalizedProtected {
			if _, err := os.Stat(destPath); err == nil {
				// File exists and is protected, skip
				return nil
			}
		}
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	if isTemplate {
		// Read template file
		// #nosec G304 - srcPath is from validated ritual template source
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", srcPath, err)
		}

		// Render template
		rendered, err := g.engine.Render(string(content), g.variables.All())
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", srcPath, err)
		}

		// Write rendered content
		if err := os.WriteFile(destPath, []byte(rendered), 0600); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}
	} else {
		// Copy static file
		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy file %s to %s: %w", srcPath, destPath, err)
		}
	}

	return nil
}

// GenerateFiles generates all files from a manifest
func (g *FileGenerator) GenerateFiles(manifest *ritual.Manifest, ritualPath, outputPath string) error {
	// Set protected files
	g.SetProtectedFiles(manifest.Files.Protected)

	// Generate template files
	for _, tmpl := range manifest.Files.Templates {
		// Evaluate condition if present
		if tmpl.Condition != "" {
			shouldGenerate, err := evaluateCondition(tmpl.Condition, g.variables.All())
			if err != nil {
				return fmt.Errorf("failed to evaluate condition for %s: %w", tmpl.Source, err)
			}
			if !shouldGenerate {
				continue // Skip this file
			}
		}

		// Resolve source path (handle _shared: prefix) - templates go in templates/
		srcPath := g.resolveTemplateSourcePath(tmpl.Source, ritualPath)

		// Render destination path (it may contain template variables)
		destPathRendered, err := g.engine.Render(tmpl.Destination, g.variables.All())
		if err != nil {
			return fmt.Errorf("failed to render destination path %s: %w", tmpl.Destination, err)
		}
		destPath := filepath.Join(outputPath, destPathRendered)

		// Check if file/directory exists
		info, err := os.Stat(srcPath)
		if err != nil {
			if tmpl.Optional {
				continue
			}
			return fmt.Errorf("template source not found: %s", srcPath)
		}

		if info.IsDir() {
			// Generate all files in directory
			if err := g.generateDirectory(srcPath, destPath, true); err != nil {
				return err
			}
		} else {
			// Generate single file
			if err := g.GenerateFile(srcPath, destPath, true); err != nil {
				return err
			}
		}
	}

	// Copy static files
	for _, static := range manifest.Files.Static {
		// Evaluate condition if present
		if static.Condition != "" {
			shouldGenerate, err := evaluateCondition(static.Condition, g.variables.All())
			if err != nil {
				return fmt.Errorf("failed to evaluate condition for %s: %w", static.Source, err)
			}
			if !shouldGenerate {
				continue // Skip this file
			}
		}

		// Resolve source path (handle _shared: prefix) - static files go in static/
		srcPath := g.resolveStaticSourcePath(static.Source, ritualPath)

		// Render destination path (it may contain template variables)
		destPathRendered, err := g.engine.Render(static.Destination, g.variables.All())
		if err != nil {
			return fmt.Errorf("failed to render destination path %s: %w", static.Destination, err)
		}
		destPath := filepath.Join(outputPath, destPathRendered)

		// Check if file/directory exists
		info, err := os.Stat(srcPath)
		if err != nil {
			if static.Optional {
				continue
			}
			return fmt.Errorf("static source not found: %s", srcPath)
		}

		if info.IsDir() {
			// Copy all files in directory
			if err := g.generateDirectory(srcPath, destPath, false); err != nil {
				return err
			}
		} else {
			// Copy single file
			if err := g.GenerateFile(srcPath, destPath, false); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateDirectory generates all files in a directory
func (g *FileGenerator) generateDirectory(srcDir, destDir string, isTemplate bool) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Strip .tmpl extension for template files
		destPath := filepath.Join(destDir, relPath)
		if isTemplate && strings.HasSuffix(destPath, ".tmpl") {
			destPath = strings.TrimSuffix(destPath, ".tmpl")
		}

		return g.GenerateFile(path, destPath, isTemplate)
	})
}

// resolveSourcePath resolves a source path, handling _shared: prefix
// Deprecated: Use resolveTemplateSourcePath or resolveStaticSourcePath instead
func (g *FileGenerator) resolveSourcePath(source, ritualPath string) string {
	// Check if source uses _shared: prefix
	if strings.HasPrefix(source, "_shared:") {
		// Extract the path after _shared:
		sharedPath := strings.TrimPrefix(source, "_shared:")
		
		// If ritualsBasePath is set, use it; otherwise try to find it from ritualPath
		if g.ritualsBasePath != "" {
			return filepath.Join(g.ritualsBasePath, "_shared", sharedPath)
		}
		
		// Fallback: try to find rituals directory from ritualPath
		// ritualPath is typically: /path/to/rituals/ritual-name
		// We need: /path/to/rituals/_shared/...
		ritualsDir := filepath.Dir(ritualPath) // Go up one level
		return filepath.Join(ritualsDir, "_shared", sharedPath)
	}
	
	// Regular path - assume it's in the ritual root
	return filepath.Join(ritualPath, source)
}

// resolveTemplateSourcePath resolves a template source path, handling _shared: prefix
func (g *FileGenerator) resolveTemplateSourcePath(source, ritualPath string) string {
	// Check if source uses _shared: prefix
	if strings.HasPrefix(source, "_shared:") {
		// Extract the path after _shared:
		sharedPath := strings.TrimPrefix(source, "_shared:")
		
		// If ritualsBasePath is set, use it; otherwise try to find it from ritualPath
		if g.ritualsBasePath != "" {
			return filepath.Join(g.ritualsBasePath, "_shared", sharedPath)
		}
		
		// Fallback: try to find rituals directory from ritualPath
		ritualsDir := filepath.Dir(ritualPath)
		return filepath.Join(ritualsDir, "_shared", sharedPath)
	}
	
	// Regular template - prepend templates/ subdirectory
	return filepath.Join(ritualPath, "templates", source)
}

// resolveStaticSourcePath resolves a static file source path, handling _shared: prefix
func (g *FileGenerator) resolveStaticSourcePath(source, ritualPath string) string {
	// Check if source uses _shared: prefix
	if strings.HasPrefix(source, "_shared:") {
		// Extract the path after _shared:
		sharedPath := strings.TrimPrefix(source, "_shared:")
		
		// If ritualsBasePath is set, use it; otherwise try to find it from ritualPath
		if g.ritualsBasePath != "" {
			return filepath.Join(g.ritualsBasePath, "_shared", sharedPath)
		}
		
		// Fallback: try to find rituals directory from ritualPath
		ritualsDir := filepath.Dir(ritualPath)
		return filepath.Join(ritualsDir, "_shared", sharedPath)
	}
	
	// Regular static file - prepend static/ subdirectory
	return filepath.Join(ritualPath, "static", source)
}

// CreateDirectoryStructure creates the directory structure for a project
func (g *FileGenerator) CreateDirectoryStructure(basePath string, dirs []string) error {
	for _, dir := range dirs {
		fullPath := filepath.Join(basePath, dir)
		if err := os.MkdirAll(fullPath, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
	}
	return nil
}

// copyFile copies a file from src to dst
// #nosec G304 - src is from validated ritual source
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			// Log but don't fail on close error
		}
		// #nosec G304 - dst is a validated destination path
	}()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			// Log but don't fail on close error
		}
	}()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

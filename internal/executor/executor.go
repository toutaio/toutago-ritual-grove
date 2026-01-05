package executor

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/internal/generator"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// ExecutionContext holds the context for ritual execution
type ExecutionContext struct {
	RitualPath string
	OutputPath string
	Variables  *generator.Variables
	DryRun     bool
	Logger     *log.Logger
}

// Executor executes ritual installation steps
type Executor struct {
	context   *ExecutionContext
	generator *generator.FileGenerator
	resolver  *DependencyResolver
}

// NewExecutor creates a new ritual executor
func NewExecutor(context *ExecutionContext) *Executor {
	if context.Logger == nil {
		context.Logger = log.New(os.Stdout, "[ritual] ", log.LstdFlags)
	}

	return &Executor{
		context:   context,
		generator: generator.NewFileGenerator("fith"),
		resolver:  NewDependencyResolver(),
	}
}

// Execute executes a ritual installation
func (e *Executor) Execute(manifest *ritual.Manifest) error {
	e.context.Logger.Printf("Starting ritual installation: %s v%s",
		manifest.Ritual.Name, manifest.Ritual.Version)

	// Step 1: Validate dependencies
	if err := e.validateDependencies(manifest); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	// Step 2: Run pre-install hooks
	if err := e.runHooks(manifest.Hooks.PreInstall, "pre-install"); err != nil {
		return fmt.Errorf("pre-install hooks failed: %w", err)
	}

	// Step 3: Generate files
	if err := e.generateFiles(manifest); err != nil {
		return fmt.Errorf("file generation failed: %w", err)
	}

	// Step 4: Install Go dependencies
	if err := e.installPackages(manifest); err != nil {
		return fmt.Errorf("package installation failed: %w", err)
	}

	// Step 5: Run post-install hooks
	if err := e.runHooks(manifest.Hooks.PostInstall, "post-install"); err != nil {
		return fmt.Errorf("post-install hooks failed: %w", err)
	}

	e.context.Logger.Printf("Ritual installation completed successfully")
	return nil
}

func (e *Executor) validateDependencies(manifest *ritual.Manifest) error {
	e.context.Logger.Println("Validating dependencies...")

	if e.context.DryRun {
		e.context.Logger.Println("[DRY RUN] Would validate dependencies")
		return nil
	}

	return e.resolver.ValidateDependencies(manifest)
}

func (e *Executor) generateFiles(manifest *ritual.Manifest) error {
	e.context.Logger.Println("Generating files...")

	if e.context.DryRun {
		e.context.Logger.Println("[DRY RUN] Would generate files:")
		for _, tmpl := range manifest.Files.Templates {
			e.context.Logger.Printf("  - Template: %s -> %s", tmpl.Source, tmpl.Destination)
		}
		for _, static := range manifest.Files.Static {
			e.context.Logger.Printf("  - Static: %s -> %s", static.Source, static.Destination)
		}
		return nil
	}

	e.generator.SetVariables(e.context.Variables)
	return e.generator.GenerateFiles(manifest, e.context.RitualPath, e.context.OutputPath)
}

func (e *Executor) installPackages(manifest *ritual.Manifest) error {
	if len(manifest.Dependencies.Packages) == 0 {
		return nil
	}

	e.context.Logger.Printf("Installing %d Go packages...", len(manifest.Dependencies.Packages))

	if e.context.DryRun {
		e.context.Logger.Println("[DRY RUN] Would install packages:")
		for _, pkg := range manifest.Dependencies.Packages {
			e.context.Logger.Printf("  - %s", pkg)
		}
		return nil
	}

	// Run go get for each package
	for _, pkg := range manifest.Dependencies.Packages {
		e.context.Logger.Printf("  Installing %s...", pkg)

		// #nosec G204 - go get command with validated package name
		cmd := exec.Command("go", "get", pkg)
		cmd.Dir = e.context.OutputPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install %s: %w", pkg, err)
		}
	}

	return nil
}

func (e *Executor) runHooks(hooks []string, phase string) error {
	if len(hooks) == 0 {
		return nil
	}

	e.context.Logger.Printf("Running %s hooks (%d)...", phase, len(hooks))

	for i, command := range hooks {
		e.context.Logger.Printf("  Hook %d/%d: %s", i+1, len(hooks), command)

		if e.context.DryRun {
			e.context.Logger.Printf("  [DRY RUN] Would run: %s", command)
			continue
		}

		if err := e.executeCommand(command); err != nil {
			return fmt.Errorf("hook command '%s' failed: %w", command, err)
		}
	}

	return nil
}

func (e *Executor) executeCommand(command string) error {
	// Parse command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	// #nosec G204 - Hook command is from validated ritual manifest

	cmd := exec.Command(parts[0], parts[1:]...) // #nosec G204 - command from trusted ritual manifest
	cmd.Dir = e.context.OutputPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}

// Rollback attempts to rollback a failed installation
func (e *Executor) Rollback() error {
	e.context.Logger.Println("Rolling back installation...")

	if e.context.DryRun {
		e.context.Logger.Println("[DRY RUN] Would rollback changes")
		return nil
	}

	// For now, just log
	// TODO: Implement proper rollback (remove generated files, restore backups)
	e.context.Logger.Println("Rollback not yet implemented")

	return nil
}

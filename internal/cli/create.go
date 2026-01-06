// Package cli provides command-line interface workflows
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/toutaio/toutago-ritual-grove/internal/generator"
	"github.com/toutaio/toutago-ritual-grove/internal/questionnaire"
	"github.com/toutaio/toutago-ritual-grove/internal/storage"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// CreateOptions holds options for project creation
type CreateOptions struct {
	RitualPath string
	TargetPath string
	Answers    map[string]interface{}
	DryRun     bool
	InitGit    bool
}

// CreateWorkflow manages the project creation process
type CreateWorkflow struct {
	scaffolder *generator.ProjectScaffolder
}

// NewCreateWorkflow creates a new create workflow
func NewCreateWorkflow() *CreateWorkflow {
	return &CreateWorkflow{
		scaffolder: generator.NewProjectScaffolder(),
	}
}

// Execute runs the create workflow (backward compatible)
func Execute(ritualPath, targetPath string, answers map[string]interface{}, dryRun bool) error {
	workflow := NewCreateWorkflow()
	return workflow.Execute(ritualPath, targetPath, answers, dryRun)
}

// Execute runs the complete project creation workflow
func (w *CreateWorkflow) Execute(ritualPath, targetPath string, answers map[string]interface{}, dryRun bool) error {
	return w.ExecuteWithOptions(CreateOptions{
		RitualPath: ritualPath,
		TargetPath: targetPath,
		Answers:    answers,
		DryRun:     dryRun,
		InitGit:    false,
	})
}

// ExecuteWithOptions runs the complete project creation workflow with options
func (w *CreateWorkflow) ExecuteWithOptions(opts CreateOptions) error {
	// Load ritual
	loader := ritual.NewLoader(opts.RitualPath)
	manifest, err := loader.Load(opts.RitualPath)
	if err != nil {
		return fmt.Errorf("failed to load ritual: %w", err)
	}

	// If no answers provided, run questionnaire
	answers := opts.Answers
	if answers == nil {
		if opts.DryRun {
			return fmt.Errorf("dry-run mode requires answers to be provided")
		}

		adapter := questionnaire.NewCLIAdapter(manifest.Questions, os.Stdin)

		answers, err = adapter.Run()
		if err != nil {
			return fmt.Errorf("questionnaire failed: %w", err)
		}
	}

	// Convert answers to Variables
	vars := generator.NewVariables()
	for key, value := range answers {
		vars.Set(key, value)
	}

	// Dry run mode - just validate, don't create files
	if opts.DryRun {
		fmt.Println("DRY RUN MODE - No files will be created")
		fmt.Printf("Would create project at: %s\n", opts.TargetPath)
		fmt.Printf("Using ritual: %s v%s\n", manifest.Ritual.Name, manifest.Ritual.Version)
		fmt.Println("\nAnswers:")
		for key, value := range answers {
			fmt.Printf("  %s: %v\n", key, value)
		}
		if opts.InitGit {
			fmt.Println("\nWould initialize git repository")
		}
		return nil
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(opts.TargetPath, 0750); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Generate project
	if err := w.scaffolder.GenerateFromRitual(opts.TargetPath, opts.RitualPath, manifest, vars); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Save state
	state := &storage.State{
		RitualName:    manifest.Ritual.Name,
		RitualVersion: manifest.Ritual.Version,
		InstalledAt:   time.Now(),
	}

	if err := state.Save(opts.TargetPath); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Initialize git repository if requested
	if opts.InitGit {
		if err := initGitRepository(opts.TargetPath); err != nil {
			return fmt.Errorf("failed to initialize git repository: %w", err)
		}
		fmt.Println("✓ Initialized git repository")
	}

	fmt.Printf("✓ Project created successfully at: %s\n", opts.TargetPath)
	fmt.Printf("✓ Used ritual: %s v%s\n", manifest.Ritual.Name, manifest.Ritual.Version)

	return nil
}

// initGitRepository initializes a git repository in the target directory
func initGitRepository(targetPath string) error {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git command not found in PATH")
	}

	// Check if .git already exists
	gitDir := filepath.Join(targetPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return fmt.Errorf("git repository already exists")
	}

	// Initialize repository
	cmd := exec.Command("git", "init")
	cmd.Dir = targetPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %s: %w", string(output), err)
	}

	// Create initial commit if there are files
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = targetPath
	if err := cmd.Run(); err != nil {
		// Non-fatal - repository is still initialized
		return nil
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit from ritual")
	cmd.Dir = targetPath
	// Ignore error - commit might fail if git user.name/email not configured
	_ = cmd.Run()

	return nil
}

// Package cli provides command-line interface workflows
package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/toutaio/toutago-ritual-grove/internal/generator"
	"github.com/toutaio/toutago-ritual-grove/internal/questionnaire"
	"github.com/toutaio/toutago-ritual-grove/internal/storage"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

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

// Execute runs the create workflow
func Execute(ritualPath, targetPath string, answers map[string]interface{}, dryRun bool) error {
	workflow := NewCreateWorkflow()
	return workflow.Execute(ritualPath, targetPath, answers, dryRun)
}

// Execute runs the complete project creation workflow
func (w *CreateWorkflow) Execute(ritualPath, targetPath string, answers map[string]interface{}, dryRun bool) error {
	// Load ritual
	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	if err != nil {
		return fmt.Errorf("failed to load ritual: %w", err)
	}

	// If no answers provided, run questionnaire
	if answers == nil {
		if dryRun {
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
	if dryRun {
		fmt.Println("DRY RUN MODE - No files will be created")
		fmt.Printf("Would create project at: %s\n", targetPath)
		fmt.Printf("Using ritual: %s v%s\n", manifest.Ritual.Name, manifest.Ritual.Version)
		fmt.Println("\nAnswers:")
		for key, value := range answers {
			fmt.Printf("  %s: %v\n", key, value)
		}
		return nil
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetPath, 0750); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Generate project
	if err := w.scaffolder.GenerateFromRitual(targetPath, ritualPath, manifest, vars); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Save state
	state := &storage.State{
		RitualName:    manifest.Ritual.Name,
		RitualVersion: manifest.Ritual.Version,
		InstalledAt:   time.Now(),
	}

	if err := state.Save(targetPath); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Printf("✓ Project created successfully at: %s\n", targetPath)
	fmt.Printf("✓ Used ritual: %s v%s\n", manifest.Ritual.Name, manifest.Ritual.Version)

	return nil
}

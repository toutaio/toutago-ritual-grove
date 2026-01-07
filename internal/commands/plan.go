package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/toutaio/toutago-ritual-grove/internal/deployment"
	"github.com/toutaio/toutago-ritual-grove/internal/registry"
	"github.com/toutaio/toutago-ritual-grove/internal/storage"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// NewPlanCommand creates a command to show deployment plan for updating a ritual
func NewPlanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Show deployment plan for updating to a newer ritual version",
		Long: `Analyzes the differences between the current ritual version and the latest
available version, showing what will change during an update.

This command helps you preview changes before actually performing an update.`,
		RunE: runPlan,
	}

	cmd.Flags().String("to-version", "", "Target version to plan for (default: latest)")
	cmd.Flags().Bool("json", false, "Output plan in JSON format")

	return cmd
}

func runPlan(cmd *cobra.Command, args []string) error {
	targetVersion, _ := cmd.Flags().GetString("to-version")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Get current directory
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load current state
	currentState, err := storage.LoadState(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load current state: %w", err)
	}

	if currentState.RitualName == "" {
		return fmt.Errorf("no ritual found in current project")
	}

	fmt.Printf("Current ritual: %s v%s\n", currentState.RitualName, currentState.RitualVersion)

	// Load current ritual manifest
	loader := ritual.NewLoader(filepath.Join(projectDir, ".ritual"))
	currentManifest, err := loader.Load(filepath.Join(projectDir, ".ritual"))
	if err != nil {
		return fmt.Errorf("failed to load current ritual manifest: %w", err)
	}

	// Initialize registry to find target version
	reg := registry.NewRegistry()

	if err := reg.Scan(); err != nil {
		return fmt.Errorf("failed to scan rituals: %w", err)
	}

	// Load target ritual (for now, just use the same ritual name)
	targetManifest, err := reg.Load(currentState.RitualName)
	if err != nil {
		return fmt.Errorf("failed to load target ritual: %w", err)
	}

	if targetVersion != "" {
		// TODO: Support specific target versions when registry supports it
		fmt.Printf("Warning: Specific target versions not yet supported, using latest\n")
	}

	fmt.Printf("Target ritual: %s v%s\n\n", currentState.RitualName, targetManifest.Ritual.Version)

	// Generate deployment plan
	planner := deployment.NewPlanner()
	plan, err := planner.AnalyzeChanges(currentManifest, targetManifest)
	if err != nil {
		return fmt.Errorf("failed to analyze changes: %w", err)
	}

	// Output plan
	if jsonOutput {
		return outputPlanJSON(plan)
	} else {
		report := planner.GenerateReport(plan)
		fmt.Println(report)

		// Show warnings if there are conflicts
		if len(plan.Conflicts) > 0 {
			fmt.Println("⚠️  Warning: There are potential conflicts that require attention.")
			fmt.Println("   Review the conflicts above before proceeding with the update.")
		}
	}

	return nil
}

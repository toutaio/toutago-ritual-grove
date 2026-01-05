package deployment

import (
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// StepType represents the type of deployment step
type StepType string

const (
	StepBackup      StepType = "backup"
	StepMigration   StepType = "migration"
	StepUpdateFiles StepType = "update_files"
	StepRunHooks    StepType = "run_hooks"
	StepValidation  StepType = "validation"
	StepRollback    StepType = "rollback"
)

// DeploymentStep represents a single step in the deployment process
type DeploymentStep struct {
	Type          StepType
	Description   string
	EstimatedTime time.Duration
	Required      bool
}

// String returns a string representation of the step
func (s DeploymentStep) String() string {
	return fmt.Sprintf("%s: %s", s.Type, s.Description)
}

// Conflict represents a potential conflict in the deployment
type Conflict struct {
	File       string
	Reason     string
	Resolution string
}

// DeploymentPlan contains the analysis of what will change during deployment
type DeploymentPlan struct {
	CurrentVersion    string
	TargetVersion     string
	Steps             []DeploymentStep
	Conflicts         []Conflict
	FilesAdded        []string
	FilesModified     []string
	FilesDeleted      []string
	MigrationsToRun   []string
	EstimatedDuration time.Duration
}

// EstimateDuration calculates total estimated duration
func (p *DeploymentPlan) EstimateDuration() time.Duration {
	var total time.Duration
	for _, step := range p.Steps {
		if step.EstimatedTime > 0 {
			total += step.EstimatedTime
		} else {
			// Default estimates
			switch step.Type {
			case StepBackup:
				total += 5 * time.Second
			case StepMigration:
				total += 10 * time.Second
			case StepUpdateFiles:
				total += 2 * time.Second
			case StepRunHooks:
				total += 3 * time.Second
			case StepValidation:
				total += 2 * time.Second
			default:
				total += 1 * time.Second
			}
		}
	}
	p.EstimatedDuration = total
	return total
}

// detectConflicts identifies files that are both modified by user and will be updated
func (p *DeploymentPlan) detectConflicts(modifiedFiles, targetFiles []string) []Conflict {
	var conflicts []Conflict
	modifiedSet := make(map[string]bool)
	for _, f := range modifiedFiles {
		modifiedSet[f] = true
	}

	for _, f := range targetFiles {
		if modifiedSet[f] {
			conflicts = append(conflicts, Conflict{
				File:       f,
				Reason:     "File was manually modified and will be updated",
				Resolution: "Review changes and merge manually",
			})
		}
	}
	return conflicts
}

// Planner analyzes deployment changes and creates execution plans
type Planner struct{}

// NewPlanner creates a new deployment planner
func NewPlanner() *Planner {
	return &Planner{}
}

// AnalyzeChanges analyzes the differences between current and target ritual versions
func (p *Planner) AnalyzeChanges(current, target *ritual.Manifest) (*DeploymentPlan, error) {
	plan := &DeploymentPlan{
		CurrentVersion: current.Ritual.Version,
		TargetVersion:  target.Ritual.Version,
	}

	// Check if this is a breaking change
	currentVer, err := semver.NewVersion(current.Ritual.Version)
	if err == nil {
		targetVer, err := semver.NewVersion(target.Ritual.Version)
		if err == nil && targetVer.Major() > currentVer.Major() {
			plan.Conflicts = append(plan.Conflicts, Conflict{
				File: "version",
				Reason: fmt.Sprintf("Major version change from %s to %s indicates breaking changes",
					current.Ritual.Version, target.Ritual.Version),
				Resolution: "Review changelog and test thoroughly",
			})
		}
	}

	// Analyze file changes
	currentFiles := makeFileMap(current.Files.Templates)
	targetFiles := makeFileMap(target.Files.Templates)

	// Find added files
	for dest := range targetFiles {
		if _, exists := currentFiles[dest]; !exists {
			plan.FilesAdded = append(plan.FilesAdded, dest)
		}
	}

	// Find modified/deleted files
	for dest := range currentFiles {
		if _, exists := targetFiles[dest]; !exists {
			plan.FilesDeleted = append(plan.FilesDeleted, dest)
		} else {
			plan.FilesModified = append(plan.FilesModified, dest)
		}
	}

	// Build deployment steps
	plan.Steps = append(plan.Steps, DeploymentStep{
		Type:        StepBackup,
		Description: "Create backup of current project state",
		Required:    true,
	})

	if len(plan.FilesAdded) > 0 || len(plan.FilesModified) > 0 {
		desc := fmt.Sprintf("Update %d files, add %d new files",
			len(plan.FilesModified), len(plan.FilesAdded))
		plan.Steps = append(plan.Steps, DeploymentStep{
			Type:        StepUpdateFiles,
			Description: desc,
			Required:    true,
		})
	}

	// Add migration steps if migrations exist
	for _, migration := range target.Migrations {
		plan.MigrationsToRun = append(plan.MigrationsToRun, migration.ToVersion)
		plan.Steps = append(plan.Steps, DeploymentStep{
			Type:        StepMigration,
			Description: fmt.Sprintf("Run migration %s -> %s", migration.FromVersion, migration.ToVersion),
			Required:    true,
		})
	}

	// Add hook execution steps
	if len(target.Hooks.PostUpdate) > 0 {
		plan.Steps = append(plan.Steps, DeploymentStep{
			Type:        StepRunHooks,
			Description: "Execute post-update hooks",
			Required:    false,
		})
	}

	// Add validation step
	plan.Steps = append(plan.Steps, DeploymentStep{
		Type:        StepValidation,
		Description: "Validate deployment success",
		Required:    true,
	})

	// Calculate estimated duration
	plan.EstimateDuration()

	return plan, nil
}

// GenerateReport creates a human-readable deployment plan report
func (p *Planner) GenerateReport(plan *DeploymentPlan) string {
	var b strings.Builder

	b.WriteString("=== Deployment Plan ===\n\n")
	b.WriteString(fmt.Sprintf("Current Version: %s\n", plan.CurrentVersion))
	b.WriteString(fmt.Sprintf("Target Version:  %s\n", plan.TargetVersion))
	b.WriteString(fmt.Sprintf("Estimated Duration: %s\n\n", plan.EstimatedDuration))

	// File changes
	if len(plan.FilesAdded) > 0 {
		b.WriteString(fmt.Sprintf("Files to Add (%d):\n", len(plan.FilesAdded)))
		for _, f := range plan.FilesAdded {
			b.WriteString(fmt.Sprintf("  + %s\n", f))
		}
		b.WriteString("\n")
	}

	if len(plan.FilesModified) > 0 {
		b.WriteString(fmt.Sprintf("Files to Modify (%d):\n", len(plan.FilesModified)))
		for _, f := range plan.FilesModified {
			b.WriteString(fmt.Sprintf("  ~ %s\n", f))
		}
		b.WriteString("\n")
	}

	if len(plan.FilesDeleted) > 0 {
		b.WriteString(fmt.Sprintf("Files to Delete (%d):\n", len(plan.FilesDeleted)))
		for _, f := range plan.FilesDeleted {
			b.WriteString(fmt.Sprintf("  - %s\n", f))
		}
		b.WriteString("\n")
	}

	// Migration steps
	if len(plan.MigrationsToRun) > 0 {
		b.WriteString(fmt.Sprintf("Migrations to Run (%d):\n", len(plan.MigrationsToRun)))
		for _, m := range plan.MigrationsToRun {
			b.WriteString(fmt.Sprintf("  → %s\n", m))
		}
		b.WriteString("\n")
	}

	// Deployment steps
	b.WriteString("Deployment Steps:\n")
	for i, step := range plan.Steps {
		required := ""
		if step.Required {
			required = " [required]"
		}
		b.WriteString(fmt.Sprintf("  %d. %s%s\n", i+1, step.Description, required))
	}
	b.WriteString("\n")

	// Conflicts
	if len(plan.Conflicts) > 0 {
		b.WriteString(fmt.Sprintf("⚠ Potential Conflicts (%d):\n", len(plan.Conflicts)))
		for _, c := range plan.Conflicts {
			b.WriteString(fmt.Sprintf("  • %s: %s\n", c.File, c.Reason))
			if c.Resolution != "" {
				b.WriteString(fmt.Sprintf("    Resolution: %s\n", c.Resolution))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("===\n")
	return b.String()
}

// makeFileMap creates a map of destination paths to file mappings
func makeFileMap(files []ritual.FileMapping) map[string]ritual.FileMapping {
	m := make(map[string]ritual.FileMapping)
	for _, f := range files {
		m[f.Destination] = f
	}
	return m
}

package commands

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/toutaio/toutago-ritual-grove/internal/deployment"
	"github.com/toutaio/toutago-ritual-grove/internal/migration"
	"github.com/toutaio/toutago-ritual-grove/internal/storage"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// UpdateOptions contains options for the update command
type UpdateOptions struct {
	ToVersion string
	DryRun    bool
	Force     bool
}

// UpdateHandler handles ritual updates
type UpdateHandler struct{}

// NewUpdateHandler creates a new update command handler
func NewUpdateHandler() *UpdateHandler {
	return &UpdateHandler{}
}

// Execute updates a project to a new ritual version
func (h *UpdateHandler) Execute(projectPath string, opts UpdateOptions) error {
	// Validate project path
	if projectPath == "" {
		projectPath = "."
	}

	// Load current state
	state, err := storage.LoadState(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load project state: %w", err)
	}

	// Determine target version
	targetVersion := opts.ToVersion
	if targetVersion == "" {
		// TODO: Get latest version from registry
		return fmt.Errorf("target version not specified and auto-detection not yet implemented")
	}

	// Check if already at target version
	if state.RitualVersion == targetVersion {
		fmt.Printf("Project is already at version %s\n", targetVersion)
		return nil
	}

	// Parse versions for comparison
	currentVer, err := semver.NewVersion(state.RitualVersion)
	if err != nil {
		return fmt.Errorf("invalid current version: %w", err)
	}
	targetVer, err := semver.NewVersion(targetVersion)
	if err != nil {
		return fmt.Errorf("invalid target version: %w", err)
	}

	// Create update detector
	detector := deployment.NewUpdateDetector()
	updateInfo := detector.GetUpdateInfo(currentVer, targetVer)

	// Show update information
	fmt.Printf("Updating from %s to %s\n", state.RitualVersion, targetVersion)
	if updateInfo.IsBreaking {
		fmt.Println("⚠️  This is a BREAKING update")
	}

	// In dry-run mode, just show what would happen
	if opts.DryRun {
		fmt.Println("\n=== DRY RUN MODE ===")
		fmt.Println("No changes will be applied")
		return nil
	}

	// Create rollback manager for backup
	rollbackMgr := deployment.NewRollbackManager()
	
	// Create backup before updating
	backupPath, err := rollbackMgr.CreateBackup(projectPath)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	fmt.Printf("✓ Backup created at: %s\n", backupPath)

	// Load new ritual manifest
	// TODO: Get ritual path from registry
	ritualPath := state.RitualName
	loader := ritual.NewLoader(ritualPath)
	newManifest, err := loader.Load(ritualPath)
	if err != nil {
		return fmt.Errorf("failed to load new ritual: %w", err)
	}

	// Run migrations
	migrationRunner := migration.NewRunner(projectPath)
	migrations := h.getMigrationsToRun(state, newManifest)
	
	fmt.Printf("\nRunning %d migration(s)...\n", len(migrations))
	for _, mig := range migrations {
		fmt.Printf("  - %s\n", mig.ToVersion)
		if err := migrationRunner.RunUp(mig); err != nil {
			// Rollback on error if not forced
			if !opts.Force {
				fmt.Println("\n⚠️  Migration failed, rolling back...")
				if rbErr := rollbackMgr.RestoreFromBackup(backupPath, projectPath); rbErr != nil {
					return fmt.Errorf("migration failed and rollback failed: %w (rollback error: %v)", err, rbErr)
				}
				return fmt.Errorf("migration failed, changes rolled back: %w", err)
			}
			return fmt.Errorf("migration failed: %w", err)
		}
		// Track migration in state
		state.AddMigration(mig.ToVersion)
	}

	// Update state version
	state.RitualVersion = targetVersion
	
	if err := state.Save(projectPath); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Println("\n✓ Update completed successfully")
	return nil
}

// getMigrationsToRun determines which migrations need to be run
func (h *UpdateHandler) getMigrationsToRun(state *storage.State, manifest *ritual.Manifest) []*ritual.Migration {
	var migrationsToRun []*ritual.Migration

	// Get all migrations from manifest
	for i := range manifest.Migrations {
		mig := &manifest.Migrations[i]
		
		// Skip if already applied
		if state.IsMigrationApplied(mig.ToVersion) {
			continue
		}

		migrationsToRun = append(migrationsToRun, mig)
	}

	return migrationsToRun
}

// CanUpdate checks if an update is available
func (h *UpdateHandler) CanUpdate(projectPath string) (bool, string, error) {
	state, err := storage.LoadState(projectPath)
	if err != nil {
		return false, "", err
	}

	// TODO: Check registry for newer version
	// For now, just return the current version
	return false, state.RitualVersion, nil
}

// ShowUpdateInfo displays information about an available update
func (h *UpdateHandler) ShowUpdateInfo(projectPath string, targetVersion string) error {
	state, err := storage.LoadState(projectPath)
	if err != nil {
		return err
	}

	currentVer, err := semver.NewVersion(state.RitualVersion)
	if err != nil {
		return fmt.Errorf("invalid current version: %w", err)
	}
	targetVer, err := semver.NewVersion(targetVersion)
	if err != nil {
		return fmt.Errorf("invalid target version: %w", err)
	}

	detector := deployment.NewUpdateDetector()
	info := detector.GetUpdateInfo(currentVer, targetVer)

	fmt.Printf("Current version: %s\n", state.RitualVersion)
	fmt.Printf("Target version:  %s\n", targetVersion)
	fmt.Printf("Type:            %s\n", info.UpdateType)
	if info.IsBreaking {
		fmt.Println("Breaking:        YES ⚠️")
	} else {
		fmt.Println("Breaking:        No")
	}

	return nil
}

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
	if projectPath == "" {
		projectPath = "."
	}

	state, err := storage.LoadState(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load project state: %w", err)
	}

	targetVersion, err := h.determineTargetVersion(opts.ToVersion)
	if err != nil {
		return err
	}

	if state.RitualVersion == targetVersion {
		fmt.Printf("Project is already at version %s\n", targetVersion)
		return nil
	}

	currentVer, targetVer, err := h.parseVersions(state.RitualVersion, targetVersion)
	if err != nil {
		return err
	}

	h.displayUpdateInfo(state.RitualVersion, targetVersion, currentVer, targetVer)

	if opts.DryRun {
		return h.handleDryRun()
	}

	backupPath, err := h.createBackup(projectPath)
	if err != nil {
		return err
	}

	newManifest, err := h.loadNewRitual(state.RitualName)
	if err != nil {
		return err
	}

	if err := h.runMigrations(projectPath, state, newManifest, backupPath, opts.Force); err != nil {
		return err
	}

	return h.saveUpdatedState(state, targetVersion, projectPath)
}

func (h *UpdateHandler) determineTargetVersion(toVersion string) (string, error) {
	if toVersion == "" {
		return "", fmt.Errorf("target version not specified and auto-detection not yet implemented")
	}
	return toVersion, nil
}

func (h *UpdateHandler) parseVersions(current, target string) (*semver.Version, *semver.Version, error) {
	currentVer, err := semver.NewVersion(current)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid current version: %w", err)
	}
	targetVer, err := semver.NewVersion(target)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid target version: %w", err)
	}
	return currentVer, targetVer, nil
}

func (h *UpdateHandler) displayUpdateInfo(current, target string, currentVer, targetVer *semver.Version) {
	detector := deployment.NewUpdateDetector()
	updateInfo := detector.GetUpdateInfo(currentVer, targetVer)

	fmt.Printf("Updating from %s to %s\n", current, target)
	if updateInfo.IsBreaking {
		fmt.Println("⚠️  This is a BREAKING update")
	}
}

func (h *UpdateHandler) handleDryRun() error {
	fmt.Println("\n=== DRY RUN MODE ===")
	fmt.Println("No changes will be applied")
	return nil
}

func (h *UpdateHandler) createBackup(projectPath string) (string, error) {
	rollbackMgr := deployment.NewRollbackManager()
	backupPath, err := rollbackMgr.CreateBackup(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}
	fmt.Printf("✓ Backup created at: %s\n", backupPath)
	return backupPath, nil
}

func (h *UpdateHandler) loadNewRitual(ritualName string) (*ritual.Manifest, error) {
	loader := ritual.NewLoader(ritualName)
	newManifest, err := loader.Load(ritualName)
	if err != nil {
		return nil, fmt.Errorf("failed to load new ritual: %w", err)
	}
	return newManifest, nil
}

func (h *UpdateHandler) runMigrations(
	projectPath string,
	state *storage.State,
	manifest *ritual.Manifest,
	backupPath string,
	force bool,
) error {
	migrationRunner := migration.NewRunner(projectPath)
	migrations := h.getMigrationsToRun(state, manifest)

	fmt.Printf("\nRunning %d migration(s)...\n", len(migrations))
	for _, mig := range migrations {
		fmt.Printf("  - %s\n", mig.ToVersion)
		if err := migrationRunner.RunUp(mig); err != nil {
			return h.handleMigrationError(err, backupPath, projectPath, force)
		}
		state.AddMigration(mig.ToVersion)
	}
	return nil
}

func (h *UpdateHandler) handleMigrationError(err error, backupPath, projectPath string, force bool) error {
	if !force {
		fmt.Println("\n⚠️  Migration failed, rolling back...")
		rollbackMgr := deployment.NewRollbackManager()
		if rbErr := rollbackMgr.RestoreFromBackup(backupPath, projectPath); rbErr != nil {
			return fmt.Errorf("migration failed and rollback failed: %w (rollback error: %v)", err, rbErr)
		}
		return fmt.Errorf("migration failed, changes rolled back: %w", err)
	}
	return fmt.Errorf("migration failed: %w", err)
}

func (h *UpdateHandler) saveUpdatedState(state *storage.State, targetVersion, projectPath string) error {
	state.RitualVersion = targetVersion
	if err := state.Save(projectPath); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	fmt.Println("\n✓ Update completed successfully")
	return nil
}

// getMigrationsToRun determines which migrations need to be run
func (h *UpdateHandler) getMigrationsToRun(state *storage.State, manifest *ritual.Manifest) []*ritual.Migration {
	migrationsToRun := make([]*ritual.Migration, 0, len(manifest.Migrations))

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

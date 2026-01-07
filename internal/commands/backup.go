package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/toutaio/toutago-ritual-grove/internal/deployment"
	"github.com/toutaio/toutago-ritual-grove/internal/storage"
)

// NewBackupCommand creates the backup command
func NewBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage project backups",
		Long: `Create and manage backups of your ritual-based project.

Backups include project files, configuration, and ritual state.
They are stored in .ritual/backups/ directory.`,
	}

	cmd.AddCommand(newBackupListCommand())
	cmd.AddCommand(newBackupCreateCommand())
	cmd.AddCommand(newBackupRestoreCommand())
	cmd.AddCommand(newBackupCleanCommand())

	return cmd
}

func newBackupListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available backups",
		RunE:  runBackupList,
	}
}

func newBackupCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new backup",
		RunE:  runBackupCreate,
	}

	cmd.Flags().String("description", "", "Backup description")

	return cmd
}

func newBackupRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore <backup-name>",
		Short: "Restore from a backup",
		Args:  cobra.ExactArgs(1),
		RunE:  runBackupRestore,
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

func newBackupCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean old backups",
		RunE:  runBackupClean,
	}

	cmd.Flags().Int("keep", 5, "Number of backups to keep")

	return cmd
}

func runBackupList(cmd *cobra.Command, args []string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	rm := deployment.NewRollbackManager()
	backups, err := rm.ListBackups(projectDir)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backups) == 0 {
		fmt.Println("No backups found")
		return nil
	}

	fmt.Printf("Available backups (%d):\n\n", len(backups))

	for i, backup := range backups {
		size, _ := rm.GetBackupSize(backup.Path)
		sizeStr := formatSize(size)

		fmt.Printf("%d. %s\n", i+1, backup.Path)
		fmt.Printf("   Ritual: %s v%s\n", backup.RitualName, backup.RitualVersion)
		fmt.Printf("   Created: %s\n", backup.CreatedAt.Format(time.RFC3339))
		fmt.Printf("   Size: %s\n", sizeStr)
		if backup.Description != "" {
			fmt.Printf("   Description: %s\n", backup.Description)
		}
		fmt.Println()
	}

	return nil
}

func runBackupCreate(cmd *cobra.Command, args []string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	description, _ := cmd.Flags().GetString("description")

	// Load current state
	state, err := storage.LoadState(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Create backup with metadata
	rm := deployment.NewRollbackManager()
	metadata := deployment.BackupMetadata{
		RitualName:    state.RitualName,
		RitualVersion: state.RitualVersion,
		Description:   description,
		CreatedAt:     time.Now(),
	}

	fmt.Println("Creating backup...")
	backupPath, err := rm.CreateBackupWithMetadata(projectDir, metadata)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	size, _ := rm.GetBackupSize(backupPath)
	fmt.Printf("✓ Backup created: %s\n", backupPath)
	fmt.Printf("  Size: %s\n", formatSize(size))

	return nil
}

func runBackupRestore(cmd *cobra.Command, args []string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	backupName := args[0]
	force, _ := cmd.Flags().GetBool("force")

	// Find backup by name or index
	rm := deployment.NewRollbackManager()
	backups, err := rm.ListBackups(projectDir)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	var backupPath string
	for _, backup := range backups {
		if backup.Path == backupName {
			backupPath = backup.Path
			break
		}
	}

	if backupPath == "" {
		return fmt.Errorf("backup not found: %s", backupName)
	}

	// Confirm restoration
	if !force {
		fmt.Printf("⚠️  This will restore your project to a previous state.\n")
		fmt.Printf("   Current files will be overwritten.\n")
		fmt.Printf("   Backup to restore: %s\n\n", backupPath)
		fmt.Print("Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Restore cancelled")
			return nil
		}
	}

	fmt.Println("Restoring from backup...")
	if err := rm.RestoreFromBackup(backupPath, projectDir); err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}

	fmt.Println("✓ Restore completed successfully")

	return nil
}

func runBackupClean(cmd *cobra.Command, args []string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	keep, _ := cmd.Flags().GetInt("keep")

	rm := deployment.NewRollbackManager()

	// List backups before cleaning
	backups, err := rm.ListBackups(projectDir)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	toDelete := len(backups) - keep
	if toDelete <= 0 {
		fmt.Printf("No backups to clean (have %d, keeping %d)\n", len(backups), keep)
		return nil
	}

	fmt.Printf("Cleaning %d old backup(s), keeping %d newest...\n", toDelete, keep)

	if err := rm.CleanOldBackups(projectDir, keep); err != nil {
		return fmt.Errorf("failed to clean backups: %w", err)
	}

	fmt.Printf("✓ Cleaned %d backup(s)\n", toDelete)

	return nil
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

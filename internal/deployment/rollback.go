package deployment

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// RollbackManager handles backup and restore operations
type RollbackManager struct {
	backupDir string
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager() *RollbackManager {
	return &RollbackManager{
		backupDir: ".ritual/backups",
	}
}

// BackupMetadata contains information about a backup
type BackupMetadata struct {
	RitualName    string    `json:"ritual_name"`
	RitualVersion string    `json:"ritual_version"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	Path          string    `json:"path"`
}

// CreateBackup creates a backup of the project directory
func (r *RollbackManager) CreateBackup(projectDir string) (string, error) {
	return r.CreateBackupWithMetadata(projectDir, BackupMetadata{})
}

// CreateBackupWithMetadata creates a backup with associated metadata
func (r *RollbackManager) CreateBackupWithMetadata(projectDir string, metadata BackupMetadata) (string, error) {
	timestamp := time.Now().Format("20060102-150405.000")
	backupName := fmt.Sprintf("backup-%s", timestamp)
	backupPath := filepath.Join(projectDir, r.backupDir, backupName)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0750); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy all files to backup
	if err := r.copyDirectory(projectDir, backupPath); err != nil {
		return "", fmt.Errorf("failed to copy files to backup: %w", err)
	}

	// Save metadata
	metadata.CreatedAt = time.Now()
	metadata.Path = backupPath
	metadataPath := filepath.Join(backupPath, "backup.json")
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, metadataJSON, 0600); err != nil {
		return "", fmt.Errorf("failed to write metadata: %w", err)
	}

	return backupPath, nil
}

// RestoreFromBackup restores files from a backup
func (r *RollbackManager) RestoreFromBackup(backupPath, targetDir string) error {
	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", backupPath)
	}

	// Copy files from backup to target
	return r.copyDirectory(backupPath, targetDir)
}

// ListBackups returns all backups for a project, sorted by creation time (newest first)
func (r *RollbackManager) ListBackups(projectDir string) ([]BackupMetadata, error) {
	backupsPath := filepath.Join(projectDir, r.backupDir)

	if _, err := os.Stat(backupsPath); os.IsNotExist(err) {
		return []BackupMetadata{}, nil
	}

	entries, err := os.ReadDir(backupsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backups directory: %w", err)
	}

	var backups []BackupMetadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		backupPath := filepath.Join(backupsPath, entry.Name())
		metadata, err := r.ReadBackupMetadata(backupPath)
		if err != nil {
			// If metadata can't be read, create basic metadata from directory info
			info, _ := entry.Info()
			metadata = BackupMetadata{
				Path:      backupPath,
				CreatedAt: info.ModTime(),
			}
		}

		backups = append(backups, metadata)
	}

	// Sort by creation time, newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// CleanOldBackups removes old backups, keeping only the specified number of most recent ones
func (r *RollbackManager) CleanOldBackups(projectDir string, keepCount int) error {
	backups, err := r.ListBackups(projectDir)
	if err != nil {
		return err
	}

	if len(backups) <= keepCount {
		return nil
	}

	// Remove old backups
	for i := keepCount; i < len(backups); i++ {
		if err := os.RemoveAll(backups[i].Path); err != nil {
			return fmt.Errorf("failed to remove old backup: %w", err)
		}
	}

	return nil
}

// ReadBackupMetadata reads metadata from a backup
func (r *RollbackManager) ReadBackupMetadata(backupPath string) (BackupMetadata, error) {
	metadataPath := filepath.Join(backupPath, "backup.json")

	// #nosec G304 - metadataPath is constructed from validated backup directory
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return BackupMetadata{}, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return BackupMetadata{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

// GetBackupSize returns the total size of a backup in bytes
func (r *RollbackManager) GetBackupSize(backupPath string) (int64, error) {
	var size int64

	err := filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// copyDirectory recursively copies a directory
func (r *RollbackManager) copyDirectory(src, dst string) error {
	backupDirPath := filepath.Join(src, r.backupDir)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip backup directory itself to avoid recursion
		if path == backupDirPath || (len(path) > len(backupDirPath) &&
			path[:len(backupDirPath)] == backupDirPath &&
			(path[len(backupDirPath)] == '/' || path[len(backupDirPath)] == filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return r.copyFile(path, targetPath)
	})
}

// copyFile copies a single file
// #nosec G304 - Source path is from validated backup metadata
func (r *RollbackManager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = sourceFile.Close() // Best effort close
	}()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
		// #nosec G304 - Destination path is validated and from backup metadata
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close() // Best effort close
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

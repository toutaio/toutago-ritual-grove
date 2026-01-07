package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/toutaio/toutago-ritual-grove/internal/deployment"
	"github.com/toutaio/toutago-ritual-grove/internal/storage"
)

// TestBackupCommand_List tests listing available backups
func TestBackupCommand_List(t *testing.T) {
	tmpDir := t.TempDir()

	// Create state
	state := &storage.State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
		InstalledAt:   time.Now(),
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create backup manager
	rm := deployment.NewRollbackManager()

	// Create test backups
	metadata1 := deployment.BackupMetadata{
		RitualName:    "test",
		RitualVersion: "1.0.0",
		Description:   "Before update to 1.1.0",
		CreatedAt:     time.Now().Add(-2 * time.Hour),
	}
	backup1, err := rm.CreateBackupWithMetadata(tmpDir, metadata1)
	if err != nil {
		t.Fatalf("Failed to create backup 1: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	metadata2 := deployment.BackupMetadata{
		RitualName:    "test",
		RitualVersion: "1.1.0",
		Description:   "Before update to 1.2.0",
		CreatedAt:     time.Now(),
	}
	backup2, err := rm.CreateBackupWithMetadata(tmpDir, metadata2)
	if err != nil {
		t.Fatalf("Failed to create backup 2: %v", err)
	}

	// List backups
	backups, err := rm.ListBackups(tmpDir)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) != 2 {
		t.Errorf("Expected 2 backups, got %d", len(backups))
	}

	// Verify backups exist
	if _, err := os.Stat(backup1); os.IsNotExist(err) {
		t.Error("Backup 1 does not exist")
	}
	if _, err := os.Stat(backup2); os.IsNotExist(err) {
		t.Error("Backup 2 does not exist")
	}
}

// TestBackupCommand_Create tests manual backup creation
func TestBackupCommand_Create(t *testing.T) {
	tmpDir := t.TempDir()

	// Create state
	state := &storage.State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create backup
	rm := deployment.NewRollbackManager()
	backupPath, err := rm.CreateBackup(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup does not exist")
	}

	// Verify backup is in correct location
	expectedDir := filepath.Join(tmpDir, ".ritual", "backups")
	if !filepath.HasPrefix(backupPath, expectedDir) {
		t.Errorf("Backup not in expected directory: %s", backupPath)
	}
}

// TestBackupCommand_Restore tests restoring from backup
func TestBackupCommand_Restore(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("original"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create state
	state := &storage.State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create backup
	rm := deployment.NewRollbackManager()
	backupPath, err := rm.CreateBackup(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Modify file
	if err := os.WriteFile(testFile, []byte("modified"), 0600); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Restore from backup
	if err := rm.RestoreFromBackup(backupPath, tmpDir); err != nil {
		t.Fatalf("Failed to restore: %v", err)
	}

	// Verify file restored
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(content) != "original" {
		t.Errorf("Expected 'original', got '%s'", string(content))
	}
}

// TestBackupCommand_Clean tests cleaning old backups
func TestBackupCommand_Clean(t *testing.T) {
	tmpDir := t.TempDir()

	// Create state
	state := &storage.State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create multiple backups
	rm := deployment.NewRollbackManager()
	for i := 0; i < 5; i++ {
		_, err := rm.CreateBackup(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create backup %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Clean old backups (keep 3)
	if err := rm.CleanOldBackups(tmpDir, 3); err != nil {
		t.Fatalf("Failed to clean old backups: %v", err)
	}

	// Verify only 3 backups remain
	backups, err := rm.ListBackups(tmpDir)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) != 3 {
		t.Errorf("Expected 3 backups after cleanup, got %d", len(backups))
	}
}

// TestBackupCommand_Metadata tests backup metadata
func TestBackupCommand_Metadata(t *testing.T) {
	tmpDir := t.TempDir()

	// Create state
	state := &storage.State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create backup with metadata
	rm := deployment.NewRollbackManager()
	metadata := deployment.BackupMetadata{
		RitualName:    "test",
		RitualVersion: "1.0.0",
		Description:   "Manual backup before changes",
		CreatedAt:     time.Now(),
	}

	backupPath, err := rm.CreateBackupWithMetadata(tmpDir, metadata)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Read metadata
	readMetadata, err := rm.ReadBackupMetadata(backupPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	if readMetadata.RitualName != "test" {
		t.Errorf("Expected RitualName 'test', got %s", readMetadata.RitualName)
	}

	if readMetadata.RitualVersion != "1.0.0" {
		t.Errorf("Expected RitualVersion 1.0.0, got %s", readMetadata.RitualVersion)
	}

	if readMetadata.Description != "Manual backup before changes" {
		t.Errorf("Expected Description 'Manual backup before changes', got %s", readMetadata.Description)
	}
}

// TestBackupCommand_Size tests getting backup size
func TestBackupCommand_Size(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "large.txt")
	content := make([]byte, 1024*10) // 10KB
	if err := os.WriteFile(testFile, content, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create state
	state := &storage.State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create backup
	rm := deployment.NewRollbackManager()
	backupPath, err := rm.CreateBackup(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Get backup size
	size, err := rm.GetBackupSize(backupPath)
	if err != nil {
		t.Fatalf("Failed to get backup size: %v", err)
	}

	if size == 0 {
		t.Error("Backup size should be greater than 0")
	}

	// Size should be at least the file size
	if size < 1024*10 {
		t.Errorf("Expected backup size >= 10KB, got %d bytes", size)
	}
}

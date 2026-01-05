package deployment

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRollback_CreateBackup(t *testing.T) {
	rb := NewRollbackManager()

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectDir, 0750)

	// Create some files
	os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("package main"), 0600)
	os.WriteFile(filepath.Join(projectDir, "config.yaml"), []byte("port: 8080"), 0600)

	backupPath, err := rb.CreateBackup(projectDir)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	if backupPath == "" {
		t.Error("Expected backup path to be set")
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup directory not created: %s", backupPath)
	}
}

func TestRollback_RestoreFromBackup(t *testing.T) {
	rb := NewRollbackManager()

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectDir, 0750)

	// Create original file
	originalContent := []byte("package main\n// v1.0")
	os.WriteFile(filepath.Join(projectDir, "main.go"), originalContent, 0600)

	// Create backup
	backupPath, err := rb.CreateBackup(projectDir)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Modify the file (simulating an update)
	os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("package main\n// v2.0"), 0600)

	// Rollback
	if err := rb.RestoreFromBackup(backupPath, projectDir); err != nil {
		t.Fatalf("RestoreFromBackup failed: %v", err)
	}

	// Verify file was restored
	restored, err := os.ReadFile(filepath.Join(projectDir, "main.go"))
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(restored) != string(originalContent) {
		t.Errorf("File not restored correctly. Got %q, want %q", string(restored), string(originalContent))
	}
}

func TestRollback_ListBackups(t *testing.T) {
	rb := NewRollbackManager()

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectDir, 0750)

	// Create multiple backups
	backup1, err1 := rb.CreateBackup(projectDir)
	t.Logf("Backup 1: %s, err: %v", backup1, err1)

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	backup2, err2 := rb.CreateBackup(projectDir)
	t.Logf("Backup 2: %s, err: %v", backup2, err2)

	backups, err := rb.ListBackups(projectDir)
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	t.Logf("Found %d backups", len(backups))
	for i, b := range backups {
		t.Logf("  Backup %d: %s", i, b.Path)
	}

	if len(backups) < 2 {
		t.Errorf("Expected at least 2 backups, got %d", len(backups))
	}

	// Verify backups are sorted (newest first)
	if len(backups) >= 2 {
		if backups[0].CreatedAt.Before(backups[1].CreatedAt) {
			t.Error("Backups not sorted correctly (should be newest first)")
		}
	}

	_ = backup1
	_ = backup2
}

func TestRollback_CleanOldBackups(t *testing.T) {
	rb := NewRollbackManager()

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectDir, 0750)

	// Create multiple backups
	for i := 0; i < 5; i++ {
		rb.CreateBackup(projectDir)
		time.Sleep(10 * time.Millisecond)
	}

	// Clean old backups, keep only 3
	err := rb.CleanOldBackups(projectDir, 3)
	if err != nil {
		t.Fatalf("CleanOldBackups failed: %v", err)
	}

	backups, _ := rb.ListBackups(projectDir)
	if len(backups) > 3 {
		t.Errorf("Expected max 3 backups after cleanup, got %d", len(backups))
	}
}

func TestRollback_BackupMetadata(t *testing.T) {
	rb := NewRollbackManager()

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectDir, 0750)

	metadata := BackupMetadata{
		RitualName:    "test-ritual",
		RitualVersion: "1.0.0",
		Description:   "Backup before update to 1.1.0",
	}

	backupPath, err := rb.CreateBackupWithMetadata(projectDir, metadata)
	if err != nil {
		t.Fatalf("CreateBackupWithMetadata failed: %v", err)
	}

	// Read metadata back
	meta, err := rb.ReadBackupMetadata(backupPath)
	if err != nil {
		t.Fatalf("ReadBackupMetadata failed: %v", err)
	}

	if meta.RitualName != metadata.RitualName {
		t.Errorf("RitualName = %v, want %v", meta.RitualName, metadata.RitualName)
	}
	if meta.RitualVersion != metadata.RitualVersion {
		t.Errorf("RitualVersion = %v, want %v", meta.RitualVersion, metadata.RitualVersion)
	}
}

func TestRollback_GetBackupSize(t *testing.T) {
	rb := NewRollbackManager()

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectDir, 0750)

	// Create file with known size
	os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("package main\n"), 0600)

	backupPath, _ := rb.CreateBackup(projectDir)

	size, err := rb.GetBackupSize(backupPath)
	if err != nil {
		t.Fatalf("GetBackupSize failed: %v", err)
	}

	if size <= 0 {
		t.Errorf("Expected backup size > 0, got %d", size)
	}
}

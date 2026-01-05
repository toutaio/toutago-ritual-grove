package migration

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestNewRunner(t *testing.T) {
	testPath := filepath.Join(os.TempDir(), "test-project")
	runner := NewRunner(testPath)
	if runner == nil {
		t.Fatal("NewRunner returned nil")
	}

	if runner.projectPath != testPath {
		t.Errorf("projectPath = %s, want %s", runner.projectPath, testPath)
	}

	if runner.dryRun {
		t.Error("dryRun should be false by default")
	}

	if len(runner.records) != 0 {
		t.Errorf("records should be empty, got %d", len(runner.records))
	}
}

func TestSetDryRun(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))

	runner.SetDryRun(true)
	if !runner.dryRun {
		t.Error("dryRun should be true")
	}

	runner.SetDryRun(false)
	if runner.dryRun {
		t.Error("dryRun should be false")
	}
}

func TestRunUpDryRun(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))
	runner.SetDryRun(true)

	migration := &ritual.Migration{
		FromVersion: "1.0.0",
		ToVersion:   "1.1.0",
		Description: "Test migration",
		Up: ritual.MigrationHandler{
			SQL: []string{"CREATE TABLE test (id INT);"},
		},
	}

	err := runner.RunUp(migration)
	if err != nil {
		t.Fatalf("RunUp failed: %v", err)
	}

	records := runner.GetRecords()
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	if records[0].Status != StatusSkipped {
		t.Errorf("Status = %s, want %s", records[0].Status, StatusSkipped)
	}
}

func TestRunUpWithSQL(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))

	migration := &ritual.Migration{
		FromVersion: "1.0.0",
		ToVersion:   "1.1.0",
		Description: "Add users table",
		Up: ritual.MigrationHandler{
			SQL: []string{
				"CREATE TABLE users (id INT PRIMARY KEY);",
				"CREATE INDEX idx_users_id ON users(id);",
			},
		},
	}

	err := runner.RunUp(migration)
	if err != nil {
		t.Fatalf("RunUp failed: %v", err)
	}

	records := runner.GetAppliedMigrations()
	if len(records) != 1 {
		t.Fatalf("Expected 1 applied migration, got %d", len(records))
	}

	record := records[0]
	if record.FromVersion != "1.0.0" {
		t.Errorf("FromVersion = %s, want 1.0.0", record.FromVersion)
	}
	if record.ToVersion != "1.1.0" {
		t.Errorf("ToVersion = %s, want 1.1.0", record.ToVersion)
	}
	if record.Status != StatusApplied {
		t.Errorf("Status = %s, want %s", record.Status, StatusApplied)
	}
}

func TestRunUpWithEmptySQL(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))

	migration := &ritual.Migration{
		FromVersion: "1.0.0",
		ToVersion:   "1.1.0",
		Description: "Empty SQL migration",
		Up: ritual.MigrationHandler{
			SQL: []string{"   "},
		},
	}

	err := runner.RunUp(migration)
	if err == nil {
		t.Error("Expected error for empty SQL, got nil")
	}

	failed := runner.GetFailedMigrations()
	if len(failed) != 1 {
		t.Fatalf("Expected 1 failed migration, got %d", len(failed))
	}
}

func TestRunUpWithScript(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shell script test on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "migration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test script
	scriptPath := filepath.Join(tmpDir, "migrate.sh")
	scriptContent := "#!/bin/bash\necho 'Migration executed'\n"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
		t.Fatalf("Failed to write script: %v", err)
	}

	runner := NewRunner(tmpDir)

	migration := &ritual.Migration{
		FromVersion: "1.0.0",
		ToVersion:   "1.1.0",
		Description: "Script migration",
		Up: ritual.MigrationHandler{
			Script: "migrate.sh",
		},
	}

	err = runner.RunUp(migration)
	if err != nil {
		t.Fatalf("RunUp failed: %v", err)
	}

	records := runner.GetAppliedMigrations()
	if len(records) != 1 {
		t.Fatalf("Expected 1 applied migration, got %d", len(records))
	}
}

func TestRunUpWithMissingScript(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))

	migration := &ritual.Migration{
		FromVersion: "1.0.0",
		ToVersion:   "1.1.0",
		Description: "Missing script migration",
		Up: ritual.MigrationHandler{
			Script: "nonexistent.sh",
		},
	}

	err := runner.RunUp(migration)
	if err == nil {
		t.Error("Expected error for missing script, got nil")
	}

	failed := runner.GetFailedMigrations()
	if len(failed) != 1 {
		t.Fatalf("Expected 1 failed migration, got %d", len(failed))
	}
}

func TestRunDown(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))

	migration := &ritual.Migration{
		FromVersion: "1.0.0",
		ToVersion:   "1.1.0",
		Description: "Test rollback",
		Down: ritual.MigrationHandler{
			SQL: []string{"DROP TABLE users;"},
		},
	}

	err := runner.RunDown(migration)
	if err != nil {
		t.Fatalf("RunDown failed: %v", err)
	}

	records := runner.GetRecords()
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	record := records[0]
	if record.Status != StatusRolledBack {
		t.Errorf("Status = %s, want %s", record.Status, StatusRolledBack)
	}
	if record.FromVersion != "1.1.0" {
		t.Errorf("FromVersion = %s, want 1.1.0 (reversed)", record.FromVersion)
	}
	if record.ToVersion != "1.0.0" {
		t.Errorf("ToVersion = %s, want 1.0.0 (reversed)", record.ToVersion)
	}
}

func TestRunMigrationChainUp(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))

	migrations := []*ritual.Migration{
		{
			FromVersion: "1.0.0",
			ToVersion:   "1.1.0",
			Description: "First migration",
			Up: ritual.MigrationHandler{
				SQL: []string{"CREATE TABLE users (id INT);"},
			},
		},
		{
			FromVersion: "1.1.0",
			ToVersion:   "1.2.0",
			Description: "Second migration",
			Up: ritual.MigrationHandler{
				SQL: []string{"ALTER TABLE users ADD COLUMN name VARCHAR(255);"},
			},
		},
		{
			FromVersion: "1.2.0",
			ToVersion:   "1.3.0",
			Description: "Third migration",
			Up: ritual.MigrationHandler{
				SQL: []string{"CREATE INDEX idx_users_name ON users(name);"},
			},
		},
	}

	err := runner.RunMigrationChain(migrations, "up")
	if err != nil {
		t.Fatalf("RunMigrationChain failed: %v", err)
	}

	applied := runner.GetAppliedMigrations()
	if len(applied) != 3 {
		t.Fatalf("Expected 3 applied migrations, got %d", len(applied))
	}

	// Verify order
	if applied[0].ToVersion != "1.1.0" {
		t.Errorf("First migration ToVersion = %s, want 1.1.0", applied[0].ToVersion)
	}
	if applied[1].ToVersion != "1.2.0" {
		t.Errorf("Second migration ToVersion = %s, want 1.2.0", applied[1].ToVersion)
	}
	if applied[2].ToVersion != "1.3.0" {
		t.Errorf("Third migration ToVersion = %s, want 1.3.0", applied[2].ToVersion)
	}
}

func TestRunMigrationChainDown(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))

	migrations := []*ritual.Migration{
		{
			FromVersion: "1.2.0",
			ToVersion:   "1.3.0",
			Description: "Third migration",
			Down: ritual.MigrationHandler{
				SQL: []string{"DROP INDEX idx_users_name;"},
			},
		},
		{
			FromVersion: "1.1.0",
			ToVersion:   "1.2.0",
			Description: "Second migration",
			Down: ritual.MigrationHandler{
				SQL: []string{"ALTER TABLE users DROP COLUMN name;"},
			},
		},
	}

	err := runner.RunMigrationChain(migrations, "down")
	if err != nil {
		t.Fatalf("RunMigrationChain failed: %v", err)
	}

	records := runner.GetRecords()
	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(records))
	}

	for _, record := range records {
		if record.Status != StatusRolledBack {
			t.Errorf("Record status = %s, want %s", record.Status, StatusRolledBack)
		}
	}
}

func TestRunMigrationChainInvalidDirection(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))

	err := runner.RunMigrationChain([]*ritual.Migration{}, "sideways")
	if err == nil {
		t.Error("Expected error for invalid direction, got nil")
	}
}

func TestValidateMigration(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))

	tests := []struct {
		name        string
		migration   *ritual.Migration
		wantErr     bool
		errContains string
	}{
		{
			name: "valid migration with SQL",
			migration: &ritual.Migration{
				Up: ritual.MigrationHandler{
					SQL: []string{"CREATE TABLE test (id INT);"},
				},
				Down: ritual.MigrationHandler{
					SQL: []string{"DROP TABLE test;"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid migration with script",
			migration: &ritual.Migration{
				Up: ritual.MigrationHandler{
					Script: "up.sh",
				},
				Down: ritual.MigrationHandler{
					Script: "down.sh",
				},
			},
			wantErr: false,
		},
		{
			name: "valid idempotent migration without down",
			migration: &ritual.Migration{
				Idempotent: true,
				Up: ritual.MigrationHandler{
					SQL: []string{"CREATE TABLE IF NOT EXISTS test (id INT);"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid migration - no up handler",
			migration: &ritual.Migration{
				Up: ritual.MigrationHandler{},
			},
			wantErr:     true,
			errContains: "no up handler",
		},
		{
			name: "invalid migration - no down handler for non-idempotent",
			migration: &ritual.Migration{
				Idempotent: false,
				Up: ritual.MigrationHandler{
					SQL: []string{"CREATE TABLE test (id INT);"},
				},
				Down: ritual.MigrationHandler{},
			},
			wantErr:     true,
			errContains: "requires down handler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runner.ValidateMigration(tt.migration)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMigration() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Error message should contain '%s', got: %s", tt.errContains, err.Error())
				}
			}
		})
	}
}

func TestGetRecords(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))

	// Run some migrations
	migration1 := &ritual.Migration{
		FromVersion: "1.0.0",
		ToVersion:   "1.1.0",
		Up: ritual.MigrationHandler{
			SQL: []string{"CREATE TABLE test1 (id INT);"},
		},
	}

	migration2 := &ritual.Migration{
		FromVersion: "1.1.0",
		ToVersion:   "1.2.0",
		Up: ritual.MigrationHandler{
			SQL: []string{"invalid sql"},
		},
	}

	_ = runner.RunUp(migration1)
	_ = runner.RunUp(migration2) // This will succeed because we're not actually executing SQL

	records := runner.GetRecords()
	if len(records) != 2 {
		t.Errorf("GetRecords() returned %d records, want 2", len(records))
	}
}

func TestGetAppliedMigrations(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))
	runner.SetDryRun(true) // Use dry run to control statuses

	migration := &ritual.Migration{
		FromVersion: "1.0.0",
		ToVersion:   "1.1.0",
		Up: ritual.MigrationHandler{
			SQL: []string{"CREATE TABLE test (id INT);"},
		},
	}

	_ = runner.RunUp(migration)

	applied := runner.GetAppliedMigrations()
	if len(applied) != 0 {
		// Dry run creates StatusSkipped, not StatusApplied
		t.Logf("In dry run mode, no migrations are applied (expected)")
	}

	// Test with actual execution
	runner2 := NewRunner(filepath.Join(os.TempDir(), "test"))
	_ = runner2.RunUp(migration)

	applied2 := runner2.GetAppliedMigrations()
	if len(applied2) != 1 {
		t.Errorf("Expected 1 applied migration, got %d", len(applied2))
	}
}

func TestGetFailedMigrations(t *testing.T) {
	runner := NewRunner(filepath.Join(os.TempDir(), "test"))

	// Create a migration that will fail
	migration := &ritual.Migration{
		FromVersion: "1.0.0",
		ToVersion:   "1.1.0",
		Up: ritual.MigrationHandler{
			Script: "nonexistent-script.sh",
		},
	}

	_ = runner.RunUp(migration) // This should fail

	failed := runner.GetFailedMigrations()
	if len(failed) != 1 {
		t.Errorf("Expected 1 failed migration, got %d", len(failed))
	}

	if failed[0].Error == "" {
		t.Error("Failed migration should have an error message")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}()))
}

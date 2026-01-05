package migration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// RunnerStatus represents the status of a migration run
type RunnerStatus string

const (
	StatusPending    RunnerStatus = "pending"
	StatusApplied    RunnerStatus = "applied"
	StatusFailed     RunnerStatus = "failed"
	StatusSkipped    RunnerStatus = "skipped"
	StatusRolledBack RunnerStatus = "rolledback"
)

// MigrationRecord tracks an applied migration
type MigrationRecord struct {
	FromVersion string
	ToVersion   string
	Description string
	AppliedAt   time.Time
	Status      RunnerStatus
	Error       string
	ServerID    string
}

// Runner executes ritual migrations
type Runner struct {
	projectPath string
	dryRun      bool
	records     []*MigrationRecord
}

// NewRunner creates a new migration runner
func NewRunner(projectPath string) *Runner {
	return &Runner{
		projectPath: projectPath,
		dryRun:      false,
		records:     make([]*MigrationRecord, 0),
	}
}

// SetDryRun enables or disables dry-run mode
func (r *Runner) SetDryRun(dryRun bool) {
	r.dryRun = dryRun
}

// RunUp executes an up migration
func (r *Runner) RunUp(migration *ritual.Migration) error {
	record := &MigrationRecord{
		FromVersion: migration.FromVersion,
		ToVersion:   migration.ToVersion,
		Description: migration.Description,
		AppliedAt:   time.Now(),
		Status:      StatusPending,
	}

	if r.dryRun {
		record.Status = StatusSkipped
		r.records = append(r.records, record)
		return nil
	}

	// Execute up migration
	if err := r.executeHandler(&migration.Up); err != nil {
		record.Status = StatusFailed
		record.Error = err.Error()
		r.records = append(r.records, record)
		return fmt.Errorf("up migration failed: %w", err)
	}

	record.Status = StatusApplied
	r.records = append(r.records, record)
	return nil
}

// RunDown executes a down migration
func (r *Runner) RunDown(migration *ritual.Migration) error {
	record := &MigrationRecord{
		FromVersion: migration.ToVersion,
		ToVersion:   migration.FromVersion,
		Description: fmt.Sprintf("Rollback: %s", migration.Description),
		AppliedAt:   time.Now(),
		Status:      StatusPending,
	}

	if r.dryRun {
		record.Status = StatusSkipped
		r.records = append(r.records, record)
		return nil
	}

	// Execute down migration
	if err := r.executeHandler(&migration.Down); err != nil {
		record.Status = StatusFailed
		record.Error = err.Error()
		r.records = append(r.records, record)
		return fmt.Errorf("down migration failed: %w", err)
	}

	record.Status = StatusRolledBack
	r.records = append(r.records, record)
	return nil
}

// executeHandler executes a migration handler
func (r *Runner) executeHandler(handler *ritual.MigrationHandler) error {
	// Execute SQL statements
	if len(handler.SQL) > 0 {
		for _, sql := range handler.SQL {
			if err := r.executeSQL(sql); err != nil {
				return fmt.Errorf("SQL execution failed: %w", err)
			}
		}
	}

	// Execute script
	if handler.Script != "" {
		if err := r.executeScript(handler.Script); err != nil {
			return fmt.Errorf("script execution failed: %w", err)
		}
	}

	// Execute Go code
	if handler.GoCode != "" {
		if err := r.executeGoCode(handler.GoCode); err != nil {
			return fmt.Errorf("go code execution failed: %w", err)
		}
	}

	return nil
}

// executeSQL executes a SQL statement
func (r *Runner) executeSQL(sql string) error {
	// TODO: Implement actual SQL execution using database driver
	// For now, just validate that SQL is not empty
	if strings.TrimSpace(sql) == "" {
		return fmt.Errorf("empty SQL statement")
	}

	// In a real implementation, this would:
	// 1. Connect to the database
	// 2. Execute the SQL
	// 3. Handle errors appropriately

	return nil
}

// executeScript executes a shell script
func (r *Runner) executeScript(scriptPath string) error {
	// Resolve script path relative to project
	fullPath := filepath.Join(r.projectPath, scriptPath)

	// Check if script exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	// Make script executable
	// #nosec G302 - Migration scripts need executable permissions to run
	if err := os.Chmod(fullPath, 0750); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Execute script
	// #nosec G204 - Script path is controlled and validated from trusted migration source
	cmd := exec.Command(fullPath)
	cmd.Dir = r.projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}

// executeGoCode executes Go migration code
func (r *Runner) executeGoCode(goCodePath string) error {
	// TODO: Implement Go code execution
	// This is complex and would require:
	// 1. Building the Go code
	// 2. Loading it as a plugin
	// 3. Executing the migration function
	// For now, return not implemented

	return fmt.Errorf("go code execution not yet implemented")
}

// GetRecords returns all migration records
func (r *Runner) GetRecords() []*MigrationRecord {
	return r.records
}

// GetAppliedMigrations returns only applied migrations
func (r *Runner) GetAppliedMigrations() []*MigrationRecord {
	var applied []*MigrationRecord
	for _, record := range r.records {
		if record.Status == StatusApplied {
			applied = append(applied, record)
		}
	}
	return applied
}

// GetFailedMigrations returns only failed migrations
func (r *Runner) GetFailedMigrations() []*MigrationRecord {
	var failed []*MigrationRecord
	for _, record := range r.records {
		if record.Status == StatusFailed {
			failed = append(failed, record)
		}
	}
	return failed
}

// RunMigrationChain executes a chain of migrations in order
func (r *Runner) RunMigrationChain(migrations []*ritual.Migration, direction string) error {
	if direction != "up" && direction != "down" {
		return fmt.Errorf("invalid direction: %s (must be 'up' or 'down')", direction)
	}

	for _, migration := range migrations {
		var err error
		if direction == "up" {
			err = r.RunUp(migration)
		} else {
			err = r.RunDown(migration)
		}

		if err != nil {
			return fmt.Errorf("migration chain failed at %s->%s: %w",
				migration.FromVersion, migration.ToVersion, err)
		}
	}

	return nil
}

// ValidateMigration validates a migration before execution
func (r *Runner) ValidateMigration(migration *ritual.Migration) error {
	// Check that at least one handler type is specified
	hasHandler := len(migration.Up.SQL) > 0 ||
		migration.Up.Script != "" ||
		migration.Up.GoCode != ""

	if !hasHandler {
		return fmt.Errorf("migration has no up handler")
	}

	// Check that down handler exists for non-idempotent migrations
	if !migration.Idempotent {
		hasDownHandler := len(migration.Down.SQL) > 0 ||
			migration.Down.Script != "" ||
			migration.Down.GoCode != ""

		if !hasDownHandler {
			return fmt.Errorf("non-idempotent migration requires down handler")
		}
	}

	return nil
}

package dbops

import (
	"context"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

func TestDBExecTask_Name(t *testing.T) {
	task := &DBExecTask{}
	if got := task.Name(); got != "db-exec" {
		t.Errorf("DBExecTask.Name() = %v, want %v", got, "db-exec")
	}
}

func TestDBExecTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    *DBExecTask
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid SQL string",
			task: &DBExecTask{
				SQL: "SELECT 1",
			},
			wantErr: false,
		},
		{
			name: "valid SQL file",
			task: &DBExecTask{
				File: "schema.sql",
			},
			wantErr: false,
		},
		{
			name:    "missing both SQL and file",
			task:    &DBExecTask{},
			wantErr: true,
			errMsg:  "either sql or file is required",
		},
		{
			name: "both SQL and file provided",
			task: &DBExecTask{
				SQL:  "SELECT 1",
				File: "schema.sql",
			},
			wantErr: true,
			errMsg:  "cannot specify both sql and file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("DBExecTask.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("DBExecTask.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("DBExecTask.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestDBSeedTask_Name(t *testing.T) {
	task := &DBSeedTask{}
	if got := task.Name(); got != "db-seed" {
		t.Errorf("DBSeedTask.Name() = %v, want %v", got, "db-seed")
	}
}

func TestDBSeedTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    *DBSeedTask
		wantErr bool
	}{
		{
			name: "valid seed file",
			task: &DBSeedTask{
				File: "seeds/users.sql",
			},
			wantErr: false,
		},
		{
			name:    "missing file",
			task:    &DBSeedTask{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DBSeedTask.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBBackupTask_Name(t *testing.T) {
	task := &DBBackupTask{}
	if got := task.Name(); got != "db-backup" {
		t.Errorf("DBBackupTask.Name() = %v, want %v", got, "db-backup")
	}
}

func TestDBBackupTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    *DBBackupTask
		wantErr bool
	}{
		{
			name: "valid backup with output",
			task: &DBBackupTask{
				Output: "backup.sql",
			},
			wantErr: false,
		},
		{
			name:    "missing output",
			task:    &DBBackupTask{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DBBackupTask.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBRestoreTask_Name(t *testing.T) {
	task := &DBRestoreTask{}
	if got := task.Name(); got != "db-restore" {
		t.Errorf("DBRestoreTask.Name() = %v, want %v", got, "db-restore")
	}
}

func TestDBRestoreTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    *DBRestoreTask
		wantErr bool
	}{
		{
			name: "valid restore with file",
			task: &DBRestoreTask{
				File: "backup.sql",
			},
			wantErr: false,
		},
		{
			name:    "missing file",
			task:    &DBRestoreTask{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DBRestoreTask.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Integration tests would require a database connection.
func TestDBTasks_Execute(t *testing.T) {
	t.Skip("Database tasks require actual database connection for integration testing")

	taskCtx := tasks.NewTaskContext()
	taskCtx.SetWorkingDir("/tmp/test")
	ctx := context.Background()

	t.Run("db-exec", func(t *testing.T) {
		task := &DBExecTask{SQL: "SELECT 1"}
		_ = task.Execute(ctx, taskCtx)
	})

	t.Run("db-seed", func(t *testing.T) {
		task := &DBSeedTask{File: "seeds/test.sql"}
		_ = task.Execute(ctx, taskCtx)
	})

	t.Run("db-backup", func(t *testing.T) {
		task := &DBBackupTask{Output: "test-backup.sql"}
		_ = task.Execute(ctx, taskCtx)
	})

	t.Run("db-restore", func(t *testing.T) {
		task := &DBRestoreTask{File: "test-backup.sql"}
		_ = task.Execute(ctx, taskCtx)
	})
}

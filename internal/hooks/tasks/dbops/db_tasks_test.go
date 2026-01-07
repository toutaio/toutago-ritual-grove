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
	taskCtx := tasks.NewTaskContext()
	taskCtx.SetWorkingDir("/tmp/test")
	ctx := context.Background()

	t.Run("db-exec with SQL returns not implemented", func(t *testing.T) {
		task := &DBExecTask{SQL: "SELECT 1"}
		err := task.Execute(ctx, taskCtx)
		if err == nil {
			t.Fatal("Expected error for unimplemented task")
		}
		if err.Error() != "db-exec task not yet implemented - requires database connection" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("db-seed returns not implemented", func(t *testing.T) {
		task := &DBSeedTask{File: "/nonexistent/seeds/test.sql"}
		err := task.Execute(ctx, taskCtx)
		if err == nil {
			t.Fatal("Expected error for unimplemented task")
		}
	})

	t.Run("db-backup returns not implemented", func(t *testing.T) {
		task := &DBBackupTask{Output: "test-backup.sql"}
		err := task.Execute(ctx, taskCtx)
		if err == nil {
			t.Fatal("Expected error for unimplemented task")
		}
		if err.Error() != "db-backup task not yet implemented - requires database connection" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("db-restore returns not implemented", func(t *testing.T) {
		task := &DBRestoreTask{File: "/nonexistent/test-backup.sql"}
		err := task.Execute(ctx, taskCtx)
		if err == nil {
			t.Fatal("Expected error for unimplemented task")
		}
	})
}

func TestDBTasks_Registration(t *testing.T) {
	tests := []struct {
		name       string
		taskName   string
		config     map[string]interface{}
		shouldFail bool
	}{
		{
			name:     "db-exec with SQL",
			taskName: "db-exec",
			config: map[string]interface{}{
				"sql": "SELECT 1",
			},
			shouldFail: false,
		},
		{
			name:     "db-exec with file",
			taskName: "db-exec",
			config: map[string]interface{}{
				"file": "schema.sql",
			},
			shouldFail: false,
		},
		{
			name:       "db-exec without SQL or file",
			taskName:   "db-exec",
			config:     map[string]interface{}{},
			shouldFail: true,
		},
		{
			name:     "db-seed with file",
			taskName: "db-seed",
			config: map[string]interface{}{
				"file": "seeds.sql",
			},
			shouldFail: false,
		},
		{
			name:       "db-seed without file",
			taskName:   "db-seed",
			config:     map[string]interface{}{},
			shouldFail: true,
		},
		{
			name:     "db-backup with output",
			taskName: "db-backup",
			config: map[string]interface{}{
				"output": "backup.sql",
			},
			shouldFail: false,
		},
		{
			name:       "db-backup without output",
			taskName:   "db-backup",
			config:     map[string]interface{}{},
			shouldFail: true,
		},
		{
			name:     "db-restore with file",
			taskName: "db-restore",
			config: map[string]interface{}{
				"file": "backup.sql",
			},
			shouldFail: false,
		},
		{
			name:       "db-restore without file",
			taskName:   "db-restore",
			config:     map[string]interface{}{},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := tasks.Create(tt.taskName, tt.config)
			if tt.shouldFail {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if task == nil {
					t.Error("Expected task but got nil")
				}
			}
		})
	}
}

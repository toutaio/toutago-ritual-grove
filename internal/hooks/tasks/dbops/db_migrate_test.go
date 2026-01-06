package dbops

import (
	"context"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

func TestDBMigrateTask_Name(t *testing.T) {
	task := &DBMigrateTask{}
	if got := task.Name(); got != "db-migrate" {
		t.Errorf("DBMigrateTask.Name() = %v, want %v", got, "db-migrate")
	}
}

func TestDBMigrateTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    *DBMigrateTask
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid migrate up",
			task: &DBMigrateTask{
				Direction: "up",
			},
			wantErr: false,
		},
		{
			name: "valid migrate down",
			task: &DBMigrateTask{
				Direction: "down",
			},
			wantErr: false,
		},
		{
			name:    "missing direction",
			task:    &DBMigrateTask{},
			wantErr: true,
			errMsg:  "direction is required (up or down)",
		},
		{
			name: "invalid direction",
			task: &DBMigrateTask{
				Direction: "sideways",
			},
			wantErr: true,
			errMsg:  "direction must be 'up' or 'down'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("DBMigrateTask.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("DBMigrateTask.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("DBMigrateTask.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestDBMigrateTask_Execute(t *testing.T) {
	tests := []struct {
		name    string
		task    *DBMigrateTask
		setup   func(taskCtx *tasks.TaskContext)
		wantErr bool
	}{
		{
			name: "migrate up with sil migrator",
			task: &DBMigrateTask{
				Direction: "up",
			},
			setup: func(taskCtx *tasks.TaskContext) {
				// In a real scenario, we'd mock the migrator
			},
			wantErr: true, // Currently not implemented
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskCtx := tasks.NewTaskContext()
			taskCtx.SetWorkingDir("/tmp/test")
			if tt.setup != nil {
				tt.setup(taskCtx)
			}

			err := tt.task.Execute(context.Background(), taskCtx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBMigrateTask.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

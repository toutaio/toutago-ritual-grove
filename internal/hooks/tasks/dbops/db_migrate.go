package dbops

import (
	"context"
	"errors"
	"fmt"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

// DBMigrateTask runs database migrations using toutago-sil-migrator.
type DBMigrateTask struct {
	Direction string // "up" or "down"
	Steps     int    // Number of migrations to run (0 = all)
	Dir       string // Directory containing migration files
}

func (t *DBMigrateTask) Name() string {
	return "db-migrate"
}

func (t *DBMigrateTask) Validate() error {
	if t.Direction == "" {
		return errors.New("direction is required (up or down)")
	}
	if t.Direction != "up" && t.Direction != "down" {
		return errors.New("direction must be 'up' or 'down'")
	}
	return nil
}

func (t *DBMigrateTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	// This is a placeholder implementation.
	// In a real scenario, this would integrate with toutago-sil-migrator
	// to run database migrations.

	// For now, we'll just validate the input and return success.
	// The actual implementation will be completed when sil-migrator
	// provides a programmatic API.

	return fmt.Errorf("db-migrate task not yet implemented - requires sil-migrator integration")
}

// Register database operation tasks.
func init() {
	tasks.Register("db-migrate", func(config map[string]interface{}) (tasks.Task, error) {
		direction, _ := config["direction"].(string)
		dir, _ := config["dir"].(string)
		steps := 0
		if s, ok := config["steps"].(int); ok {
			steps = s
		} else if s, ok := config["steps"].(float64); ok {
			steps = int(s)
		}

		task := &DBMigrateTask{
			Direction: direction,
			Steps:     steps,
			Dir:       dir,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})
}

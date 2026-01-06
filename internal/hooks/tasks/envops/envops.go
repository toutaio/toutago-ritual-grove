// Package envops provides environment variable operation tasks for hooks.
package envops

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

// EnvSetTask sets an environment variable in a .env file.
type EnvSetTask struct {
	File  string
	Key   string
	Value string
}

func (t *EnvSetTask) Name() string {
	return "env-set"
}

func (t *EnvSetTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	file := t.File
	if !filepath.IsAbs(file) {
		file = filepath.Join(taskCtx.WorkingDir(), file)
	}

	// Read existing file or create new.
	var lines []string
	keyFound := false

	if content, err := os.ReadFile(file); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			line := scanner.Text()

			// Check if this line sets our key.
			if strings.HasPrefix(line, t.Key+"=") || strings.HasPrefix(line, t.Key+" =") {
				// Replace with new value.
				lines = append(lines, formatEnvLine(t.Key, t.Value))
				keyFound = true
			} else {
				// Keep existing line (including comments).
				lines = append(lines, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to scan env file: %w", err)
		}
	}

	// If key wasn't found, append it.
	if !keyFound {
		lines = append(lines, formatEnvLine(t.Key, t.Value))
	}

	// Write updated content.
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(file, []byte(content), 0600)
}

func (t *EnvSetTask) Validate() error {
	if t.Key == "" {
		return errors.New("key is required")
	}
	return nil
}

// formatEnvLine formats a key=value line, quoting if necessary.
func formatEnvLine(key, value string) string {
	// Quote value if it contains spaces.
	if strings.Contains(value, " ") {
		return fmt.Sprintf("%s=\"%s\"", key, value)
	}
	return fmt.Sprintf("%s=%s", key, value)
}

// Register all environment operation tasks.
func init() {
	tasks.Register("env-set", func(config map[string]interface{}) (tasks.Task, error) {
		file, _ := config["file"].(string)
		key, _ := config["key"].(string)
		value, _ := config["value"].(string)

		task := &EnvSetTask{
			File:  file,
			Key:   key,
			Value: value,
		}

		if err := task.Validate(); err != nil {
			return nil, fmt.Errorf("invalid env-set task config: %w", err)
		}

		return task, nil
	})
}

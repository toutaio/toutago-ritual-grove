package validationops

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

// ValidateGoVersionTask checks if Go version meets minimum requirement.
type ValidateGoVersionTask struct {
	MinVersion string // Minimum Go version (e.g., "1.21")
}

func (t *ValidateGoVersionTask) Name() string {
	return "validate-go-version"
}

func (t *ValidateGoVersionTask) Validate() error {
	if t.MinVersion == "" {
		return errors.New("min_version is required")
	}
	return nil
}

func (t *ValidateGoVersionTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	// Get current Go version.
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New("failed to read build info")
	}

	current := info.GoVersion
	// Remove "go" prefix if present.
	current = strings.TrimPrefix(current, "go")

	if !isVersionGreaterOrEqual(current, t.MinVersion) {
		return fmt.Errorf("Go version %s is required, but %s is installed", t.MinVersion, current)
	}

	return nil
}

// ValidateDependenciesTask checks if required commands are available.
type ValidateDependenciesTask struct {
	Dependencies []string // List of required commands
}

func (t *ValidateDependenciesTask) Name() string {
	return "validate-dependencies"
}

func (t *ValidateDependenciesTask) Validate() error {
	if len(t.Dependencies) == 0 {
		return errors.New("dependencies list is required")
	}
	return nil
}

func (t *ValidateDependenciesTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	var missing []string

	for _, dep := range t.Dependencies {
		if _, err := exec.LookPath(dep); err != nil {
			missing = append(missing, dep)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required dependencies: %s", strings.Join(missing, ", "))
	}

	return nil
}

// ValidateConfigTask checks if configuration file exists and is valid.
type ValidateConfigTask struct {
	File string // Path to config file
}

func (t *ValidateConfigTask) Name() string {
	return "validate-config"
}

func (t *ValidateConfigTask) Validate() error {
	if t.File == "" {
		return errors.New("file is required")
	}
	return nil
}

func (t *ValidateConfigTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	if _, err := os.Stat(t.File); err != nil {
		return fmt.Errorf("config file not found: %s", t.File)
	}

	// Basic validation - file exists and is readable.
	file, err := os.Open(t.File)
	if err != nil {
		return fmt.Errorf("cannot read config file: %w", err)
	}
	defer file.Close()

	return nil
}

// EnvCheckTask validates required environment variables are set.
type EnvCheckTask struct {
	Required []string // List of required environment variables
}

func (t *EnvCheckTask) Name() string {
	return "env-check"
}

func (t *EnvCheckTask) Validate() error {
	if len(t.Required) == 0 {
		return errors.New("required list is required")
	}
	return nil
}

func (t *EnvCheckTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	var missing []string

	for _, envVar := range t.Required {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// PortCheckTask checks if a port is available.
type PortCheckTask struct {
	Port int // Port number to check
}

func (t *PortCheckTask) Name() string {
	return "port-check"
}

func (t *PortCheckTask) Validate() error {
	if t.Port <= 0 || t.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	return nil
}

func (t *PortCheckTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	address := fmt.Sprintf(":%d", t.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("port %d is not available: %w", t.Port, err)
	}
	listener.Close()

	return nil
}

// isVersionGreaterOrEqual compares two semantic versions.
func isVersionGreaterOrEqual(current, required string) bool {
	// Parse versions.
	currentParts := parseVersion(current)
	requiredParts := parseVersion(required)

	// Compare major, minor, patch.
	for i := 0; i < 3 && i < len(currentParts) && i < len(requiredParts); i++ {
		if currentParts[i] > requiredParts[i] {
			return true
		}
		if currentParts[i] < requiredParts[i] {
			return false
		}
	}

	return true
}

// parseVersion extracts major.minor.patch numbers from version string.
func parseVersion(version string) []int {
	// Remove any non-numeric prefix.
	re := regexp.MustCompile(`(\d+)\.(\d+)(?:\.(\d+))?`)
	matches := re.FindStringSubmatch(version)
	if matches == nil {
		return []int{0, 0, 0}
	}

	parts := make([]int, 3)
	for i := 1; i < len(matches) && i <= 3; i++ {
		if matches[i] != "" {
			parts[i-1], _ = strconv.Atoi(matches[i])
		}
	}

	return parts
}

// Register validation tasks.
func init() {
	tasks.Register("validate-go-version", func(config map[string]interface{}) (tasks.Task, error) {
		minVersion, _ := config["min_version"].(string)

		task := &ValidateGoVersionTask{
			MinVersion: minVersion,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})

	tasks.Register("validate-dependencies", func(config map[string]interface{}) (tasks.Task, error) {
		var dependencies []string
		if deps, ok := config["dependencies"].([]interface{}); ok {
			for _, dep := range deps {
				if s, ok := dep.(string); ok {
					dependencies = append(dependencies, s)
				}
			}
		}

		task := &ValidateDependenciesTask{
			Dependencies: dependencies,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})

	tasks.Register("validate-config", func(config map[string]interface{}) (tasks.Task, error) {
		file, _ := config["file"].(string)

		task := &ValidateConfigTask{
			File: file,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})

	tasks.Register("env-check", func(config map[string]interface{}) (tasks.Task, error) {
		var required []string
		if reqs, ok := config["required"].([]interface{}); ok {
			for _, req := range reqs {
				if s, ok := req.(string); ok {
					required = append(required, s)
				}
			}
		}

		task := &EnvCheckTask{
			Required: required,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})

	tasks.Register("port-check", func(config map[string]interface{}) (tasks.Task, error) {
		var port int
		switch v := config["port"].(type) {
		case int:
			port = v
		case float64:
			port = int(v)
		}

		task := &PortCheckTask{
			Port: port,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})
}

package commands

import (
	"fmt"
	"os"

	"github.com/toutaio/toutago-ritual-grove/internal/generator"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// CreateOptions contains options for the create command
type CreateOptions struct {
	SkipQuestionnaire bool
	UseDefaults       bool
	ConfigFile        string
}

// CreateHandler handles project creation from rituals
type CreateHandler struct{}

// NewCreateHandler creates a new create command handler
func NewCreateHandler() *CreateHandler {
	return &CreateHandler{}
}

// Execute creates a project from a ritual with given answers
func (h *CreateHandler) Execute(ritualPath, targetPath string, answers map[string]interface{}, opts CreateOptions) error {
	// Validate inputs
	if ritualPath == "" {
		return fmt.Errorf("ritual path is required")
	}

	if targetPath == "" {
		return fmt.Errorf("target path is required")
	}

	// Check ritual exists
	if _, err := os.Stat(ritualPath); os.IsNotExist(err) {
		return fmt.Errorf("ritual not found: %s", ritualPath)
	}

	// Load ritual
	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	if err != nil {
		return fmt.Errorf("failed to load ritual: %w", err)
	}

	// Get default answers if needed
	defaultAnswers, err := h.ExtractDefaultAnswers(ritualPath)
	if err != nil {
		return fmt.Errorf("failed to extract defaults: %w", err)
	}

	// Merge user answers with defaults
	finalAnswers := h.MergeAnswersWithDefaults(answers, defaultAnswers)

	// Convert to Variables
	vars := generator.NewVariables()
	for key, value := range finalAnswers {
		vars.Set(key, value)
	}

	// Generate project
	scaffolder := generator.NewProjectScaffolder()
	if err := scaffolder.GenerateFromRitual(targetPath, ritualPath, manifest, vars); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	return nil
}

// ExecuteWithDefaults creates a project using all default values
func (h *CreateHandler) ExecuteWithDefaults(ritualPath, targetPath string) error {
	return h.Execute(ritualPath, targetPath, nil, CreateOptions{
		UseDefaults: true,
	})
}

// ExtractDefaultAnswers extracts default values from ritual questions
func (h *CreateHandler) ExtractDefaultAnswers(ritualPath string) (map[string]interface{}, error) {
	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	if err != nil {
		return nil, err
	}

	defaults := make(map[string]interface{})

	for _, question := range manifest.Questions {
		if question.Default != nil {
			defaults[question.Name] = question.Default
		}
	}

	return defaults, nil
}

// MergeAnswersWithDefaults merges user answers with defaults
// User answers take precedence over defaults
func (h *CreateHandler) MergeAnswersWithDefaults(userAnswers, defaults map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Add all defaults
	for key, value := range defaults {
		merged[key] = value
	}

	// Override with user answers
	for key, value := range userAnswers {
		merged[key] = value
	}

	return merged
}

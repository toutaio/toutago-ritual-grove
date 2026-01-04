// Package questionnaire provides interactive question handling for rituals.
package questionnaire

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/toutaio/toutago-ritual-grove/internal/ritual"
)

// CLIAdapter is a command-line adapter for the questionnaire system
type CLIAdapter struct {
	controller *Controller
	useDefaults bool
	configFile  string
	answers     map[string]interface{}
}

// NewCLIAdapter creates a new CLI adapter for questionnaire
func NewCLIAdapter(questions []*ritual.Question, useDefaults bool, configFile string) (*CLIAdapter, error) {
	controller := NewController(questions)
	
	adapter := &CLIAdapter{
		controller: controller,
		useDefaults: useDefaults,
		configFile:  configFile,
		answers:     make(map[string]interface{}),
	}
	
	// Load answers from config file if provided
	if configFile != "" {
		if err := adapter.loadConfig(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}
	
	return adapter, nil
}

// Run executes the questionnaire and returns collected answers
func (a *CLIAdapter) Run() (map[string]interface{}, error) {
	total := len(a.controller.questions)
	current := 0
	currentGroup := ""
	
	for {
		question := a.controller.NextQuestion()
		if question == nil {
			break
		}
		
		current++
		
		// Show section header if group changed
		if question.Group != "" && question.Group != currentGroup {
			currentGroup = question.Group
			fmt.Printf("\n=== %s ===\n", currentGroup)
		}
		
		// Skip if we already have an answer (from config or --yes)
		if answer, exists := a.answers[question.ID]; exists {
			if err := a.controller.SetAnswer(question.ID, answer); err != nil {
				return nil, err
			}
			continue
		}
		
		// Use defaults if --yes flag is set
		if a.useDefaults && question.Default != nil {
			if err := a.controller.SetAnswer(question.ID, question.Default); err != nil {
				return nil, err
			}
			a.answers[question.ID] = question.Default
			continue
		}
		
		// Show progress
		fmt.Printf("\n[%d/%d] ", current, total)
			break
		}
		
		current++
		
		// Skip if we already have an answer (from config or --yes)
		if answer, exists := a.answers[question.ID]; exists {
			if err := a.controller.SetAnswer(question.ID, answer); err != nil {
				return nil, err
			}
			continue
		}
		
		// Use defaults if --yes flag is set
		if a.useDefaults && question.Default != nil {
			if err := a.controller.SetAnswer(question.ID, question.Default); err != nil {
				return nil, err
			}
			a.answers[question.ID] = question.Default
			continue
		}
		
		// Show progress
		fmt.Printf("\n[%d/%d] ", current, total)
		
		// Ask the question
		answer, err := a.askQuestion(question)
		if err != nil {
			return nil, err
		}
		
		// Set the answer
		if err := a.controller.SetAnswer(question.ID, answer); err != nil {
			return nil, err
		}
		
		a.answers[question.ID] = answer
	}
	
	return a.controller.GetAnswers(), nil
}

// askQuestion prompts the user for a single question
func (a *CLIAdapter) askQuestion(q *ritual.Question) (interface{}, error) {
	var prompt survey.Prompt
	var answer interface{}
	
	// Build the message
	message := q.Label
	if q.Required {
		message += " *"
	}
	if q.Help != "" {
		message += fmt.Sprintf("\n  %s", q.Help)
	}
	
	// Create appropriate prompt based on question type
	switch q.Type {
	case "text", "path", "url", "email":
		defaultStr := ""
		if q.Default != nil {
			defaultStr = fmt.Sprint(q.Default)
		}
		
		textPrompt := &survey.Input{
			Message: message,
			Default: defaultStr,
		}
		
		var result string
		if err := survey.AskOne(textPrompt, &result, a.getValidatorOpts(q)...); err != nil {
			return nil, err
		}
		answer = result
		
	case "password":
		passPrompt := &survey.Password{
			Message: message,
		}
		
		var result string
		if err := survey.AskOne(passPrompt, &result, a.getValidatorOpts(q)...); err != nil {
			return nil, err
		}
		answer = result
		
	case "boolean":
		defaultBool := false
		if q.Default != nil {
			if b, ok := q.Default.(bool); ok {
				defaultBool = b
			}
		}
		
		boolPrompt := &survey.Confirm{
			Message: message,
			Default: defaultBool,
		}
		
		var result bool
		if err := survey.AskOne(boolPrompt, &result); err != nil {
			return nil, err
		}
		answer = result
		
	case "choice":
		if q.Choices == nil || len(q.Choices) == 0 {
			return nil, fmt.Errorf("no choices provided for question %s", q.ID)
		}
		
		defaultStr := ""
		if q.Default != nil {
			defaultStr = fmt.Sprint(q.Default)
		}
		
		selectPrompt := &survey.Select{
			Message: message,
			Options: q.Choices,
			Default: defaultStr,
		}
		
		var result string
		if err := survey.AskOne(selectPrompt, &result); err != nil {
			return nil, err
		}
		answer = result
		
	case "multi-choice":
		if q.Choices == nil || len(q.Choices) == 0 {
			return nil, fmt.Errorf("no choices provided for question %s", q.ID)
		}
		
		var defaults []string
		if q.Default != nil {
			if arr, ok := q.Default.([]interface{}); ok {
				for _, v := range arr {
					defaults = append(defaults, fmt.Sprint(v))
				}
			}
		}
		
		multiPrompt := &survey.MultiSelect{
			Message: message,
			Options: q.Choices,
			Default: defaults,
		}
		
		var result []string
		if err := survey.AskOne(multiPrompt, &result); err != nil {
			return nil, err
		}
		answer = result
		
	case "number":
		defaultStr := ""
		if q.Default != nil {
			defaultStr = fmt.Sprint(q.Default)
		}
		
		numberPrompt := &survey.Input{
			Message: message,
			Default: defaultStr,
		}
		
		var resultStr string
		opts := append(a.getValidatorOpts(q), survey.WithValidator(func(ans interface{}) error {
			str := ans.(string)
			if str == "" && !q.Required {
				return nil
			}
			if _, err := strconv.ParseFloat(str, 64); err != nil {
				return fmt.Errorf("invalid number")
			}
			return nil
		}))
		
		if err := survey.AskOne(numberPrompt, &resultStr, opts...); err != nil {
			return nil, err
		}
		
		if resultStr == "" {
			answer = nil
		} else {
			num, _ := strconv.ParseFloat(resultStr, 64)
			answer = num
		}
		
	default:
		return nil, fmt.Errorf("unsupported question type: %s", q.Type)
	}
	
	// Validate the answer
	if err := a.controller.validator.Validate(q, answer); err != nil {
		fmt.Printf("  Error: %s\n", err)
		return a.askQuestion(q) // Retry
	}
	
	return answer, nil
}

// getValidatorOpts creates survey validator options from question validation rules
func (a *CLIAdapter) getValidatorOpts(q *ritual.Question) []survey.AskOpt {
	var opts []survey.AskOpt
	
	if q.Required {
		opts = append(opts, survey.WithValidator(survey.Required))
	}
	
	if q.Validation != nil {
		if q.Validation.Pattern != "" {
			opts = append(opts, survey.WithValidator(func(ans interface{}) error {
				return a.controller.validator.Validate(q, ans)
			}))
		}
		
		if q.Validation.MinLength > 0 || q.Validation.MaxLength > 0 {
			opts = append(opts, survey.WithValidator(func(ans interface{}) error {
				str := fmt.Sprint(ans)
				if q.Validation.MinLength > 0 && len(str) < q.Validation.MinLength {
					return fmt.Errorf("minimum length is %d", q.Validation.MinLength)
				}
				if q.Validation.MaxLength > 0 && len(str) > q.Validation.MaxLength {
					return fmt.Errorf("maximum length is %d", q.Validation.MaxLength)
				}
				return nil
			}))
		}
	}
	
	return opts
}

// loadConfig loads answers from a configuration file
func (a *CLIAdapter) loadConfig() error {
	// TODO: Implement YAML/JSON config loading
	// For now, just return nil
	return nil
}

// SaveAnswers persists answers to .ritual/answers.yaml
func (a *CLIAdapter) SaveAnswers(path string) error {
	// TODO: Implement answer persistence
	return nil
}

// printProgress shows a progress indicator
func printProgress(current, total int) {
	percentage := float64(current) / float64(total) * 100
	fmt.Printf("Progress: [%d/%d] %.0f%%\n", current, total, percentage)
}

// formatError formats validation errors for display
func formatError(err error) string {
	msg := err.Error()
	// Clean up common error message patterns
	msg = strings.ReplaceAll(msg, "validation error: ", "")
	msg = strings.TrimPrefix(msg, "error: ")
	return msg
}

// RunWithoutInteraction runs the questionnaire using only defaults and config
func (a *CLIAdapter) RunWithoutInteraction() (map[string]interface{}, error) {
	for {
		question := a.controller.NextQuestion()
		if question == nil {
			break
		}
		
		var answer interface{}
		
		// Check config first
		if configAnswer, exists := a.answers[question.ID]; exists {
			answer = configAnswer
		} else if question.Default != nil {
			answer = question.Default
		} else if question.Required {
			return nil, fmt.Errorf("no answer provided for required question: %s", question.ID)
		} else {
			answer = nil
		}
		
		if err := a.controller.SetAnswer(question.ID, answer); err != nil {
			return nil, err
		}
	}
	
	return a.controller.GetAnswers(), nil
}

// GetController returns the underlying controller
func (a *CLIAdapter) GetController() *Controller {
	return a.controller
}

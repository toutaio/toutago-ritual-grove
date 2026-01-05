package questionnaire

import (
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// Validator validates question answers
type Validator struct {
	customValidators map[string]ValidationFunc
}

// ValidationFunc is a custom validation function
type ValidationFunc func(value interface{}) error

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		customValidators: make(map[string]ValidationFunc),
	}
}

// RegisterCustomValidator registers a custom validation function
func (v *Validator) RegisterCustomValidator(name string, fn ValidationFunc) {
	v.customValidators[name] = fn
}

// ValidateAnswer validates an answer against question constraints
func (v *Validator) ValidateAnswer(question *ritual.Question, value interface{}) error {
	// Check required
	if question.Required && v.isEmpty(value) {
		return fmt.Errorf("answer is required")
	}

	// Skip further validation if empty and not required
	if !question.Required && v.isEmpty(value) {
		return nil
	}

	// Type-specific validation
	if err := v.validateByType(question, value); err != nil {
		return err
	}

	// Custom validation rules
	if question.Validate != nil {
		if err := v.validateRules(question.Validate, value); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch val := value.(type) {
	case string:
		return val == ""
	case []string:
		return len(val) == 0
	case []interface{}:
		return len(val) == 0
	default:
		return false
	}
}

func (v *Validator) validateByType(question *ritual.Question, value interface{}) error {
	switch question.Type {
	case ritual.QuestionTypeText, ritual.QuestionTypePassword:
		return v.validateText(value)

	case ritual.QuestionTypeChoice:
		return v.validateChoice(question, value)

	case ritual.QuestionTypeMultiChoice:
		return v.validateMultiChoice(question, value)

	case ritual.QuestionTypeBoolean:
		return v.validateBoolean(value)

	case ritual.QuestionTypeNumber:
		return v.validateNumber(value)

	case ritual.QuestionTypePath:
		return v.validatePath(value)

	case ritual.QuestionTypeURL:
		return v.validateURL(value)

	case ritual.QuestionTypeEmail:
		return v.validateEmail(value)

	default:
		return nil
	}
}

func (v *Validator) validateText(value interface{}) error {
	_, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}
	return nil
}

func (v *Validator) validateChoice(question *ritual.Question, value interface{}) error {
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	if len(question.Choices) == 0 {
		return nil
	}

	for _, choice := range question.Choices {
		if choice == strVal {
			return nil
		}
	}

	return fmt.Errorf("invalid choice: %s (must be one of: %v)", strVal, question.Choices)
}

func (v *Validator) validateMultiChoice(question *ritual.Question, value interface{}) error {
	var choices []string

	switch val := value.(type) {
	case []string:
		choices = val
	case []interface{}:
		choices = make([]string, len(val))
		for i, item := range val {
			strItem, ok := item.(string)
			if !ok {
				return fmt.Errorf("expected string values in array")
			}
			choices[i] = strItem
		}
	default:
		return fmt.Errorf("expected array of strings")
	}

	if len(question.Choices) == 0 {
		return nil
	}

	// Validate each choice
	for _, choice := range choices {
		valid := false
		for _, validChoice := range question.Choices {
			if choice == validChoice {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid choice: %s (must be one of: %v)", choice, question.Choices)
		}
	}

	return nil
}

func (v *Validator) validateBoolean(value interface{}) error {
	_, ok := value.(bool)
	if !ok {
		return fmt.Errorf("expected boolean value")
	}
	return nil
}

func (v *Validator) validateNumber(value interface{}) error {
	switch val := value.(type) {
	case int, int64, float64:
		return nil
	case string:
		// Try to parse as number
		if _, err := strconv.ParseFloat(val, 64); err != nil {
			return fmt.Errorf("expected numeric value")
		}
		return nil
	default:
		return fmt.Errorf("expected numeric value")
	}
}

func (v *Validator) validatePath(value interface{}) error {
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	// Check if path is valid (not necessarily existing)
	cleanPath := filepath.Clean(strVal)
	if cleanPath == "." || cleanPath == "/" {
		return nil
	}

	// Check if it's an absolute or valid relative path
	if filepath.IsAbs(cleanPath) || filepath.IsLocal(cleanPath) {
		return nil
	}

	return fmt.Errorf("invalid path: %s", strVal)
}

func (v *Validator) validateURL(value interface{}) error {
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	parsedURL, err := url.Parse(strVal)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must have a scheme (http, https, etc.)")
	}

	return nil
}

func (v *Validator) validateEmail(value interface{}) error {
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	_, err := mail.ParseAddress(strVal)
	if err != nil {
		return fmt.Errorf("invalid email address: %w", err)
	}

	return nil
}

func (v *Validator) validateRules(rules *ritual.ValidationRule, value interface{}) error {
	// Pattern validation
	if rules.Pattern != "" {
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("pattern validation requires string value")
		}

		matched, err := regexp.MatchString(rules.Pattern, strVal)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf("value does not match required pattern")
		}
	}

	// Min/max validation for numbers
	if rules.Min != nil || rules.Max != nil {
		numVal, err := v.toNumber(value)
		if err != nil {
			return err
		}

		if rules.Min != nil && numVal < float64(*rules.Min) {
			return fmt.Errorf("value must be at least %d", *rules.Min)
		}
		if rules.Max != nil && numVal > float64(*rules.Max) {
			return fmt.Errorf("value must be at most %d", *rules.Max)
		}
	}

	// Min/max length validation for strings
	if rules.MinLen != nil || rules.MaxLen != nil {
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("length validation requires string value")
		}

		length := len(strVal)
		if rules.MinLen != nil && length < *rules.MinLen {
			return fmt.Errorf("value must be at least %d characters", *rules.MinLen)
		}
		if rules.MaxLen != nil && length > *rules.MaxLen {
			return fmt.Errorf("value must be at most %d characters", *rules.MaxLen)
		}
	}

	// Custom validator
	if rules.Custom != "" {
		if fn, exists := v.customValidators[rules.Custom]; exists {
			if err := fn(value); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("custom validator not found: %s", rules.Custom)
		}
	}

	return nil
}

func (v *Validator) toNumber(value interface{}) (float64, error) {
	switch val := value.(type) {
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert to number")
	}
}

// ValidatePath checks if a path exists and is accessible
func ValidatePath(value interface{}) error {
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	_, err := os.Stat(strVal)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", strVal)
		}
		return fmt.Errorf("cannot access path: %w", err)
	}

	return nil
}

// ValidateWritablePath checks if a path is writable
func ValidateWritablePath(value interface{}) error {
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	// Try to create a temp file to test writability
	testFile := filepath.Join(strVal, ".writetest")
	// #nosec G304 - testFile is created in a temporary directory for validation
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("path is not writable: %w", err)
	}
	_ = f.Close()
	_ = os.Remove(testFile)

	return nil
}

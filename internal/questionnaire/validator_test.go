package questionnaire

import (
	"fmt"
	"os"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestValidator_RequiredField(t *testing.T) {
	validator := NewValidator()

	question := &ritual.Question{
		Name:     "required_field",
		Required: true,
		Type:     ritual.QuestionTypeText,
	}

	// Empty value should fail
	err := validator.ValidateAnswer(question, "")
	if err == nil {
		t.Error("Expected error for empty required field")
	}

	// Valid value should pass
	err = validator.ValidateAnswer(question, "value")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestValidator_TextType(t *testing.T) {
	validator := NewValidator()

	question := &ritual.Question{
		Name: "text_field",
		Type: ritual.QuestionTypeText,
	}

	// String should pass
	err := validator.ValidateAnswer(question, "hello")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Non-string should fail
	err = validator.ValidateAnswer(question, 123)
	if err == nil {
		t.Error("Expected error for non-string value")
	}
}

func TestValidator_ChoiceType(t *testing.T) {
	validator := NewValidator()

	question := &ritual.Question{
		Name:    "choice_field",
		Type:    ritual.QuestionTypeChoice,
		Choices: []string{"option1", "option2", "option3"},
	}

	// Valid choice should pass
	err := validator.ValidateAnswer(question, "option1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Invalid choice should fail
	err = validator.ValidateAnswer(question, "invalid")
	if err == nil {
		t.Error("Expected error for invalid choice")
	}
}

func TestValidator_MultiChoiceType(t *testing.T) {
	validator := NewValidator()

	question := &ritual.Question{
		Name:    "multi_choice_field",
		Type:    ritual.QuestionTypeMultiChoice,
		Choices: []string{"option1", "option2", "option3"},
	}

	// Valid choices should pass
	err := validator.ValidateAnswer(question, []string{"option1", "option2"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Invalid choice should fail
	err = validator.ValidateAnswer(question, []string{"option1", "invalid"})
	if err == nil {
		t.Error("Expected error for invalid choice")
	}
}

func TestValidator_BooleanType(t *testing.T) {
	validator := NewValidator()

	question := &ritual.Question{
		Name: "bool_field",
		Type: ritual.QuestionTypeBoolean,
	}

	// Boolean should pass
	err := validator.ValidateAnswer(question, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Non-boolean should fail
	err = validator.ValidateAnswer(question, "yes")
	if err == nil {
		t.Error("Expected error for non-boolean value")
	}
}

func TestValidator_NumberType(t *testing.T) {
	validator := NewValidator()

	question := &ritual.Question{
		Name: "number_field",
		Type: ritual.QuestionTypeNumber,
	}

	// Int should pass
	err := validator.ValidateAnswer(question, 42)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Float should pass
	err = validator.ValidateAnswer(question, 3.14)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Numeric string should pass
	err = validator.ValidateAnswer(question, "123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Non-numeric should fail
	err = validator.ValidateAnswer(question, "abc")
	if err == nil {
		t.Error("Expected error for non-numeric value")
	}
}

func TestValidator_PathType(t *testing.T) {
	validator := NewValidator()

	question := &ritual.Question{
		Name: "path_field",
		Type: ritual.QuestionTypePath,
	}

	// Valid path should pass
	err := validator.ValidateAnswer(question, "/path/to/file")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Relative path should pass
	err = validator.ValidateAnswer(question, "./relative/path")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestValidator_URLType(t *testing.T) {
	validator := NewValidator()

	question := &ritual.Question{
		Name: "url_field",
		Type: ritual.QuestionTypeURL,
	}

	// Valid URL should pass
	err := validator.ValidateAnswer(question, "https://example.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// URL without scheme should fail
	err = validator.ValidateAnswer(question, "example.com")
	if err == nil {
		t.Error("Expected error for URL without scheme")
	}

	// Invalid URL should fail
	err = validator.ValidateAnswer(question, "ht!tp://invalid")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestValidator_EmailType(t *testing.T) {
	validator := NewValidator()

	question := &ritual.Question{
		Name: "email_field",
		Type: ritual.QuestionTypeEmail,
	}

	// Valid email should pass
	err := validator.ValidateAnswer(question, "user@example.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Invalid email should fail
	err = validator.ValidateAnswer(question, "invalid-email")
	if err == nil {
		t.Error("Expected error for invalid email")
	}
}

func TestValidator_PatternValidation(t *testing.T) {
	validator := NewValidator()

	minLen := 3
	maxLen := 10

	question := &ritual.Question{
		Name: "pattern_field",
		Type: ritual.QuestionTypeText,
		Validate: &ritual.ValidationRule{
			Pattern: `^[a-z]+$`,
			MinLen:  &minLen,
			MaxLen:  &maxLen,
		},
	}

	// Valid value should pass
	err := validator.ValidateAnswer(question, "hello")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Invalid pattern should fail
	err = validator.ValidateAnswer(question, "Hello")
	if err == nil {
		t.Error("Expected error for pattern mismatch")
	}

	// Too short should fail
	err = validator.ValidateAnswer(question, "ab")
	if err == nil {
		t.Error("Expected error for too short")
	}

	// Too long should fail
	err = validator.ValidateAnswer(question, "verylongstring")
	if err == nil {
		t.Error("Expected error for too long")
	}
}

func TestValidator_MinMaxValidation(t *testing.T) {
	validator := NewValidator()

	min := 10
	max := 100

	question := &ritual.Question{
		Name: "number_field",
		Type: ritual.QuestionTypeNumber,
		Validate: &ritual.ValidationRule{
			Min: &min,
			Max: &max,
		},
	}

	// Valid value should pass
	err := validator.ValidateAnswer(question, 50)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Too small should fail
	err = validator.ValidateAnswer(question, 5)
	if err == nil {
		t.Error("Expected error for value below min")
	}

	// Too large should fail
	err = validator.ValidateAnswer(question, 150)
	if err == nil {
		t.Error("Expected error for value above max")
	}
}

func TestValidator_CustomValidator(t *testing.T) {
	validator := NewValidator()

	// Register custom validator
	validator.RegisterCustomValidator("even_number", func(value interface{}) error {
		num, ok := value.(int)
		if !ok {
			return nil
		}
		if num%2 != 0 {
			return fmt.Errorf("number must be even")
		}
		return nil
	})

	question := &ritual.Question{
		Name: "even_field",
		Type: ritual.QuestionTypeNumber,
		Validate: &ritual.ValidationRule{
			Custom: "even_number",
		},
	}

	// Even number should pass
	err := validator.ValidateAnswer(question, 42)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Odd number should fail
	err = validator.ValidateAnswer(question, 41)
	if err == nil {
		t.Error("Expected error for odd number")
	}
}

func TestValidator_OptionalField(t *testing.T) {
	validator := NewValidator()

	question := &ritual.Question{
		Name:     "optional_field",
		Type:     ritual.QuestionTypeText,
		Required: false,
	}

	// Empty value should pass for optional field
	err := validator.ValidateAnswer(question, "")
	if err != nil {
		t.Errorf("Unexpected error for optional field: %v", err)
	}
}

func TestValidator_toNumber(t *testing.T) {
	v := &Validator{}
	
	tests := []struct {
		name      string
		input     interface{}
		wantFloat float64
		wantErr   bool
	}{
		{"int", 42, 42.0, false},
		{"int64", int64(100), 100.0, false},
		{"float64", 3.14, 3.14, false},
		{"string valid", "123", 123.0, false},
		{"string invalid", "abc", 0, true},
		{"bool", true, 0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFloat, gotErr := v.toNumber(tt.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("toNumber(%v) error = %v, wantErr %v", tt.input, gotErr, tt.wantErr)
				return
			}
			if !tt.wantErr && gotFloat != tt.wantFloat {
				t.Errorf("toNumber(%v) = %v, want %v", tt.input, gotFloat, tt.wantFloat)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test"
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid absolute path", tmpFile, false},
		{"valid relative path", "./validator_test.go", false},
		{"invalid path with null", "/tmp/test\x00", true},
		{"empty path", "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidateWritablePath(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"writable dir", tmpDir, false},
		{"non-existent path", "/nonexistent/path/that/does/not/exist", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWritablePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWritablePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

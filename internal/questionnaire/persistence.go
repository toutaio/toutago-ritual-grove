package questionnaire

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// SecretPlaceholder is used to mask secrets in saved answers
	SecretPlaceholder = "<SECRET_FROM_ENV>"
)

// AnswerPersistence handles saving and loading questionnaire answers
type AnswerPersistence struct {
	filePath string
}

// NewAnswerPersistence creates a new answer persistence handler
func NewAnswerPersistence(filePath string) *AnswerPersistence {
	return &AnswerPersistence{
		filePath: filePath,
	}
}

// Save writes answers to the configured file path
func (p *AnswerPersistence) Save(answers map[string]interface{}) error {
	// Ensure directory exists
	dir := filepath.Dir(p.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(answers)
	if err != nil {
		return fmt.Errorf("failed to marshal answers: %w", err)
	}

	// Write to file
	if err := os.WriteFile(p.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write answers file: %w", err)
	}

	return nil
}

// Load reads answers from the configured file path
func (p *AnswerPersistence) Load() (map[string]interface{}, error) {
	// Read file
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read answers file: %w", err)
	}

	// Unmarshal from YAML
	var answers map[string]interface{}
	if err := yaml.Unmarshal(data, &answers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal answers: %w", err)
	}

	return answers, nil
}

// SaveWithSecrets saves answers with specified fields masked as secrets
func (p *AnswerPersistence) SaveWithSecrets(answers map[string]interface{}, secretFields []string) error {
	// Create a copy of answers with secrets masked
	maskedAnswers := make(map[string]interface{})
	secretSet := make(map[string]bool)

	for _, field := range secretFields {
		secretSet[field] = true
	}

	for key, value := range answers {
		if secretSet[key] {
			// Store the environment variable name hint
			envVarName := toEnvVarName(key)
			maskedAnswers[key] = fmt.Sprintf("%s (from $%s)", SecretPlaceholder, envVarName)
		} else {
			maskedAnswers[key] = value
		}
	}

	return p.Save(maskedAnswers)
}

// LoadWithSecrets loads answers and restores secrets from environment variables
func (p *AnswerPersistence) LoadWithSecrets(secretEnvMapping map[string]string) (map[string]interface{}, error) {
	// Load answers
	answers, err := p.Load()
	if err != nil {
		return nil, err
	}

	// Restore secrets from environment
	for field, envVar := range secretEnvMapping {
		if value, exists := os.LookupEnv(envVar); exists {
			answers[field] = value
		} else {
			// Check if the answer contains the placeholder
			if answerValue, ok := answers[field].(string); ok {
				if strings.Contains(answerValue, SecretPlaceholder) {
					// Secret was masked but env var not set
					delete(answers, field)
				}
			}
		}
	}

	return answers, nil
}

// Exists checks if the answers file exists
func (p *AnswerPersistence) Exists() bool {
	_, err := os.Stat(p.filePath)
	return err == nil
}

// Delete removes the answers file
func (p *AnswerPersistence) Delete() error {
	if !p.Exists() {
		return nil // Nothing to delete
	}

	if err := os.Remove(p.filePath); err != nil {
		return fmt.Errorf("failed to delete answers file: %w", err)
	}

	return nil
}

// toEnvVarName converts a field name to environment variable naming convention
// Example: "db_password" -> "DB_PASSWORD"
func toEnvVarName(fieldName string) string {
	return strings.ToUpper(fieldName)
}

package generator

import (
	"fmt"
	"strings"
	"text/template"
)

// evaluateCondition evaluates a template condition expression and returns true/false.
// Empty conditions return true (always generate).
// Conditions use Go template syntax: {{ eq .field "value" }}
func evaluateCondition(condition string, variables map[string]interface{}) (bool, error) {
	// Empty condition means always generate
	if condition == "" {
		return true, nil
	}

	// Trim whitespace
	condition = strings.TrimSpace(condition)

	// Create a template to evaluate the condition
	tmpl, err := template.New("condition").Parse(condition)
	if err != nil {
		return false, fmt.Errorf("failed to parse condition template: %w", err)
	}

	// Render the condition with variables
	var buf strings.Builder
	if err := tmpl.Execute(&buf, variables); err != nil {
		return false, fmt.Errorf("failed to execute condition template: %w", err)
	}

	// Parse the result as boolean
	result := strings.TrimSpace(buf.String())

	// Convert string result to boolean
	switch result {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no", "":
		return false, nil
	default:
		// If result is not empty, consider it truthy
		return result != "", nil
	}
}

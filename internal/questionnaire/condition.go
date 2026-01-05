package questionnaire

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// ConditionEvaluator evaluates conditional expressions
type ConditionEvaluator struct{}

// NewConditionEvaluator creates a new condition evaluator
func NewConditionEvaluator() *ConditionEvaluator {
	return &ConditionEvaluator{}
}

// Evaluate evaluates a condition against current answers
func (ce *ConditionEvaluator) Evaluate(condition *ritual.QuestionCondition, answers map[string]interface{}) (bool, error) {
	if condition == nil {
		return true, nil
	}

	// Simple field equality check
	if condition.Field != "" && condition.Equals != nil {
		value, exists := answers[condition.Field]
		if !exists {
			return false, nil
		}
		return ce.compareValues(value, condition.Equals)
	}

	// Expression-based evaluation
	if condition.Expression != "" {
		return ce.evaluateExpression(condition.Expression, answers)
	}

	// AND logic
	if len(condition.And) > 0 {
		for _, subCond := range condition.And {
			result, err := ce.Evaluate(&subCond, answers)
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil
			}
		}
		return true, nil
	}

	// OR logic
	if len(condition.Or) > 0 {
		for _, subCond := range condition.Or {
			result, err := ce.Evaluate(&subCond, answers)
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil
	}

	// NOT logic
	if condition.Not != nil {
		result, err := ce.Evaluate(condition.Not, answers)
		if err != nil {
			return false, err
		}
		return !result, nil
	}

	return true, nil
}

func (ce *ConditionEvaluator) compareValues(actual, expected interface{}) (bool, error) {
	// Handle nil
	if actual == nil && expected == nil {
		return true, nil
	}
	if actual == nil || expected == nil {
		return false, nil
	}

	// Convert to strings for comparison
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)

	return actualStr == expectedStr, nil
}

// evaluateExpression evaluates a simple expression like "database_type == 'postgres'"
func (ce *ConditionEvaluator) evaluateExpression(expr string, answers map[string]interface{}) (bool, error) {
	expr = strings.TrimSpace(expr)

	// Handle equality: field == value
	if strings.Contains(expr, "==") {
		parts := strings.SplitN(expr, "==", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid expression: %s", expr)
		}

		field := strings.TrimSpace(parts[0])
		expectedVal := strings.Trim(strings.TrimSpace(parts[1]), "'\"")

		actualVal, exists := answers[field]
		if !exists {
			return false, nil
		}

		return ce.compareValues(actualVal, expectedVal)
	}

	// Handle inequality: field != value
	if strings.Contains(expr, "!=") {
		parts := strings.SplitN(expr, "!=", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid expression: %s", expr)
		}

		field := strings.TrimSpace(parts[0])
		expectedVal := strings.Trim(strings.TrimSpace(parts[1]), "'\"")

		actualVal, exists := answers[field]
		if !exists {
			return true, nil
		}

		result, err := ce.compareValues(actualVal, expectedVal)
		if err != nil {
			return false, err
		}
		return !result, nil
	}

	// Handle simple boolean field reference
	if matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, expr); matched {
		val, exists := answers[expr]
		if !exists {
			return false, nil
		}

		// Convert to boolean
		return ce.toBool(val), nil
	}

	// Handle AND operator
	if strings.Contains(expr, " && ") || strings.Contains(expr, " AND ") {
		sep := " && "
		if strings.Contains(expr, " AND ") {
			sep = " AND "
		}

		parts := strings.Split(expr, sep)
		for _, part := range parts {
			result, err := ce.evaluateExpression(part, answers)
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil
			}
		}
		return true, nil
	}

	// Handle OR operator
	if strings.Contains(expr, " || ") || strings.Contains(expr, " OR ") {
		sep := " || "
		if strings.Contains(expr, " OR ") {
			sep = " OR "
		}

		parts := strings.Split(expr, sep)
		for _, part := range parts {
			result, err := ce.evaluateExpression(part, answers)
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil
	}

	return false, fmt.Errorf("unsupported expression: %s", expr)
}

func (ce *ConditionEvaluator) toBool(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case string:
		v = strings.ToLower(v)
		return v == "true" || v == "yes" || v == "y" || v == "1"
	case int, int64:
		return v != 0
	case float64:
		return v != 0.0
	default:
		return false
	}
}

// EvaluateDefault evaluates a default value which may be dynamic
func (ce *ConditionEvaluator) EvaluateDefault(defaultVal interface{}, answers map[string]interface{}) interface{} {
	// If default is a string starting with $, it's a reference to another answer
	if strVal, ok := defaultVal.(string); ok {
		if strings.HasPrefix(strVal, "$") {
			fieldName := strings.TrimPrefix(strVal, "$")
			if val, exists := answers[fieldName]; exists {
				return val
			}
		}

		// Check for simple template-like substitution
		if strings.Contains(strVal, "{{") && strings.Contains(strVal, "}}") {
			return ce.substituteTemplateVars(strVal, answers)
		}
	}

	return defaultVal
}

func (ce *ConditionEvaluator) substituteTemplateVars(template string, answers map[string]interface{}) string {
	result := template

	// Find all {{var}} patterns
	re := regexp.MustCompile(`\{\{\.?([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			varName := match[1]
			if val, exists := answers[varName]; exists {
				placeholder := match[0]
				replacement := fmt.Sprintf("%v", val)
				result = strings.ReplaceAll(result, placeholder, replacement)
			}
		}
	}

	return result
}

// ConvertValue converts a string value to the appropriate type
func ConvertValue(valueStr string, targetType ritual.QuestionType) (interface{}, error) {
	switch targetType {
	case ritual.QuestionTypeBoolean:
		valueStr = strings.ToLower(strings.TrimSpace(valueStr))
		return valueStr == "true" || valueStr == "yes" || valueStr == "y" || valueStr == "1", nil

	case ritual.QuestionTypeNumber:
		// Try int first
		if intVal, err := strconv.Atoi(valueStr); err == nil {
			return intVal, nil
		}
		// Try float
		if floatVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
			return floatVal, nil
		}
		return nil, fmt.Errorf("invalid number: %s", valueStr)

	case ritual.QuestionTypeMultiChoice:
		// Split by comma
		parts := strings.Split(valueStr, ",")
		choices := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				choices = append(choices, trimmed)
			}
		}
		return choices, nil

	default:
		return valueStr, nil
	}
}

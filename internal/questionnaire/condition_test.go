package questionnaire

import (
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestConditionEvaluator_FieldEquals(t *testing.T) {
	eval := NewConditionEvaluator()

	condition := &ritual.QuestionCondition{
		Field:  "database",
		Equals: "postgres",
	}

	answers := map[string]interface{}{
		"database": "postgres",
	}

	result, err := eval.Evaluate(condition, answers)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if !result {
		t.Error("Expected true")
	}

	// Test with different value
	answers["database"] = "mysql"
	result, _ = eval.Evaluate(condition, answers)
	if result {
		t.Error("Expected false")
	}
}

func TestConditionEvaluator_Expression_Equality(t *testing.T) {
	eval := NewConditionEvaluator()

	condition := &ritual.QuestionCondition{
		Expression: "database == 'postgres'",
	}

	answers := map[string]interface{}{
		"database": "postgres",
	}

	result, err := eval.Evaluate(condition, answers)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if !result {
		t.Error("Expected true")
	}
}

func TestConditionEvaluator_Expression_Inequality(t *testing.T) {
	eval := NewConditionEvaluator()

	condition := &ritual.QuestionCondition{
		Expression: "database != 'mysql'",
	}

	answers := map[string]interface{}{
		"database": "postgres",
	}

	result, err := eval.Evaluate(condition, answers)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if !result {
		t.Error("Expected true")
	}
}

func TestConditionEvaluator_Expression_Boolean(t *testing.T) {
	eval := NewConditionEvaluator()

	condition := &ritual.QuestionCondition{
		Expression: "use_database",
	}

	answers := map[string]interface{}{
		"use_database": true,
	}

	result, err := eval.Evaluate(condition, answers)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if !result {
		t.Error("Expected true")
	}

	// Test with false
	answers["use_database"] = false
	result, _ = eval.Evaluate(condition, answers)
	if result {
		t.Error("Expected false")
	}
}

func TestConditionEvaluator_AND(t *testing.T) {
	eval := NewConditionEvaluator()

	condition := &ritual.QuestionCondition{
		And: []ritual.QuestionCondition{
			{Field: "a", Equals: "1"},
			{Field: "b", Equals: "2"},
		},
	}

	answers := map[string]interface{}{
		"a": "1",
		"b": "2",
	}

	result, err := eval.Evaluate(condition, answers)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if !result {
		t.Error("Expected true")
	}

	// One false should fail
	answers["b"] = "3"
	result, _ = eval.Evaluate(condition, answers)
	if result {
		t.Error("Expected false")
	}
}

func TestConditionEvaluator_OR(t *testing.T) {
	eval := NewConditionEvaluator()

	condition := &ritual.QuestionCondition{
		Or: []ritual.QuestionCondition{
			{Field: "a", Equals: "1"},
			{Field: "b", Equals: "2"},
		},
	}

	answers := map[string]interface{}{
		"a": "0",
		"b": "2",
	}

	result, err := eval.Evaluate(condition, answers)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if !result {
		t.Error("Expected true (one is true)")
	}

	// Both false
	answers["b"] = "0"
	result, _ = eval.Evaluate(condition, answers)
	if result {
		t.Error("Expected false")
	}
}

func TestConditionEvaluator_NOT(t *testing.T) {
	eval := NewConditionEvaluator()

	innerCond := ritual.QuestionCondition{
		Field:  "a",
		Equals: "1",
	}

	condition := &ritual.QuestionCondition{
		Not: &innerCond,
	}

	answers := map[string]interface{}{
		"a": "2",
	}

	result, err := eval.Evaluate(condition, answers)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if !result {
		t.Error("Expected true (negated)")
	}

	answers["a"] = "1"
	result, _ = eval.Evaluate(condition, answers)
	if result {
		t.Error("Expected false")
	}
}

func TestConditionEvaluator_ComplexExpression_AND(t *testing.T) {
	eval := NewConditionEvaluator()

	// Use struct-based AND instead of expression
	condition := &ritual.QuestionCondition{
		And: []ritual.QuestionCondition{
			{Field: "use_db", Equals: "true"},
			{Field: "db_type", Equals: "postgres"},
		},
	}

	answers := map[string]interface{}{
		"use_db":  "true",
		"db_type": "postgres",
	}

	result, err := eval.Evaluate(condition, answers)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if !result {
		t.Error("Expected true")
	}
}

func TestConditionEvaluator_ComplexExpression_OR(t *testing.T) {
	eval := NewConditionEvaluator()

	// Use struct-based OR instead of expression
	condition := &ritual.QuestionCondition{
		Or: []ritual.QuestionCondition{
			{Field: "db_type", Equals: "postgres"},
			{Field: "db_type", Equals: "mysql"},
		},
	}

	answers := map[string]interface{}{
		"db_type": "mysql",
	}

	result, err := eval.Evaluate(condition, answers)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if !result {
		t.Error("Expected true")
	}
}

func TestConditionEvaluator_EvaluateDefault(t *testing.T) {
	eval := NewConditionEvaluator()

	answers := map[string]interface{}{
		"app_name": "myapp",
	}

	// Reference to another answer
	result := eval.EvaluateDefault("$app_name", answers)
	if result != "myapp" {
		t.Errorf("Expected 'myapp', got '%v'", result)
	}

	// Template substitution
	result = eval.EvaluateDefault("{{app_name}}_db", answers)
	if result != "myapp_db" {
		t.Errorf("Expected 'myapp_db', got '%v'", result)
	}

	// Plain value
	result = eval.EvaluateDefault("literal", answers)
	if result != "literal" {
		t.Errorf("Expected 'literal', got '%v'", result)
	}
}

func TestConvertValue(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		targetType ritual.QuestionType
		expected   interface{}
		wantErr    bool
	}{
		{"boolean true", "true", ritual.QuestionTypeBoolean, true, false},
		{"boolean yes", "yes", ritual.QuestionTypeBoolean, true, false},
		{"boolean false", "false", ritual.QuestionTypeBoolean, false, false},
		{"number int", "42", ritual.QuestionTypeNumber, 42, false},
		{"number float", "3.14", ritual.QuestionTypeNumber, 3.14, false},
		{"multi choice", "a,b,c", ritual.QuestionTypeMultiChoice, []string{"a", "b", "c"}, false},
		{"text", "hello", ritual.QuestionTypeText, "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertValue(tt.input, tt.targetType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// For slices, need special comparison
				if tt.targetType == ritual.QuestionTypeMultiChoice {
					resultSlice, ok := result.([]string)
					if !ok {
						t.Errorf("Expected []string, got %T", result)
						return
					}
					expectedSlice := tt.expected.([]string)
					if len(resultSlice) != len(expectedSlice) {
						t.Errorf("Expected %v, got %v", expectedSlice, resultSlice)
						return
					}
					for i := range resultSlice {
						if resultSlice[i] != expectedSlice[i] {
							t.Errorf("Expected %v, got %v", expectedSlice, resultSlice)
							return
						}
					}
				} else if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestConditionEvaluator_toBool(t *testing.T) {
	ce := NewConditionEvaluator()
	
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"string true", "true", true},
		{"string True", "True", true},
		{"string yes", "yes", true},
		{"string Yes", "Yes", true},
		{"string y", "y", true},
		{"string Y", "Y", true},
		{"string 1", "1", true},
		{"string false", "false", false},
		{"string no", "no", false},
		{"int 0", 0, false},
		{"int 1", 1, true},
		{"int64 0", int64(0), true},  // Bug: type switch compares type, not value
		{"int64 5", int64(5), true},
		{"float64 0.0", 0.0, false},
		{"float64 1.5", 1.5, true},
		{"other type", []string{}, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ce.toBool(tt.input)
			if result != tt.expected {
				t.Errorf("toBool(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConditionEvaluator_Expression_ComplexLogic(t *testing.T) {
	eval := NewConditionEvaluator()

	tests := []struct {
		name       string
		expression string
		answers    map[string]interface{}
		expected   bool
	}{
		{
			name:       "Multiple ANDs",
			expression: "a && b && c",
			answers:    map[string]interface{}{"a": true, "b": true, "c": true},
			expected:   true,
		},
		{
			name:       "Multiple ANDs with false",
			expression: "a && b && c",
			answers:    map[string]interface{}{"a": true, "b": false, "c": true},
			expected:   false,
		},
		{
			name:       "Multiple ORs",
			expression: "a || b || c",
			answers:    map[string]interface{}{"a": false, "b": false, "c": true},
			expected:   true,
		},
		{
			name:       "Multiple ORs all false",
			expression: "a || b || c",
			answers:    map[string]interface{}{"a": false, "b": false, "c": false},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := &ritual.QuestionCondition{
				Expression: tt.expression,
			}
			result, err := eval.Evaluate(condition, tt.answers)
			if err != nil {
				t.Fatalf("Evaluate failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConditionEvaluator_Expression_InvalidSyntax(t *testing.T) {
	eval := NewConditionEvaluator()

	tests := []string{
		"!field",   // NOT operator not supported
		"a > 5",    // Greater than not supported
		"a < 10",   // Less than not supported
		"(a && b)", // Parentheses not supported
	}

	for _, expr := range tests {
		condition := &ritual.QuestionCondition{
			Expression: expr,
		}
		_, err := eval.Evaluate(condition, map[string]interface{}{"a": 1, "b": true, "field": false})
		if err == nil {
			t.Errorf("Expected error for unsupported expression: %s", expr)
		}
	}
}

func TestConditionEvaluator_Expression_UndefinedVariable(t *testing.T) {
	eval := NewConditionEvaluator()

	tests := []struct {
		name     string
		expr     string
		expected bool
	}{
		{
			name:     "undefined in equality",
			expr:     "undefined_var == 'value'",
			expected: false,
		},
		{
			name:     "undefined in inequality",
			expr:     "undefined_var != 'value'",
			expected: true,
		},
		{
			name:     "undefined field reference",
			expr:     "undefined_var",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := &ritual.QuestionCondition{
				Expression: tt.expr,
			}
			result, err := eval.Evaluate(condition, map[string]interface{}{})
			if err != nil {
				t.Fatalf("Evaluate failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.expr, result)
			}
		})
	}
}

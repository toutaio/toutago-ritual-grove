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

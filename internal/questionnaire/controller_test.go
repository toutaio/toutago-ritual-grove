package questionnaire

import (
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestController_BasicFlow(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:     "name",
			Prompt:   "What is your name?",
			Type:     ritual.QuestionTypeText,
			Required: true,
		},
		{
			Name:   "age",
			Prompt: "What is your age?",
			Type:   ritual.QuestionTypeNumber,
		},
	}

	ctrl := NewController(questions)

	// Get first question
	q, err := ctrl.GetNextQuestion()
	if err != nil {
		t.Fatalf("GetNextQuestion failed: %v", err)
	}
	if q == nil {
		t.Fatal("Expected question, got nil")
	}
	if q.Name != "name" {
		t.Errorf("Expected 'name', got '%s'", q.Name)
	}

	// Submit answer
	err = ctrl.SubmitAnswer("name", "John")
	if err != nil {
		t.Fatalf("SubmitAnswer failed: %v", err)
	}

	// Get second question
	q, err = ctrl.GetNextQuestion()
	if err != nil {
		t.Fatalf("GetNextQuestion failed: %v", err)
	}
	if q.Name != "age" {
		t.Errorf("Expected 'age', got '%s'", q.Name)
	}

	// Submit answer
	err = ctrl.SubmitAnswer("age", 25)
	if err != nil {
		t.Fatalf("SubmitAnswer failed: %v", err)
	}

	// Should be complete
	if !ctrl.IsComplete() {
		t.Error("Expected controller to be complete")
	}

	// Check answers
	answers := ctrl.GetAnswers()
	if answers["name"] != "John" {
		t.Errorf("Expected 'John', got '%v'", answers["name"])
	}
	if answers["age"] != 25 {
		t.Errorf("Expected 25, got '%v'", answers["age"])
	}
}

func TestController_RequiredValidation(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:     "required_field",
			Prompt:   "Required field",
			Type:     ritual.QuestionTypeText,
			Required: true,
		},
	}

	ctrl := NewController(questions)

	// Try to submit empty answer
	err := ctrl.SubmitAnswer("required_field", "")
	if err == nil {
		t.Error("Expected validation error for required field")
	}

	// Submit valid answer
	err = ctrl.SubmitAnswer("required_field", "value")
	if err != nil {
		t.Fatalf("SubmitAnswer failed: %v", err)
	}

	if !ctrl.IsComplete() {
		t.Error("Expected controller to be complete")
	}
}

func TestController_ConditionalQuestions(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:   "use_database",
			Prompt: "Use database?",
			Type:   ritual.QuestionTypeBoolean,
		},
		{
			Name:   "database_type",
			Prompt: "Database type?",
			Type:   ritual.QuestionTypeChoice,
			Choices: []string{"postgres", "mysql"},
			Condition: &ritual.QuestionCondition{
				Field:  "use_database",
				Equals: true,
			},
		},
	}

	ctrl := NewController(questions)

	// First question
	q, _ := ctrl.GetNextQuestion()
	if q.Name != "use_database" {
		t.Errorf("Expected 'use_database', got '%s'", q.Name)
	}

	// Answer false - should skip database_type
	ctrl.SubmitAnswer("use_database", false)

	// Next question should be nil (no more questions)
	q, _ = ctrl.GetNextQuestion()
	if q != nil {
		t.Errorf("Expected no more questions, got '%s'", q.Name)
	}

	// Reset and try with true
	ctrl.Reset()

	q, _ = ctrl.GetNextQuestion()
	ctrl.SubmitAnswer("use_database", true)

	// Should now ask for database_type
	q, _ = ctrl.GetNextQuestion()
	if q == nil {
		t.Fatal("Expected database_type question")
	}
	if q.Name != "database_type" {
		t.Errorf("Expected 'database_type', got '%s'", q.Name)
	}
}

func TestController_Progress(t *testing.T) {
	questions := []ritual.Question{
		{Name: "q1", Prompt: "Q1", Type: ritual.QuestionTypeText},
		{Name: "q2", Prompt: "Q2", Type: ritual.QuestionTypeText},
		{Name: "q3", Prompt: "Q3", Type: ritual.QuestionTypeText},
	}

	ctrl := NewController(questions)

	answered, total := ctrl.GetProgress()
	if answered != 0 || total != 3 {
		t.Errorf("Expected 0/3, got %d/%d", answered, total)
	}

	ctrl.GetNextQuestion()
	ctrl.SubmitAnswer("q1", "a1")

	answered, total = ctrl.GetProgress()
	if answered != 1 || total != 3 {
		t.Errorf("Expected 1/3, got %d/%d", answered, total)
	}

	ctrl.GetNextQuestion()
	ctrl.SubmitAnswer("q2", "a2")

	answered, total = ctrl.GetProgress()
	if answered != 2 || total != 3 {
		t.Errorf("Expected 2/3, got %d/%d", answered, total)
	}
}

func TestQuestionFlow_StateManagement(t *testing.T) {
	flow := NewQuestionFlow()

	flow.AddQuestion("q1")
	flow.AddQuestion("q2")

	// Initial state
	if flow.GetState("q1") != StateNotReached {
		t.Error("Expected StateNotReached")
	}

	// Set active
	flow.SetState("q1", StateActive)
	if flow.GetState("q1") != StateActive {
		t.Error("Expected StateActive")
	}

	// Set answer
	flow.SetAnswer("q1", "answer")
	if flow.GetState("q1") != StateAnswered {
		t.Error("Expected StateAnswered")
	}

	// Get answer
	answer, exists := flow.GetAnswer("q1")
	if !exists || answer != "answer" {
		t.Errorf("Expected 'answer', got '%v'", answer)
	}
}

func TestQuestionFlow_AnswerValidation(t *testing.T) {
	flow := NewQuestionFlow()
	flow.AddQuestion("q1")

	// Set invalid answer
	flow.SetAnswerWithError("q1", "bad", "validation error")

	answerVal, exists := flow.GetAnswerValue("q1")
	if !exists {
		t.Fatal("Expected answer to exist")
	}

	if answerVal.IsValid {
		t.Error("Expected answer to be invalid")
	}

	if answerVal.ErrorMsg != "validation error" {
		t.Errorf("Expected 'validation error', got '%s'", answerVal.ErrorMsg)
	}

	// Invalid answer should not be in AllAnswers
	all := flow.AllAnswers()
	if _, exists := all["q1"]; exists {
		t.Error("Invalid answer should not be in AllAnswers")
	}
}

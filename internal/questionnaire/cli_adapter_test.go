package questionnaire

import (
	"strings"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestCLIAdapter_SimpleTextQuestion(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:   "name",
			Prompt: "What is your name?",
			Type:   ritual.QuestionTypeText,
		},
	}

	input := "John Doe\n"
	adapter := NewCLIAdapter(questions, strings.NewReader(input))
	
	answers, err := adapter.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if answers["name"] != "John Doe" {
		t.Errorf("Expected name to be 'John Doe', got %v", answers["name"])
	}
}

func TestCLIAdapter_RequiredQuestion(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:     "email",
			Prompt:   "Email:",
			Type:     ritual.QuestionTypeEmail,
			Required: true,
		},
	}

	// Empty input should fail
	input := "\ntest@example.com\n"
	adapter := NewCLIAdapter(questions, strings.NewReader(input))
	
	answers, err := adapter.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if answers["email"] != "test@example.com" {
		t.Errorf("Expected email to be 'test@example.com', got %v", answers["email"])
	}
}

func TestCLIAdapter_ChoiceQuestion(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:    "color",
			Prompt:  "Choose a color:",
			Type:    ritual.QuestionTypeChoice,
			Choices: []string{"red", "green", "blue"},
		},
	}

	input := "green\n"
	adapter := NewCLIAdapter(questions, strings.NewReader(input))
	
	answers, err := adapter.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if answers["color"] != "green" {
		t.Errorf("Expected color to be 'green', got %v", answers["color"])
	}
}

func TestCLIAdapter_BooleanQuestion(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:   "confirm",
			Prompt: "Continue?",
			Type:   ritual.QuestionTypeBoolean,
		},
	}

	input := "yes\n"
	adapter := NewCLIAdapter(questions, strings.NewReader(input))
	
	answers, err := adapter.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if answers["confirm"] != true {
		t.Errorf("Expected confirm to be true, got %v", answers["confirm"])
	}
}

func TestCLIAdapter_NumberQuestion(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:   "port",
			Prompt: "Port number:",
			Type:   ritual.QuestionTypeNumber,
		},
	}

	input := "8080\n"
	adapter := NewCLIAdapter(questions, strings.NewReader(input))
	
	answers, err := adapter.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if answers["port"] != 8080 {
		t.Errorf("Expected port to be 8080, got %v", answers["port"])
	}
}

func TestCLIAdapter_DefaultValue(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:    "host",
			Prompt:  "Host:",
			Type:    ritual.QuestionTypeText,
			Default: "localhost",
		},
	}

	input := "\n" // Just press enter to use default
	adapter := NewCLIAdapter(questions, strings.NewReader(input))
	
	answers, err := adapter.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if answers["host"] != "localhost" {
		t.Errorf("Expected host to be 'localhost', got %v", answers["host"])
	}
}

func TestCLIAdapter_ConditionalQuestion(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:   "use_db",
			Prompt: "Use database?",
			Type:   ritual.QuestionTypeBoolean,
		},
		{
			Name:   "db_type",
			Prompt: "Database type:",
			Type:   ritual.QuestionTypeChoice,
			Choices: []string{"postgres", "mysql"},
			Condition: &ritual.QuestionCondition{
				Field:  "use_db",
				Equals: true,
			},
		},
	}

	// Answer yes to database, choose postgres
	input := "yes\npostgres\n"
	adapter := NewCLIAdapter(questions, strings.NewReader(input))
	
	answers, err := adapter.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if answers["use_db"] != true {
		t.Errorf("Expected use_db to be true, got %v", answers["use_db"])
	}

	if answers["db_type"] != "postgres" {
		t.Errorf("Expected db_type to be 'postgres', got %v", answers["db_type"])
	}
}

func TestCLIAdapter_SkipConditionalQuestion(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:   "use_db",
			Prompt: "Use database?",
			Type:   ritual.QuestionTypeBoolean,
		},
		{
			Name:   "db_type",
			Prompt: "Database type:",
			Type:   ritual.QuestionTypeChoice,
			Choices: []string{"postgres", "mysql"},
			Condition: &ritual.QuestionCondition{
				Field:  "use_db",
				Equals: true,
			},
		},
	}

	// Answer no to database - db_type should be skipped
	input := "no\n"
	adapter := NewCLIAdapter(questions, strings.NewReader(input))
	
	answers, err := adapter.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if answers["use_db"] != false {
		t.Errorf("Expected use_db to be false, got %v", answers["use_db"])
	}

	if _, exists := answers["db_type"]; exists {
		t.Errorf("Expected db_type to not exist in answers")
	}
}

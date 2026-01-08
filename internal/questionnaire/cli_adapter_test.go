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
			Name:    "db_type",
			Prompt:  "Database type:",
			Type:    ritual.QuestionTypeChoice,
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
			Name:    "db_type",
			Prompt:  "Database type:",
			Type:    ritual.QuestionTypeChoice,
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

func TestCLIAdapter_SetWriter(t *testing.T) {
	questions := []ritual.Question{
		{
			Name:   "test",
			Prompt: "Test:",
			Type:   ritual.QuestionTypeText,
		},
	}

	input := "value\n"
	adapter := NewCLIAdapter(questions, strings.NewReader(input))

	var buf strings.Builder
	adapter.SetWriter(&buf)

	_, err := adapter.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test:") {
		t.Errorf("Expected output to contain prompt, got: %s", output)
	}
}

func TestConvertAnswer_EdgeCases(t *testing.T) {
	adapter := &CLIAdapter{}

	tests := []struct {
		name      string
		question  *ritual.Question
		value     string
		expected  interface{}
		wantError bool
	}{
		{"text type", &ritual.Question{Type: ritual.QuestionTypeText}, "hello", "hello", false},
		{"password type", &ritual.Question{Type: ritual.QuestionTypePassword}, "secret", "secret", false},
		{"path type", &ritual.Question{Type: ritual.QuestionTypePath}, "/home/user", "/home/user", false},
		{"url type", &ritual.Question{Type: ritual.QuestionTypeURL}, "https://example.com", "https://example.com", false},
		{"email type", &ritual.Question{Type: ritual.QuestionTypeEmail}, "test@test.com", "test@test.com", false},
		{"choice type", &ritual.Question{Type: ritual.QuestionTypeChoice, Choices: []string{"option1"}}, "option1", "option1", false},
		{"multi-choice type", &ritual.Question{Type: ritual.QuestionTypeMultiChoice}, "a,b,c", []string{"a", "b", "c"}, false},
		{"invalid number", &ritual.Question{Type: ritual.QuestionTypeNumber}, "abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.convertAnswer(tt.question, tt.value)
			if tt.wantError {
				if err == nil {
					t.Errorf("convertAnswer() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("convertAnswer() unexpected error: %v", err)
				return
			}

			// Special handling for slices
			if tt.question.Type == ritual.QuestionTypeMultiChoice {
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
		})
	}
}

func TestStripQuotes(t *testing.T) {
tests := []struct {
name     string
input    string
expected string
}{
{
name:     "double quotes",
input:    `"inertia-vue"`,
expected: "inertia-vue",
},
{
name:     "single quotes",
input:    "'inertia-vue'",
expected: "inertia-vue",
},
{
name:     "no quotes",
input:    "inertia-vue",
expected: "inertia-vue",
},
{
name:     "mixed quotes",
input:    `"inertia-vue'`,
expected: `"inertia-vue'`,
},
{
name:     "empty string",
input:    "",
expected: "",
},
{
name:     "only quotes",
input:    `""`,
expected: "",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
result := stripQuotes(tt.input)
if result != tt.expected {
t.Errorf("stripQuotes(%q) = %q, want %q", tt.input, result, tt.expected)
}
})
}
}

func TestCLIAdapter_ChoiceQuestionWithQuotes(t *testing.T) {
questions := []ritual.Question{
{
Name:   "frontend",
Prompt: "Select frontend:",
Type:   ritual.QuestionTypeChoice,
Choices: []string{"traditional", "inertia-vue", "htmx"},
},
}

// Test with quoted input
input := `"inertia-vue"` + "\n"
adapter := NewCLIAdapter(questions, strings.NewReader(input))

answers, err := adapter.Run()
if err != nil {
t.Fatalf("Run() failed: %v", err)
}

if answers["frontend"] != "inertia-vue" {
t.Errorf("Expected frontend to be 'inertia-vue', got %v", answers["frontend"])
}
}

func TestCLIAdapter_ChoiceQuestionWithNumber(t *testing.T) {
questions := []ritual.Question{
{
Name:   "frontend",
Prompt: "Select frontend:",
Type:   ritual.QuestionTypeChoice,
Choices: []string{"traditional", "inertia-vue", "htmx"},
},
}

// Test with number input (1-based)
input := "2\n"
adapter := NewCLIAdapter(questions, strings.NewReader(input))

answers, err := adapter.Run()
if err != nil {
t.Fatalf("Run() failed: %v", err)
}

if answers["frontend"] != "inertia-vue" {
t.Errorf("Expected frontend to be 'inertia-vue' (choice #2), got %v", answers["frontend"])
}
}

func TestCLIAdapter_ChoiceQuestionInvalidNumber(t *testing.T) {
questions := []ritual.Question{
{
Name:   "frontend",
Prompt: "Select frontend:",
Type:   ritual.QuestionTypeChoice,
Choices: []string{"traditional", "inertia-vue", "htmx"},
},
}

// Test with invalid number, then valid choice
input := "5\nhtmx\n"
adapter := NewCLIAdapter(questions, strings.NewReader(input))

answers, err := adapter.Run()
if err != nil {
t.Fatalf("Run() failed: %v", err)
}

if answers["frontend"] != "htmx" {
t.Errorf("Expected frontend to be 'htmx', got %v", answers["frontend"])
}
}

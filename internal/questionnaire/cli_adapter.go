package questionnaire

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// CLIAdapter provides a command-line interface for questionnaires
type CLIAdapter struct {
	controller *Controller
	reader     io.Reader
	writer     io.Writer
	scanner    *bufio.Scanner
}

// NewCLIAdapter creates a new CLI adapter
func NewCLIAdapter(questions []ritual.Question, reader io.Reader) *CLIAdapter {
	if reader == nil {
		reader = os.Stdin
	}
	
	return &CLIAdapter{
		controller: NewController(questions),
		reader:     reader,
		writer:     os.Stdout,
		scanner:    bufio.NewScanner(reader),
	}
}

// SetWriter sets the output writer (for testing)
func (a *CLIAdapter) SetWriter(w io.Writer) {
	a.writer = w
}

// Run executes the questionnaire and returns collected answers
func (a *CLIAdapter) Run() (map[string]interface{}, error) {
	for {
		question, err := a.controller.GetNextQuestion()
		if err != nil {
			return nil, fmt.Errorf("failed to get next question: %w", err)
		}
		
		if question == nil {
			// No more questions
			break
		}

		// Ask the question
		answer, err := a.askQuestion(question)
		if err != nil {
			return nil, fmt.Errorf("failed to ask question %s: %w", question.Name, err)
		}

		// Submit the answer
		if err := a.controller.SubmitAnswer(question.Name, answer); err != nil {
			// Show error and retry
			_, _ = fmt.Fprintf(a.writer, "Error: %v\n", err)
			continue
		}
	}

	return a.controller.GetAnswers(), nil
}

// askQuestion prompts the user and reads their answer
func (a *CLIAdapter) askQuestion(q *ritual.Question) (interface{}, error) {
	// Display prompt
	prompt := q.Prompt
	if q.Default != nil {
		prompt = fmt.Sprintf("%s [%v]", prompt, q.Default)
	}
	_, _ = fmt.Fprintf(a.writer, "%s: ", prompt)

	// Read input
	if !a.scanner.Scan() {
		if err := a.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	input := strings.TrimSpace(a.scanner.Text())

	// Use default if input is empty
	if input == "" && q.Default != nil {
		return q.Default, nil
	}

	// Convert based on question type
	return a.convertAnswer(q, input)
}

// convertAnswer converts string input to the appropriate type
func (a *CLIAdapter) convertAnswer(q *ritual.Question, input string) (interface{}, error) {
	switch q.Type {
	case ritual.QuestionTypeText, ritual.QuestionTypePassword, 
	     ritual.QuestionTypePath, ritual.QuestionTypeURL, ritual.QuestionTypeEmail:
		return input, nil

	case ritual.QuestionTypeBoolean:
		return a.parseBoolean(input)

	case ritual.QuestionTypeNumber:
		return a.parseNumber(input)

	case ritual.QuestionTypeChoice:
		// Validate choice
		if len(q.Choices) > 0 {
			for _, choice := range q.Choices {
				if input == choice {
					return input, nil
				}
			}
			return nil, fmt.Errorf("invalid choice: %s (valid: %v)", input, q.Choices)
		}
		return input, nil

	case ritual.QuestionTypeMultiChoice:
		// Split by comma and trim
		parts := strings.Split(input, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result, nil

	default:
		return input, nil
	}
}

// parseBoolean parses boolean input
func (a *CLIAdapter) parseBoolean(input string) (bool, error) {
	lower := strings.ToLower(input)
	switch lower {
	case "yes", "y", "true", "t", "1":
		return true, nil
	case "no", "n", "false", "f", "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s (expected yes/no)", input)
	}
}

// parseNumber parses numeric input
func (a *CLIAdapter) parseNumber(input string) (int, error) {
	val, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", input)
	}
	return val, nil
}

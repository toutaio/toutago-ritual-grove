package questionnaire

import (
	"fmt"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// QuestionState represents the state of a question in the flow
type QuestionState int

const (
	StateNotReached QuestionState = iota
	StateActive
	StateAnswered
	StateSkipped
)

// AnswerValue holds the answer to a question
type AnswerValue struct {
	Value    interface{}
	IsValid  bool
	ErrorMsg string
}

// QuestionFlow manages question state and navigation
type QuestionFlow struct {
	states  map[string]QuestionState
	answers map[string]AnswerValue
	order   []string
}

// NewQuestionFlow creates a new question flow
func NewQuestionFlow() *QuestionFlow {
	return &QuestionFlow{
		states:  make(map[string]QuestionState),
		answers: make(map[string]AnswerValue),
		order:   make([]string, 0),
	}
}

// AddQuestion adds a question to the flow
func (qf *QuestionFlow) AddQuestion(name string) {
	qf.states[name] = StateNotReached
	qf.order = append(qf.order, name)
}

// SetState sets the state of a question
func (qf *QuestionFlow) SetState(name string, state QuestionState) {
	qf.states[name] = state
}

// GetState gets the state of a question
func (qf *QuestionFlow) GetState(name string) QuestionState {
	if state, exists := qf.states[name]; exists {
		return state
	}
	return StateNotReached
}

// SetAnswer stores an answer
func (qf *QuestionFlow) SetAnswer(name string, value interface{}) {
	qf.answers[name] = AnswerValue{
		Value:   value,
		IsValid: true,
	}
	qf.states[name] = StateAnswered
}

// SetAnswerWithError stores an answer with validation error
func (qf *QuestionFlow) SetAnswerWithError(name string, value interface{}, err string) {
	qf.answers[name] = AnswerValue{
		Value:    value,
		IsValid:  false,
		ErrorMsg: err,
	}
}

// GetAnswer retrieves an answer
func (qf *QuestionFlow) GetAnswer(name string) (interface{}, bool) {
	if answer, exists := qf.answers[name]; exists && answer.IsValid {
		return answer.Value, true
	}
	return nil, false
}

// GetAnswerValue retrieves the full answer value
func (qf *QuestionFlow) GetAnswerValue(name string) (AnswerValue, bool) {
	answer, exists := qf.answers[name]
	return answer, exists
}

// AllAnswers returns all valid answers
func (qf *QuestionFlow) AllAnswers() map[string]interface{} {
	result := make(map[string]interface{})
	for name, answer := range qf.answers {
		if answer.IsValid {
			result[name] = answer.Value
		}
	}
	return result
}

// Controller manages the questionnaire flow
type Controller struct {
	questions     []ritual.Question
	flow          *QuestionFlow
	condEvaluator *ConditionEvaluator
	validator     *Validator
}

// NewController creates a new questionnaire controller
func NewController(questions []ritual.Question) *Controller {
	flow := NewQuestionFlow()
	for _, q := range questions {
		flow.AddQuestion(q.Name)
	}

	return &Controller{
		questions:     questions,
		flow:          flow,
		condEvaluator: NewConditionEvaluator(),
		validator:     NewValidator(),
	}
}

// GetNextQuestion returns the next question to ask
func (c *Controller) GetNextQuestion() (*ritual.Question, error) {
	for _, q := range c.questions {
		state := c.flow.GetState(q.Name)

		// Skip already answered questions
		if state == StateAnswered {
			continue
		}

		// Check if condition is met
		if q.Condition != nil {
			// print for debug
			fmt.Printf("Evaluating condition for question %s with answers: %+v\n", q.Name, c.flow.AllAnswers())
			fmt.Printf("Condition: %+v\n", q.Condition)
			shouldShow, err := c.condEvaluator.Evaluate(q.Condition, c.flow.AllAnswers())
			if err != nil {
				return nil, fmt.Errorf("failed A to evaluate condition for %s: %w", q.Name, err)
			}
			if !shouldShow {
				c.flow.SetState(q.Name, StateSkipped)
				continue
			}
		}

		// This is the next question
		c.flow.SetState(q.Name, StateActive)
		return &q, nil
	}

	return nil, nil // No more questions
}

// SubmitAnswer submits an answer to the current question
func (c *Controller) SubmitAnswer(questionName string, value interface{}) error {
	// Find the question
	var question *ritual.Question
	for _, q := range c.questions {
		if q.Name == questionName {
			question = &q
			break
		}
	}

	if question == nil {
		return fmt.Errorf("question not found: %s", questionName)
	}

	// Validate the answer
	if err := c.validator.ValidateAnswer(question, value); err != nil {
		c.flow.SetAnswerWithError(questionName, value, err.Error())
		return err
	}

	// Store the answer
	c.flow.SetAnswer(questionName, value)

	return nil
}

// GetProgress returns the current progress (answered/total)
func (c *Controller) GetProgress() (answered int, total int) {
	total = len(c.questions)
	for _, q := range c.questions {
		state := c.flow.GetState(q.Name)
		if state == StateAnswered {
			answered++
		}
	}
	return answered, total
}

// IsComplete checks if all required questions are answered
func (c *Controller) IsComplete() bool {
	for _, q := range c.questions {
		state := c.flow.GetState(q.Name)

		// Skip questions with unmet conditions
		if q.Condition != nil {
			shouldShow, _ := c.condEvaluator.Evaluate(q.Condition, c.flow.AllAnswers())
			if !shouldShow {
				continue
			}
		}

		// Check required questions
		if q.Required && state != StateAnswered {
			return false
		}
	}
	return true
}

// GetAnswers returns all collected answers
func (c *Controller) GetAnswers() map[string]interface{} {
	return c.flow.AllAnswers()
}

// Reset resets the controller state
func (c *Controller) Reset() {
	c.flow = NewQuestionFlow()
	for _, q := range c.questions {
		c.flow.AddQuestion(q.Name)
	}
}

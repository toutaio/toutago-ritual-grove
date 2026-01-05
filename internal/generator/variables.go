package generator

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Variables manages template variables and substitutions
type Variables struct {
	data map[string]interface{}
}

// NewVariables creates a new variables manager
func NewVariables() *Variables {
	return &Variables{
		data: make(map[string]interface{}),
	}
}

// Set sets a variable value
func (v *Variables) Set(key string, value interface{}) {
	v.data[key] = value
}

// Get gets a variable value
func (v *Variables) Get(key string) (interface{}, bool) {
	val, ok := v.data[key]
	return val, ok
}

// GetString gets a variable as string
func (v *Variables) GetString(key string) string {
	if val, ok := v.data[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// GetBool gets a variable as bool
func (v *Variables) GetBool(key string) bool {
	if val, ok := v.data[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return v == "true" || v == "1" || v == "yes"
		case int:
			return v != 0
		}
	}
	return false
}

// SetFromAnswers loads variables from questionnaire answers
func (v *Variables) SetFromAnswers(answers map[string]interface{}) {
	for key, value := range answers {
		v.data[key] = value
	}
}

// SetFromEnvironment loads variables from environment
func (v *Variables) SetFromEnvironment(prefix string) {
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Only load if matches prefix
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			continue
		}

		// Remove prefix if specified
		if prefix != "" {
			key = strings.TrimPrefix(key, prefix)
		}

		// Convert to lowercase for consistency
		key = strings.ToLower(key)

		v.data[key] = value
	}
}

// AddComputed adds computed variables
func (v *Variables) AddComputed() {
	// Add timestamp
	v.data["now"] = time.Now().Format(time.RFC3339)
	v.data["timestamp"] = time.Now().Unix()
	v.data["year"] = time.Now().Year()

	// Add common transformations for each variable
	for key, value := range v.data {
		strValue := fmt.Sprintf("%v", value)

		// Add case transformations
		caser := cases.Title(language.English)
		v.data[key+"_upper"] = strings.ToUpper(strValue)
		v.data[key+"_lower"] = strings.ToLower(strValue)
		v.data[key+"_title"] = caser.String(strValue)
		v.data[key+"_pascal"] = toPascalCase(strValue)
		v.data[key+"_camel"] = toCamelCase(strValue)
		v.data[key+"_snake"] = toSnakeCase(strValue)
		v.data[key+"_kebab"] = toKebabCase(strValue)
	}
}

// All returns all variables
func (v *Variables) All() map[string]interface{} {
	result := make(map[string]interface{})
	for k, val := range v.data {
		result[k] = val
	}
	return result
}

// MaskSecrets returns a copy with secrets masked for logging
func (v *Variables) MaskSecrets(secretKeys []string) map[string]interface{} {
	result := make(map[string]interface{})
	secretSet := make(map[string]bool)
	for _, key := range secretKeys {
		secretSet[key] = true
	}

	for k, val := range v.data {
		if secretSet[k] || strings.Contains(strings.ToLower(k), "password") ||
			strings.Contains(strings.ToLower(k), "secret") ||
			strings.Contains(strings.ToLower(k), "token") {
			result[k] = "***"
		} else {
			result[k] = val
		}
	}
	return result
}

// Case conversion helpers

func toPascalCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[0:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) > 0 {
		return strings.ToLower(pascal[0:1]) + pascal[1:]
	}
	return pascal
}

func toSnakeCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "_")
}

func toKebabCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "-")
}

func splitWords(s string) []string {
	// Split on common delimiters
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")

	// Split on uppercase letters (camelCase/PascalCase)
	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
		if r != ' ' {
			currentWord.WriteRune(r)
		} else if currentWord.Len() > 0 {
			words = append(words, currentWord.String())
			currentWord.Reset()
		}
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

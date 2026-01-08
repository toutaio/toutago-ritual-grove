package generator

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TemplateEngine defines the interface for template rendering
type TemplateEngine interface {
	Render(templateContent string, data map[string]interface{}) (string, error)
	RenderFile(templatePath string, data map[string]interface{}) (string, error)
}

// GoTemplateEngine implements TemplateEngine using Go's text/template
type GoTemplateEngine struct {
	funcMap   template.FuncMap
	leftDelim  string
	rightDelim string
}

// NewGoTemplateEngine creates a new Go template engine with default delimiters
func NewGoTemplateEngine() *GoTemplateEngine {
	return NewGoTemplateEngineWithDelimiters("[[", "]]")
}

// NewGoTemplateEngineWithDelimiters creates a new Go template engine with custom delimiters
func NewGoTemplateEngineWithDelimiters(left, right string) *GoTemplateEngine {
	caser := cases.Title(language.English)
	funcMap := template.FuncMap{
		"upper":  strings.ToUpper,
		"lower":  strings.ToLower,
		"title":  caser.String,
		"pascal": toPascalCase,
		"camel":  toCamelCase,
		"snake":  toSnakeCase,
		"kebab":  toKebabCase,
	}

	return &GoTemplateEngine{
		funcMap:    funcMap,
		leftDelim:  left,
		rightDelim: right,
	}
}

// Render renders a template string with data
func (e *GoTemplateEngine) Render(templateContent string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("template").
		Delims(e.leftDelim, e.rightDelim).
		Funcs(e.funcMap).
		Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// RenderFile renders a template file with data
func (e *GoTemplateEngine) RenderFile(templatePath string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New(filepath.Base(templatePath)).
		Delims(e.leftDelim, e.rightDelim).
		Funcs(e.funcMap).
		ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template file: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// FithTemplateEngine is a placeholder for Fíth integration
// TODO: Integrate with toutago-fith-renderer when available
type FithTemplateEngine struct {
	// Will be implemented when integrating with Fíth
}

// NewFithTemplateEngine creates a new Fíth template engine
func NewFithTemplateEngine() *FithTemplateEngine {
	return &FithTemplateEngine{}
}

// Render renders a template string with data (placeholder)
func (e *FithTemplateEngine) Render(templateContent string, data map[string]interface{}) (string, error) {
	// TODO: Integrate with toutago-fith-renderer
	// For now, fallback to Go templates
	goEngine := NewGoTemplateEngine()
	return goEngine.Render(templateContent, data)
}

// RenderFile renders a template file with data (placeholder)
func (e *FithTemplateEngine) RenderFile(templatePath string, data map[string]interface{}) (string, error) {
	// TODO: Integrate with toutago-fith-renderer
	// For now, fallback to Go templates
	goEngine := NewGoTemplateEngine()
	return goEngine.RenderFile(templatePath, data)
}

// NewTemplateEngine creates a template engine based on the specified type
func NewTemplateEngine(engineType string) TemplateEngine {
	switch engineType {
	case "go-template":
		return NewGoTemplateEngine()
	case "fith":
		return NewFithTemplateEngine()
	default:
		// Default to Fíth
		return NewFithTemplateEngine()
	}
}

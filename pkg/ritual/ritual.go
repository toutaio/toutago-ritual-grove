// Package ritual provides the public API for the Ritual Grove system.
//
// Ritual Grove is a powerful application recipe system for creating, managing,
// and deploying complete applications from templates (rituals).
package ritual

// Version is the current version of Ritual Grove
const Version = "0.2.0"

// Ritual represents a complete application recipe/template (for backward compatibility)
// Use Manifest for the full ritual definition
type Ritual struct {
	// Metadata
	Name        string
	Version     string
	Description string
	Author      string

	// Configuration
	TemplateEngine  string
	MinToutaVersion string
	MaxToutaVersion string

	// Components
	Questions []Question
	Templates []Template
	Packages  []string
	Mixins    []Mixin

	// Lifecycle
	Hooks Hooks
}

// Template represents a file template to be generated (for backward compatibility)
type Template struct {
	Source      string
	Destination string
	Protected   bool // preserve user modifications
}

// Mixin represents an optional composable feature
type Mixin struct {
	Name        string
	Description string
	Templates   []Template
	Packages    []string
}

// Hooks contains lifecycle hooks for rituals
type Hooks struct {
	PreInstall  []string
	PostInstall []string
	PreUpdate   []string
	PostUpdate  []string
	PreDeploy   []string
	PostDeploy  []string
}

// Registry manages ritual discovery and loading
type Registry interface {
	// List returns all available rituals
	List() ([]Ritual, error)

	// Get retrieves a specific ritual by name
	Get(name string) (*Ritual, error)

	// Install downloads and installs a ritual from a source
	Install(source string) error
}

// Generator handles code generation from templates
type Generator interface {
	// Generate creates files from ritual templates
	Generate(ritual *Ritual, answers map[string]interface{}) error

	// ApplyMixin adds a mixin to an existing project
	ApplyMixin(mixin *Mixin, answers map[string]interface{}) error
}

// TemplateEngine renders templates with variables
type TemplateEngine interface {
	// Render renders a template string with data
	Render(template string, data map[string]interface{}) (string, error)

	// RenderFile renders a template file with data
	RenderFile(path string, data map[string]interface{}) (string, error)
}

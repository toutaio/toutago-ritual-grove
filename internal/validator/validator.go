package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// Validator validates ritual manifests
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a ritual manifest
func (v *Validator) Validate(manifest *ritual.Manifest) error {
	if err := v.validateMetadata(manifest); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	if err := v.validateCompatibility(manifest); err != nil {
		return fmt.Errorf("compatibility validation failed: %w", err)
	}

	if err := v.validateQuestions(manifest); err != nil {
		return fmt.Errorf("questions validation failed: %w", err)
	}

	if err := v.validateFiles(manifest); err != nil {
		return fmt.Errorf("files validation failed: %w", err)
	}

	if err := v.validateMigrations(manifest); err != nil {
		return fmt.Errorf("migrations validation failed: %w", err)
	}

	return nil
}

func (v *Validator) validateMetadata(manifest *ritual.Manifest) error {
	if manifest.Ritual.Name == "" {
		return fmt.Errorf("ritual name is required")
	}

	// Validate name format (lowercase, alphanumeric, hyphens)
	if !regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(manifest.Ritual.Name) {
		return fmt.Errorf("ritual name must start with lowercase letter and contain only lowercase letters, numbers, and hyphens")
	}

	if manifest.Ritual.Version == "" {
		return fmt.Errorf("ritual version is required")
	}

	// Validate semantic version format
	if !v.isValidSemver(manifest.Ritual.Version) {
		return fmt.Errorf("ritual version must be valid semantic version (e.g., 1.0.0)")
	}

	// Validate template engine
	if manifest.Ritual.TemplateEngine != "" {
		validEngines := []string{"fith", "go-template"}
		if !contains(validEngines, manifest.Ritual.TemplateEngine) {
			return fmt.Errorf("invalid template engine: %s (must be one of: %s)", 
				manifest.Ritual.TemplateEngine, strings.Join(validEngines, ", "))
		}
	}

	return nil
}

func (v *Validator) validateCompatibility(manifest *ritual.Manifest) error {
	if manifest.Compatibility.MinToutaVersion != "" {
		if !v.isValidSemver(manifest.Compatibility.MinToutaVersion) {
			return fmt.Errorf("min_touta_version must be valid semantic version")
		}
	}

	if manifest.Compatibility.MaxToutaVersion != "" {
		if !v.isValidSemver(manifest.Compatibility.MaxToutaVersion) {
			return fmt.Errorf("max_touta_version must be valid semantic version")
		}
	}

	if manifest.Compatibility.MinGoVersion != "" {
		if !v.isValidGoVersion(manifest.Compatibility.MinGoVersion) {
			return fmt.Errorf("min_go_version must be valid Go version")
		}
	}

	return nil
}

func (v *Validator) validateQuestions(manifest *ritual.Manifest) error {
	questionNames := make(map[string]bool)

	for i, q := range manifest.Questions {
		if q.Name == "" {
			return fmt.Errorf("question %d: name is required", i)
		}

		// Check for duplicate names
		if questionNames[q.Name] {
			return fmt.Errorf("question %s: duplicate question name", q.Name)
		}
		questionNames[q.Name] = true

		if q.Prompt == "" {
			return fmt.Errorf("question %s: prompt is required", q.Name)
		}

		if q.Type == "" {
			return fmt.Errorf("question %s: type is required", q.Name)
		}

		// Validate type-specific requirements
		switch q.Type {
		case ritual.QuestionTypeChoice, ritual.QuestionTypeMultiChoice:
			if len(q.Choices) == 0 {
				return fmt.Errorf("question %s: choices required for choice type", q.Name)
			}
		}

		// Validate conditions reference existing questions
		if q.Condition != nil {
			if q.Condition.Field == "" {
				return fmt.Errorf("question %s: condition field is required", q.Name)
			}
			// Note: We can't validate if field exists yet since questions are processed in order
			// This would need a second pass
		}

		// Validate validation rules
		if q.Validate != nil {
			if q.Validate.Pattern != "" {
				if _, err := regexp.Compile(q.Validate.Pattern); err != nil {
					return fmt.Errorf("question %s: invalid regex pattern: %w", q.Name, err)
				}
			}
		}
	}

	return nil
}

func (v *Validator) validateFiles(manifest *ritual.Manifest) error {
	// Validate template mappings
	for i, tmpl := range manifest.Files.Templates {
		if tmpl.Source == "" {
			return fmt.Errorf("template %d: source is required", i)
		}
		if tmpl.Destination == "" {
			return fmt.Errorf("template %d: destination is required", i)
		}
	}

	// Validate static file mappings
	for i, static := range manifest.Files.Static {
		if static.Source == "" {
			return fmt.Errorf("static file %d: source is required", i)
		}
		if static.Destination == "" {
			return fmt.Errorf("static file %d: destination is required", i)
		}
	}

	return nil
}

func (v *Validator) validateMigrations(manifest *ritual.Manifest) error {
	for i, m := range manifest.Migrations {
		if m.FromVersion == "" {
			return fmt.Errorf("migration %d: from_version is required", i)
		}
		if m.ToVersion == "" {
			return fmt.Errorf("migration %d: to_version is required", i)
		}

		if !v.isValidSemver(m.FromVersion) {
			return fmt.Errorf("migration %d: from_version must be valid semantic version", i)
		}
		if !v.isValidSemver(m.ToVersion) {
			return fmt.Errorf("migration %d: to_version must be valid semantic version", i)
		}

		// Check that at least one up handler is defined
		if len(m.Up.SQL) == 0 && m.Up.Script == "" && m.Up.GoCode == "" {
			return fmt.Errorf("migration %s->%s: at least one up handler (sql, script, or go_code) is required",
				m.FromVersion, m.ToVersion)
		}

		// Warn if no down handler (but don't error)
		if len(m.Down.SQL) == 0 && m.Down.Script == "" && m.Down.GoCode == "" {
			// Down handler is optional but recommended
		}
	}

	return nil
}

// isValidSemver checks if a version string is valid semantic version
func (v *Validator) isValidSemver(version string) bool {
	pattern := `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
	return regexp.MustCompile(pattern).MatchString(version)
}

// isValidGoVersion checks if a version string is valid Go version
func (v *Validator) isValidGoVersion(version string) bool {
	pattern := `^(0|[1-9]\d*)\.(0|[1-9]\d*)(?:\.(0|[1-9]\d*))?$`
	return regexp.MustCompile(pattern).MatchString(version)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

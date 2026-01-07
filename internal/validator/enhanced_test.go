package validator

import (
	"strings"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// TestValidateTemplateReferences tests that template files referenced in ritual.yaml exist
func TestValidateTemplateReferences(t *testing.T) {
	tmpDir := t.TempDir()
	
	v := NewValidator()
	v.SetRitualPath(tmpDir)

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test",
			Version: "1.0.0",
		},
		Files: ritual.FilesSection{
			Templates: []ritual.FileMapping{
				{Source: "templates/main.go.tmpl", Destination: "main.go"},
				{Source: "templates/missing.go.tmpl", Destination: "missing.go"},
			},
		},
	}

	err := v.ValidateFileReferences(manifest)
	if err == nil {
		t.Error("Expected error for missing template file")
	}
}

// TestValidateDependencyVersions tests version constraint validation
func TestValidateDependencyVersions(t *testing.T) {
	v := NewValidator()

	testCases := []struct {
		name        string
		manifest    *ritual.Manifest
		shouldError bool
	}{
		{
			name: "valid version constraints",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Compatibility: ritual.Compatibility{
					MinToutaVersion: "1.0.0",
					MaxToutaVersion: "2.0.0",
				},
			},
			shouldError: false,
		},
		{
			name: "min > max version",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Compatibility: ritual.Compatibility{
					MinToutaVersion: "2.0.0",
					MaxToutaVersion: "1.0.0",
				},
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateVersionConstraints(tc.manifest)
			if tc.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestValidateQuestionConditions tests questionnaire conditional logic
func TestValidateQuestionConditions(t *testing.T) {
	v := NewValidator()

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test",
			Version: "1.0.0",
		},
		Questions: []ritual.Question{
			{
				Name:     "database",
				Type:     "choice",
				Choices:  []string{"postgres", "mysql", "none"},
				Required: true,
			},
			{
				Name: "db_host",
				Type: "text",
				Condition: &ritual.QuestionCondition{
					Field:     "database",
					NotEquals: "none",
				},
			},
			{
				Name: "invalid_condition",
				Type: "text",
				Condition: &ritual.QuestionCondition{
					Field:     "nonexistent_field",
					NotEquals: "value",
				},
			},
		},
	}

	err := v.ValidateQuestionConditions(manifest)
	if err == nil {
		t.Error("Expected error for invalid condition field reference")
	}
}

// TestValidateCircularConditions tests for circular conditional dependencies
func TestValidateCircularConditions(t *testing.T) {
	v := NewValidator()

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test",
			Version: "1.0.0",
		},
		Questions: []ritual.Question{
			{
				Name: "field_a",
				Type: "text",
				Condition: &ritual.QuestionCondition{
					Field:  "field_b",
					Equals: "yes",
				},
			},
			{
				Name: "field_b",
				Type: "text",
				Condition: &ritual.QuestionCondition{
					Field:  "field_a",
					Equals: "yes",
				},
			},
		},
	}

	err := v.ValidateQuestionConditions(manifest)
	if err == nil {
		t.Error("Expected error for circular condition dependencies")
	}
}

// TestValidateCommonMistakes tests detection of common ritual authoring mistakes
func TestValidateCommonMistakes(t *testing.T) {
	v := NewValidator()

	testCases := []struct {
		name        string
		manifest    *ritual.Manifest
		shouldWarn  bool
		description string
	}{
		{
			name: "unprotected config files",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Files: ritual.FilesSection{
					Templates: []ritual.FileMapping{
						{Source: "config.yaml.tmpl", Destination: "config.yaml"},
					},
					Protected: []string{}, // Should warn - config.yaml not protected
				},
			},
			shouldWarn:  true,
			description: "Config files should be protected",
		},
		{
			name: "env files not protected",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Files: ritual.FilesSection{
					Templates: []ritual.FileMapping{
						{Source: ".env.tmpl", Destination: ".env"},
					},
					Protected: []string{},
				},
			},
			shouldWarn:  true,
			description: ".env files should be protected",
		},
		{
			name: "proper protection",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Files: ritual.FilesSection{
					Templates: []ritual.FileMapping{
						{Source: "config.yaml.tmpl", Destination: "config.yaml"},
					},
					Protected: []string{"config.yaml"},
				},
			},
			shouldWarn:  false,
			description: "Properly protected config",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			warnings := v.CheckCommonMistakes(tc.manifest)
			hasRelevantWarning := false
			for _, w := range warnings {
				if strings.Contains(w, "protected") || strings.Contains(w, "config") || strings.Contains(w, ".env") {
					hasRelevantWarning = true
					break
				}
			}

			if tc.shouldWarn && !hasRelevantWarning {
				t.Errorf("Expected warning but got none. Description: %s", tc.description)
			}
			if !tc.shouldWarn && hasRelevantWarning {
				t.Errorf("Expected no warning but got: %v. Description: %s", warnings, tc.description)
			}
		})
	}
}

// TestValidateMigrationReversibility tests that migrations have down handlers
func TestValidateMigrationReversibility(t *testing.T) {
	v := NewValidator()

	manifest := &ritual.Manifest{
		Ritual: ritual.RitualMeta{
			Name:    "test",
			Version: "1.0.0",
		},
		Migrations: []ritual.Migration{
			{
				FromVersion: "1.0.0",
				ToVersion:   "1.1.0",
				Up: ritual.MigrationHandler{
					SQL: []string{"CREATE TABLE users"},
				},
				Down: ritual.MigrationHandler{
					SQL: []string{"DROP TABLE users"},
				},
			},
			{
				FromVersion: "1.1.0",
				ToVersion:   "1.2.0",
				Up: ritual.MigrationHandler{
					SQL: []string{"ALTER TABLE users ADD COLUMN email"},
				},
				// Missing Down handler
			},
		},
	}

	warnings := v.CheckMigrationReversibility(manifest)
	if len(warnings) == 0 {
		t.Error("Expected warning for non-reversible migration")
	}
}

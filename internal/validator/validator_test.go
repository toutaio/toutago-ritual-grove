package validator

import (
	"testing"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestValidateMetadata(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		manifest  *ritual.Manifest
		wantError bool
	}{
		{
			name: "valid metadata",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test-app",
					Version: "1.0.0",
				},
			},
			wantError: false,
		},
		{
			name: "missing name",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Version: "1.0.0",
				},
			},
			wantError: true,
		},
		{
			name: "invalid name format",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "Test-App",
					Version: "1.0.0",
				},
			},
			wantError: true,
		},
		{
			name: "missing version",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name: "test-app",
				},
			},
			wantError: true,
		},
		{
			name: "invalid version format",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test-app",
					Version: "1.0",
				},
			},
			wantError: true,
		},
		{
			name: "invalid template engine",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:           "test-app",
					Version:        "1.0.0",
					TemplateEngine: "invalid",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.manifest)
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateQuestions(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		manifest  *ritual.Manifest
		wantError bool
	}{
		{
			name: "valid questions",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Questions: []ritual.Question{
					{
						Name:   "app_name",
						Prompt: "App name?",
						Type:   ritual.QuestionTypeText,
					},
				},
			},
			wantError: false,
		},
		{
			name: "duplicate question names",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Questions: []ritual.Question{
					{
						Name:   "app_name",
						Prompt: "App name?",
						Type:   ritual.QuestionTypeText,
					},
					{
						Name:   "app_name",
						Prompt: "App name again?",
						Type:   ritual.QuestionTypeText,
					},
				},
			},
			wantError: true,
		},
		{
			name: "choice without choices",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Questions: []ritual.Question{
					{
						Name:   "database",
						Prompt: "Database?",
						Type:   ritual.QuestionTypeChoice,
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid regex pattern",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Questions: []ritual.Question{
					{
						Name:   "app_name",
						Prompt: "App name?",
						Type:   ritual.QuestionTypeText,
						Validate: &ritual.ValidationRule{
							Pattern: "[invalid",
						},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.manifest)
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateMigrations(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		manifest  *ritual.Manifest
		wantError bool
	}{
		{
			name: "valid migration",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Migrations: []ritual.Migration{
					{
						FromVersion: "1.0.0",
						ToVersion:   "1.1.0",
						Up: ritual.MigrationHandler{
							SQL: []string{"CREATE TABLE users"},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "missing from_version",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Migrations: []ritual.Migration{
					{
						ToVersion: "1.1.0",
						Up: ritual.MigrationHandler{
							SQL: []string{"CREATE TABLE users"},
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "missing up handler",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test-app",
					Version: "1.0.0",
				},
				Migrations: []ritual.Migration{
					{
						FromVersion: "1.0.0",
						ToVersion:   "1.1.0",
						Up:          ritual.MigrationHandler{},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.manifest)
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestIsValidSemver(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"0.0.1", true},
		{"1.2.3", true},
		{"1.0.0-alpha", true},
		{"1.0.0-alpha.1", true},
		{"1.0.0+build.123", true},
		{"1.0", false},
		{"1", false},
		{"v1.0.0", false},
		{"1.0.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := validator.isValidSemver(tt.version)
			if result != tt.valid {
				t.Errorf("isValidSemver(%s) = %v, want %v", tt.version, result, tt.valid)
			}
		})
	}
}

func TestIsValidGoVersion(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		version string
		valid   bool
	}{
		{"1.22.0", true},
		{"1.22", true},
		{"1.23", true},
		{"1.22.1", true},
		{"0.0.1", true},
		{"1.0", true},
		{"1", false},
		{"v1.22", false},
		{"1.22.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := validator.isValidGoVersion(tt.version)
			if result != tt.valid {
				t.Errorf("isValidGoVersion(%s) = %v, want %v", tt.version, result, tt.valid)
			}
		})
	}
}

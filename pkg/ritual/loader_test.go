package ritual

import (
	"testing"
)

func TestLoadFromBytes(t *testing.T) {
	yamlData := []byte(`
ritual:
  name: test-ritual
  version: 1.0.0
  description: Test ritual
  author: Test Author
  template_engine: fith

questions:
  - name: app_name
    prompt: Application name
    type: text
    required: true

files:
  templates:
    - src: templates/main.go
      dest: main.go
`)

	manifest, err := LoadFromBytes(yamlData)
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	if manifest.Ritual.Name != "test-ritual" {
		t.Errorf("Expected name 'test-ritual', got '%s'", manifest.Ritual.Name)
	}

	if manifest.Ritual.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", manifest.Ritual.Version)
	}

	if manifest.Ritual.TemplateEngine != "fith" {
		t.Errorf("Expected template_engine 'fith', got '%s'", manifest.Ritual.TemplateEngine)
	}

	if len(manifest.Questions) != 1 {
		t.Errorf("Expected 1 question, got %d", len(manifest.Questions))
	}

	if len(manifest.Files.Templates) != 1 {
		t.Errorf("Expected 1 template, got %d", len(manifest.Files.Templates))
	}
}

func TestManifestValidate(t *testing.T) {
	tests := []struct {
		name      string
		manifest  *Manifest
		wantError bool
	}{
		{
			name: "valid manifest",
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Questions: []Question{
					{
						Name:   "app_name",
						Prompt: "Name?",
						Type:   QuestionTypeText,
					},
				},
			},
			wantError: false,
		},
		{
			name: "missing name",
			manifest: &Manifest{
				Ritual: RitualMeta{
					Version: "1.0.0",
				},
			},
			wantError: true,
		},
		{
			name: "missing version",
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name: "test",
				},
			},
			wantError: true,
		},
		{
			name: "question missing name",
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Questions: []Question{
					{
						Prompt: "Name?",
						Type:   QuestionTypeText,
					},
				},
			},
			wantError: true,
		},
		{
			name: "choice question without choices",
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Questions: []Question{
					{
						Name:   "db",
						Prompt: "Database?",
						Type:   QuestionTypeChoice,
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate()
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestQuestionTypes(t *testing.T) {
	types := []QuestionType{
		QuestionTypeText,
		QuestionTypePassword,
		QuestionTypeChoice,
		QuestionTypeMultiChoice,
		QuestionTypeBoolean,
		QuestionTypeNumber,
		QuestionTypePath,
		QuestionTypeURL,
		QuestionTypeEmail,
	}

	for _, typ := range types {
		if typ == "" {
			t.Errorf("Question type should not be empty")
		}
	}
}

func TestDefaultTemplateEngine(t *testing.T) {
	yamlData := []byte(`
ritual:
  name: test-ritual
  version: 1.0.0
  description: Test without template engine
`)

	manifest, err := LoadFromBytes(yamlData)
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	if manifest.Ritual.TemplateEngine != "fith" {
		t.Errorf("Expected default template_engine 'fith', got '%s'", manifest.Ritual.TemplateEngine)
	}
}

package ritual

import (
	"os"
	"path/filepath"
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

func TestNewLoader(t *testing.T) {
	loader := NewLoader("/test/path")
	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}
}

func TestLoader_Load(t *testing.T) {
	// Create a temporary directory with a ritual.yaml
	tmpDir := t.TempDir()
	ritualDir := filepath.Join(tmpDir, "test-ritual")
	if err := os.MkdirAll(ritualDir, 0755); err != nil {
		t.Fatalf("Failed to create ritual directory: %v", err)
	}

	yamlContent := []byte(`
ritual:
  name: file-ritual
  version: 1.0.0
  description: Test ritual loaded from file
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

	if err := os.WriteFile(filepath.Join(ritualDir, "ritual.yaml"), yamlContent, 0644); err != nil {
		t.Fatalf("Failed to write test ritual file: %v", err)
	}

	loader := NewLoader("")
	manifest, err := loader.Load(ritualDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if manifest.Ritual.Name != "file-ritual" {
		t.Errorf("Expected name 'file-ritual', got '%s'", manifest.Ritual.Name)
	}
}

func TestLoader_Load_FileNotFound(t *testing.T) {
	loader := NewLoader("/tmp")
	_, err := loader.Load("/nonexistent/ritual.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestLoadFromBytes_InvalidYAML(t *testing.T) {
	yamlData := []byte(`
ritual:
  name: test
  invalid yaml here
    badly: indented
`)

	_, err := LoadFromBytes(yamlData)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestManifestValidate_ComprehensiveTests(t *testing.T) {
	tests := []struct {
		name      string
		manifest  *Manifest
		wantError bool
	}{
		{
			name: "valid with multiple questions",
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Questions: []Question{
					{Name: "q1", Prompt: "Q1", Type: QuestionTypeText},
					{Name: "q2", Prompt: "Q2", Type: QuestionTypePassword},
					{Name: "q3", Prompt: "Q3", Type: QuestionTypeBoolean},
				},
			},
			wantError: false,
		},
		{
			name: "valid choice with choices",
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Questions: []Question{
					{
						Name:    "db",
						Prompt:  "Database",
						Type:    QuestionTypeChoice,
						Choices: []string{"mysql", "postgres"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid multichoice with choices",
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Questions: []Question{
					{
						Name:    "features",
						Prompt:  "Features",
						Type:    QuestionTypeMultiChoice,
						Choices: []string{"auth", "api", "admin"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "multichoice without choices",
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Questions: []Question{
					{
						Name:   "features",
						Prompt: "Features",
						Type:   QuestionTypeMultiChoice,
					},
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
						Prompt: "Q1",
						Type:   QuestionTypeText,
					},
				},
			},
			wantError: true,
		},
		{
			name: "question missing prompt",
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Questions: []Question{
					{
						Name: "q1",
						Type: QuestionTypeText,
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

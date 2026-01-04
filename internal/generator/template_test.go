package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoTemplateEngineRender(t *testing.T) {
	engine := NewGoTemplateEngine()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "simple variable",
			template: "Hello {{ .name }}!",
			data:     map[string]interface{}{"name": "World"},
			want:     "Hello World!",
			wantErr:  false,
		},
		{
			name:     "multiple variables",
			template: "{{ .greeting }} {{ .name }}!",
			data: map[string]interface{}{
				"greeting": "Hello",
				"name":     "World",
			},
			want:    "Hello World!",
			wantErr: false,
		},
		{
			name:     "upper filter",
			template: "{{ upper .text }}",
			data:     map[string]interface{}{"text": "hello"},
			want:     "HELLO",
			wantErr:  false,
		},
		{
			name:     "pascal filter",
			template: "{{ pascal .name }}",
			data:     map[string]interface{}{"name": "my-app"},
			want:     "MyApp",
			wantErr:  false,
		},
		{
			name:     "snake filter",
			template: "{{ snake .name }}",
			data:     map[string]interface{}{"name": "MyApp"},
			want:     "my_app",
			wantErr:  false,
		},
		{
			name:     "kebab filter",
			template: "{{ kebab .name }}",
			data:     map[string]interface{}{"name": "MyApp"},
			want:     "my-app",
			wantErr:  false,
		},
		{
			name:     "invalid template",
			template: "{{ .missing",
			data:     map[string]interface{}{},
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, tt.data)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.want {
				t.Errorf("Expected '%s', got '%s'", tt.want, result)
			}
		})
	}
}

func TestGoTemplateEngineRenderFile(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Create a temporary template file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.tmpl")
	
	templateContent := "Hello {{ .name }}!\nPort: {{ .port }}"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	data := map[string]interface{}{
		"name": "TestApp",
		"port": 8080,
	}

	result, err := engine.RenderFile(templatePath, data)
	if err != nil {
		t.Fatalf("RenderFile failed: %v", err)
	}

	expected := "Hello TestApp!\nPort: 8080"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestNewTemplateEngine(t *testing.T) {
	tests := []struct {
		engineType string
		wantType   string
	}{
		{"go-template", "*generator.GoTemplateEngine"},
		{"fith", "*generator.FithTemplateEngine"},
		{"", "*generator.FithTemplateEngine"}, // default
		{"unknown", "*generator.FithTemplateEngine"}, // fallback to default
	}

	for _, tt := range tests {
		t.Run(tt.engineType, func(t *testing.T) {
			engine := NewTemplateEngine(tt.engineType)
			if engine == nil {
				t.Error("Expected engine but got nil")
			}
		})
	}
}

func TestFithTemplateEngineFallback(t *testing.T) {
	engine := NewFithTemplateEngine()

	template := "Hello {{ .name }}!"
	data := map[string]interface{}{"name": "World"}

	result, err := engine.Render(template, data)
	if err != nil {
		t.Errorf("Render failed: %v", err)
	}

	expected := "Hello World!"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

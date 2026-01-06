package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
	"gopkg.in/yaml.v3"
)

// TestFrontendQuestionSchema validates that rituals can include frontend choice.
func TestFrontendQuestionSchema(t *testing.T) {
	tests := []struct {
		name            string
		yamlContent     string
		expectValid     bool
		expectedChoices []string
	}{
		{
			name: "valid inertia-vue frontend choice",
			yamlContent: `
ritual:
  name: test-ritual
  version: 1.0.0
  description: Test
  author: Test
  license: MIT

questions:
  - name: frontend_type
    type: choice
    prompt: "Select frontend framework:"
    choices:
      - traditional
      - inertia-vue
      - htmx
    default: traditional
    required: true
`,
			expectValid:     true,
			expectedChoices: []string{"traditional", "inertia-vue", "htmx"},
		},
		{
			name: "valid ssr question depends on inertia",
			yamlContent: `
ritual:
  name: test-ritual
  version: 1.0.0
  description: Test
  author: Test
  license: MIT

questions:
  - name: frontend_type
    type: choice
    prompt: "Select frontend framework:"
    choices:
      - traditional
      - inertia-vue
    default: traditional
    required: true
    
  - name: enable_ssr
    type: boolean
    prompt: "Enable Server-Side Rendering (SSR)?"
    default: false
    condition:
      expression: "{{ eq .frontend_type \"inertia-vue\" }}"
`,
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config ritual.Manifest
			err := yaml.Unmarshal([]byte(tt.yamlContent), &config)

			if tt.expectValid {
				require.NoError(t, err, "Should parse valid YAML")
				assert.NotEmpty(t, config.Ritual.Name)

				// Find frontend_type question
				var frontendQ *ritual.Question
				for i := range config.Questions {
					if config.Questions[i].Name == "frontend_type" {
						frontendQ = &config.Questions[i]
						break
					}
				}

				if len(tt.expectedChoices) > 0 {
					require.NotNil(t, frontendQ, "Should have frontend_type question")
					assert.Equal(t, string(ritual.QuestionTypeChoice), string(frontendQ.Type))
					assert.Equal(t, tt.expectedChoices, frontendQ.Choices)
				}
			} else {
				assert.Error(t, err, "Should fail parsing invalid YAML")
			}
		})
	}
}

// TestBlogRitualFrontendChoice validates blog ritual has frontend choice.
func TestBlogRitualFrontendChoice(t *testing.T) {
	ritualPath := filepath.Join("..", "rituals", "blog", "ritual.yaml")

	data, err := os.ReadFile(ritualPath)
	require.NoError(t, err, "Should read blog ritual.yaml")

	var config ritual.Manifest
	err = yaml.Unmarshal(data, &config)
	require.NoError(t, err, "Should parse blog ritual.yaml")

	// Find frontend_type question
	var frontendQ *ritual.Question
	for i := range config.Questions {
		if config.Questions[i].Name == "frontend_type" {
			frontendQ = &config.Questions[i]
			break
		}
	}

	require.NotNil(t, frontendQ, "Blog ritual should have frontend_type question")
	assert.Equal(t, string(ritual.QuestionTypeChoice), string(frontendQ.Type))
	assert.Contains(t, frontendQ.Choices, "traditional")
	assert.Contains(t, frontendQ.Choices, "inertia-vue")
	assert.Contains(t, frontendQ.Choices, "htmx")
	assert.Equal(t, "traditional", frontendQ.Default)
}

// TestInertiaVueDependencies validates Inertia adds correct dependencies.
func TestInertiaVueDependencies(t *testing.T) {
	yamlContent := `
ritual:
  name: test-ritual
  version: 1.0.0
  description: Test
  author: Test
  license: MIT

dependencies:
  packages:
    - github.com/toutaio/toutago-cosan-router
    - github.com/toutaio/toutago-inertia
  npm_packages:
    - "@toutaio/inertia-vue"
    - "vue@^3.4.0"
  conditions:
    - if: "{{ eq .frontend_type \"inertia-vue\" }}"
      packages:
        - github.com/toutaio/toutago-inertia
`
	var config ritual.Manifest
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	require.NoError(t, err, "Should parse YAML with conditional dependencies")

	assert.Contains(t, config.Dependencies.Packages, "github.com/toutaio/toutago-inertia")
}

// TestFrontendTemplateGeneration validates conditional file generation.
func TestFrontendTemplateGeneration(t *testing.T) {
	yamlContent := `
ritual:
  name: test-ritual
  version: 1.0.0
  description: Test
  author: Test
  license: MIT

files:
  templates:
    - src: templates/main.go.tmpl
      dest: main.go
    
    - src: templates/frontend/inertia/app.js.tmpl
      dest: frontend/app.js
      condition: "{{ eq .frontend_type \"inertia-vue\" }}"
    
    - src: templates/frontend/htmx/index.html.tmpl
      dest: views/index.html
      condition: "{{ eq .frontend_type \"htmx\" }}"
`
	var config ritual.Manifest
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	require.NoError(t, err, "Should parse YAML with conditional files")

	assert.Len(t, config.Files.Templates, 3)

	// Verify conditional files have conditions
	hasCondition := false
	for _, file := range config.Files.Templates {
		if file.Condition != "" {
			hasCondition = true
			break
		}
	}
	assert.True(t, hasCondition, "Should have conditional files")
}

package ritual

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantMeta     map[string]interface{}
		wantTemplate string
		wantErr      bool
	}{
		{
			name: "yaml frontmatter",
			content: `---
output: cmd/main.go
mode: 0755
overwrite: false
---
package main

func main() {
	// {{ .ProjectName }}
}`,
			wantMeta: map[string]interface{}{
				"output":    "cmd/main.go",
				"mode":      0755,
				"overwrite": false,
			},
			wantTemplate: `package main

func main() {
	// {{ .ProjectName }}
}`,
		},
		{
			name: "no frontmatter",
			content: `package main

func main() {}`,
			wantMeta:     nil,
			wantTemplate: `package main

func main() {}`,
		},
		{
			name: "empty frontmatter",
			content: `---
---
package main`,
			wantMeta:     map[string]interface{}{},
			wantTemplate: `package main`,
		},
		{
			name: "complex frontmatter",
			content: `---
output: internal/{{ .Package }}/handler.go
mode: 0644
overwrite: true
depends:
  - name: github.com/gorilla/mux
    version: "1.8.0"
conditions:
  - type: feature
    value: api
---
package {{ .Package }}`,
			wantMeta: map[string]interface{}{
				"output":    "internal/{{ .Package }}/handler.go",
				"mode":      0644,
				"overwrite": true,
				"depends": []interface{}{
					map[string]interface{}{
						"name":    "github.com/gorilla/mux",
						"version": "1.8.0",
					},
				},
				"conditions": []interface{}{
					map[string]interface{}{
						"type":  "feature",
						"value": "api",
					},
				},
			},
			wantTemplate: `package {{ .Package }}`,
		},
		{
			name: "invalid yaml in frontmatter",
			content: `---
invalid: [yaml
---
content`,
			wantErr: true,
		},
		{
			name: "only opening delimiter",
			content: `---
output: file.go
package main`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, template, err := ParseFrontmatter(tt.content)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			
			if tt.wantMeta == nil {
				assert.Nil(t, meta)
			} else {
				assert.Equal(t, tt.wantMeta, meta)
			}
			
			assert.Equal(t, tt.wantTemplate, template)
		})
	}
}

func TestTemplateFrontmatter_GetString(t *testing.T) {
	fm := TemplateFrontmatter{
		"output": "file.go",
		"number": 123,
	}
	
	assert.Equal(t, "file.go", fm.GetString("output"))
	assert.Equal(t, "", fm.GetString("nonexistent"))
	assert.Equal(t, "", fm.GetString("number")) // Not a string
}

func TestTemplateFrontmatter_GetInt(t *testing.T) {
	fm := TemplateFrontmatter{
		"mode":   0755,
		"string": "hello",
	}
	
	assert.Equal(t, 0755, fm.GetInt("mode"))
	assert.Equal(t, 0, fm.GetInt("nonexistent"))
	assert.Equal(t, 0, fm.GetInt("string")) // Not an int
}

func TestTemplateFrontmatter_GetBool(t *testing.T) {
	fm := TemplateFrontmatter{
		"overwrite": true,
		"skip":      false,
		"string":    "hello",
	}
	
	assert.True(t, fm.GetBool("overwrite"))
	assert.False(t, fm.GetBool("skip"))
	assert.False(t, fm.GetBool("nonexistent"))
	assert.False(t, fm.GetBool("string")) // Not a bool
}

func TestTemplateFrontmatter_GetStringSlice(t *testing.T) {
	fm := TemplateFrontmatter{
		"tags": []interface{}{"tag1", "tag2"},
		"none": nil,
	}
	
	tags := fm.GetStringSlice("tags")
	assert.Equal(t, []string{"tag1", "tag2"}, tags)
	
	none := fm.GetStringSlice("none")
	assert.Empty(t, none)
}

func TestTemplateFrontmatter_Has(t *testing.T) {
	fm := TemplateFrontmatter{
		"output": "file.go",
		"empty":  nil,
	}
	
	assert.True(t, fm.Has("output"))
	assert.True(t, fm.Has("empty"))
	assert.False(t, fm.Has("nonexistent"))
}

func TestParseFrontmatter_RealWorldExample(t *testing.T) {
	content := `---
output: internal/handlers/{{ .Resource | snake_case }}_handler.go
mode: 0644
overwrite: false
tags:
  - handler
  - api
depends:
  - name: github.com/gorilla/mux
    version: "^1.8.0"
conditions:
  - answer: enable_api
    equals: true
---
package handlers

import (
	"net/http"
	"github.com/gorilla/mux"
)

type {{ .Resource | pascal_case }}Handler struct {}

func (h *{{ .Resource | pascal_case }}Handler) List(w http.ResponseWriter, r *http.Request) {
	// List all {{ .Resource }}
}
`

	meta, template, err := ParseFrontmatter(content)
	require.NoError(t, err)
	
	fm := TemplateFrontmatter(meta)
	assert.Equal(t, "internal/handlers/{{ .Resource | snake_case }}_handler.go", fm.GetString("output"))
	assert.Equal(t, 0644, fm.GetInt("mode"))
	assert.False(t, fm.GetBool("overwrite"))
	
	tags := fm.GetStringSlice("tags")
	assert.Equal(t, []string{"handler", "api"}, tags)
	
	assert.Contains(t, template, "package handlers")
	assert.Contains(t, template, "type {{ .Resource | pascal_case }}Handler struct {}")
}

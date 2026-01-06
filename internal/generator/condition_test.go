package generator

import (
	"testing"
)

func TestEvaluateCondition(t *testing.T) {
	tests := []struct {
		name      string
		condition string
		variables map[string]interface{}
		want      bool
		wantErr   bool
	}{
		{
			name:      "simple equality true",
			condition: "{{ eq .frontend_type \"inertia-vue\" }}",
			variables: map[string]interface{}{
				"frontend_type": "inertia-vue",
			},
			want: true,
		},
		{
			name:      "simple equality false",
			condition: "{{ eq .frontend_type \"inertia-vue\" }}",
			variables: map[string]interface{}{
				"frontend_type": "htmx",
			},
			want: false,
		},
		{
			name:      "empty condition is true",
			condition: "",
			variables: map[string]interface{}{},
			want:      true,
		},
		{
			name:      "boolean field true",
			condition: "{{ .enable_ssr }}",
			variables: map[string]interface{}{
				"enable_ssr": true,
			},
			want: true,
		},
		{
			name:      "boolean field false",
			condition: "{{ .enable_ssr }}",
			variables: map[string]interface{}{
				"enable_ssr": false,
			},
			want: false,
		},
		{
			name:      "not equals",
			condition: "{{ ne .frontend_type \"traditional\" }}",
			variables: map[string]interface{}{
				"frontend_type": "htmx",
			},
			want: true,
		},
		{
			name:      "complex condition with and",
			condition: "{{ and (eq .frontend_type \"inertia-vue\") .enable_ssr }}",
			variables: map[string]interface{}{
				"frontend_type": "inertia-vue",
				"enable_ssr":    true,
			},
			want: true,
		},
		{
			name:      "complex condition with or",
			condition: "{{ or (eq .frontend_type \"inertia-vue\") (eq .frontend_type \"htmx\") }}",
			variables: map[string]interface{}{
				"frontend_type": "htmx",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateCondition(tt.condition, tt.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("evaluateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldGenerateFile(t *testing.T) {
	tests := []struct {
		name      string
		condition string
		variables map[string]interface{}
		want      bool
	}{
		{
			name:      "no condition means generate",
			condition: "",
			variables: map[string]interface{}{},
			want:      true,
		},
		{
			name:      "inertia-vue frontend generates inertia files",
			condition: "{{ eq .frontend_type \"inertia-vue\" }}",
			variables: map[string]interface{}{
				"frontend_type": "inertia-vue",
			},
			want: true,
		},
		{
			name:      "traditional frontend skips inertia files",
			condition: "{{ eq .frontend_type \"inertia-vue\" }}",
			variables: map[string]interface{}{
				"frontend_type": "traditional",
			},
			want: false,
		},
		{
			name:      "htmx frontend generates htmx files",
			condition: "{{ eq .frontend_type \"htmx\" }}",
			variables: map[string]interface{}{
				"frontend_type": "htmx",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateCondition(tt.condition, tt.variables)
			if err != nil {
				t.Errorf("evaluateCondition() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("evaluateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

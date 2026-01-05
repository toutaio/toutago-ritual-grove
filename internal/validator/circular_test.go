package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestCircularDependencyDetector(t *testing.T) {
	tests := []struct {
		name      string
		manifests map[string]*ritual.Manifest
		startID   string
		wantCycle []string
		wantErr   bool
	}{
		{
			name: "no circular dependencies",
			manifests: map[string]*ritual.Manifest{
				"a": {
					Ritual: ritual.RitualMeta{Name: "a"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"b"},
					},
				},
				"b": {
					Ritual: ritual.RitualMeta{Name: "b"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"c"},
					},
				},
				"c": {
					Ritual: ritual.RitualMeta{Name: "c"},
				},
			},
			startID: "a",
			wantErr: false,
		},
		{
			name: "direct circular dependency",
			manifests: map[string]*ritual.Manifest{
				"a": {
					Ritual: ritual.RitualMeta{Name: "a"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"a"},
					},
				},
			},
			startID:   "a",
			wantCycle: []string{"a", "a"},
			wantErr:   true,
		},
		{
			name: "circular dependency through chain",
			manifests: map[string]*ritual.Manifest{
				"a": {
					Ritual: ritual.RitualMeta{Name: "a"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"b"},
					},
				},
				"b": {
					Ritual: ritual.RitualMeta{Name: "b"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"c"},
					},
				},
				"c": {
					Ritual: ritual.RitualMeta{Name: "c"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"a"},
					},
				},
			},
			startID:   "a",
			wantCycle: []string{"a", "b", "c", "a"},
			wantErr:   true,
		},
		{
			name: "multiple dependencies no cycle",
			manifests: map[string]*ritual.Manifest{
				"a": {
					Ritual: ritual.RitualMeta{Name: "a"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"b", "c"},
					},
				},
				"b": {
					Ritual: ritual.RitualMeta{Name: "b"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"d"},
					},
				},
				"c": {
					Ritual: ritual.RitualMeta{Name: "c"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"d"},
					},
				},
				"d": {
					Ritual: ritual.RitualMeta{Name: "d"},
				},
			},
			startID: "a",
			wantErr: false,
		},
		{
			name: "diamond with cycle",
			manifests: map[string]*ritual.Manifest{
				"a": {
					Ritual: ritual.RitualMeta{Name: "a"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"b", "c"},
					},
				},
				"b": {
					Ritual: ritual.RitualMeta{Name: "b"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"d"},
					},
				},
				"c": {
					Ritual: ritual.RitualMeta{Name: "c"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"d"},
					},
				},
				"d": {
					Ritual: ritual.RitualMeta{Name: "d"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"b"},
					},
				},
			},
			startID:   "a",
			wantCycle: []string{"b", "d", "b"}, // Cycle starts where detected
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewCircularDependencyDetector(tt.manifests)
			cycle, err := detector.DetectCycle(tt.startID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantCycle != nil {
					assert.Equal(t, tt.wantCycle, cycle)
				}
			} else {
				assert.NoError(t, err)
				assert.Nil(t, cycle)
			}
		})
	}
}

func TestCircularDependencyValidator(t *testing.T) {
	v := &Validator{}

	tests := []struct {
		name      string
		manifest  *ritual.Manifest
		context   map[string]*ritual.Manifest
		wantError bool
		errorMsg  string
	}{
		{
			name: "no dependencies",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{Name: "standalone"},
			},
			context:   map[string]*ritual.Manifest{},
			wantError: false,
		},
		{
			name: "valid dependency chain",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{Name: "main"},
				Dependencies: ritual.Dependencies{
					Rituals: []string{"base"},
				},
			},
			context: map[string]*ritual.Manifest{
				"base": {
					Ritual: ritual.RitualMeta{Name: "base"},
				},
			},
			wantError: false,
		},
		{
			name: "circular dependency detected",
			manifest: &ritual.Manifest{
				Ritual: ritual.RitualMeta{Name: "main"},
				Dependencies: ritual.Dependencies{
					Rituals: []string{"base"},
				},
			},
			context: map[string]*ritual.Manifest{
				"base": {
					Ritual: ritual.RitualMeta{Name: "base"},
					Dependencies: ritual.Dependencies{
						Rituals: []string{"main"},
					},
				},
			},
			wantError: true,
			errorMsg:  "circular dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.ValidateCircularDependencies(tt.manifest, tt.context)

			if tt.wantError {
				assert.NotEmpty(t, errs)
				if tt.errorMsg != "" {
					assert.Contains(t, errs[0].Error(), tt.errorMsg)
				}
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

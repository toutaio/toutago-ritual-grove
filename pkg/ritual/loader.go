package ritual

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader loads and parses ritual manifests
type Loader struct {
	basePath string
}

// NewLoader creates a new ritual loader
func NewLoader(basePath string) *Loader {
	return &Loader{
		basePath: basePath,
	}
}

// Load loads a ritual manifest from a directory
func (l *Loader) Load(ritualPath string) (*Manifest, error) {
	manifestPath := filepath.Join(ritualPath, "ritual.yaml")

	// #nosec G304 - manifestPath is a validated file path parameter
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ritual.yaml: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse ritual.yaml: %w", err)
	}

	// Set defaults
	if manifest.Ritual.TemplateEngine == "" {
		manifest.Ritual.TemplateEngine = "fith"
	}

	return &manifest, nil
}

// LoadFromBytes loads a manifest from byte data
func LoadFromBytes(data []byte) (*Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse ritual manifest: %w", err)
	}

	// Set defaults
	if manifest.Ritual.TemplateEngine == "" {
		manifest.Ritual.TemplateEngine = "fith"
	}

	return &manifest, nil
}

// Validate validates a ritual manifest
func (m *Manifest) Validate() error {
	if m.Ritual.Name == "" {
		return fmt.Errorf("ritual name is required")
	}

	if m.Ritual.Version == "" {
		return fmt.Errorf("ritual version is required")
	}

	// Validate questions
	for i, q := range m.Questions {
		if q.Name == "" {
			return fmt.Errorf("question %d: name is required", i)
		}
		if q.Prompt == "" {
			return fmt.Errorf("question %s: prompt is required", q.Name)
		}
		if q.Type == "" {
			return fmt.Errorf("question %s: type is required", q.Name)
		}

		// Validate choices for choice questions
		if (q.Type == QuestionTypeChoice || q.Type == QuestionTypeMultiChoice) && len(q.Choices) == 0 {
			return fmt.Errorf("question %s: choices required for choice type", q.Name)
		}
	}

	// Validate migrations
	for i, m := range m.Migrations {
		if m.FromVersion == "" {
			return fmt.Errorf("migration %d: from_version is required", i)
		}
		if m.ToVersion == "" {
			return fmt.Errorf("migration %d: to_version is required", i)
		}
		if len(m.Up.SQL) == 0 && m.Up.Script == "" && m.Up.GoCode == "" {
			return fmt.Errorf("migration %s->%s: up handler is required", m.FromVersion, m.ToVersion)
		}
	}

	return nil
}

package ritual

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// TemplateFrontmatter represents metadata extracted from template frontmatter
type TemplateFrontmatter map[string]interface{}

// ParseFrontmatter extracts YAML frontmatter from template content
// Frontmatter is delimited by --- at the start and end
func ParseFrontmatter(content string) (map[string]interface{}, string, error) {
	const delimiter = "---"
	
	// Check if content starts with frontmatter delimiter
	if !strings.HasPrefix(content, delimiter) {
		// No frontmatter, return entire content as template
		return nil, content, nil
	}
	
	// Find closing delimiter
	rest := content[len(delimiter):]
	endIndex := strings.Index(rest, "\n"+delimiter)
	
	if endIndex == -1 {
		return nil, "", fmt.Errorf("frontmatter opening delimiter found but no closing delimiter")
	}
	
	// Extract frontmatter and template
	frontmatterContent := rest[:endIndex]
	templateContent := rest[endIndex+len(delimiter)+1:]
	
	// Trim leading newline from template if present
	templateContent = strings.TrimPrefix(templateContent, "\n")
	
	// Parse frontmatter YAML
	var frontmatter map[string]interface{}
	if strings.TrimSpace(frontmatterContent) != "" {
		if err := yaml.Unmarshal([]byte(frontmatterContent), &frontmatter); err != nil {
			return nil, "", fmt.Errorf("failed to parse frontmatter YAML: %w", err)
		}
	} else {
		frontmatter = make(map[string]interface{})
	}
	
	return frontmatter, templateContent, nil
}

// GetString safely retrieves a string value from frontmatter
func (f TemplateFrontmatter) GetString(key string) string {
	if v, ok := f[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetInt safely retrieves an int value from frontmatter
func (f TemplateFrontmatter) GetInt(key string) int {
	if v, ok := f[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

// GetBool safely retrieves a bool value from frontmatter
func (f TemplateFrontmatter) GetBool(key string) bool {
	if v, ok := f[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetStringSlice safely retrieves a string slice from frontmatter
func (f TemplateFrontmatter) GetStringSlice(key string) []string {
	if v, ok := f[key]; ok {
		if slice, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return []string{}
}

// Has checks if a key exists in frontmatter
func (f TemplateFrontmatter) Has(key string) bool {
	_, ok := f[key]
	return ok
}

// Get retrieves a raw value from frontmatter
func (f TemplateFrontmatter) Get(key string) (interface{}, bool) {
	v, ok := f[key]
	return v, ok
}

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadAnswersFromFile loads ritual answers from a YAML or JSON file
func LoadAnswersFromFile(configPath string) (map[string]interface{}, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read file content
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(configPath))
	
	var answers map[string]interface{}
	
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(content, &answers); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(content, &answers); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s (supported: .yaml, .yml, .json)", ext)
	}

	return answers, nil
}

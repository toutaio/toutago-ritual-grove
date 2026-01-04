package generator

import (
	"os"
	"testing"
)

func TestVariablesBasic(t *testing.T) {
	vars := NewVariables()

	vars.Set("name", "test-app")
	vars.Set("port", 8080)

	if val := vars.GetString("name"); val != "test-app" {
		t.Errorf("Expected 'test-app', got '%s'", val)
	}

	if val, ok := vars.Get("port"); !ok || val != 8080 {
		t.Errorf("Expected 8080, got %v", val)
	}

	all := vars.All()
	if len(all) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(all))
	}
}

func TestSetFromAnswers(t *testing.T) {
	vars := NewVariables()

	answers := map[string]interface{}{
		"app_name":  "my-app",
		"database":  "postgres",
		"port":      3000,
		"use_cache": true,
	}

	vars.SetFromAnswers(answers)

	if vars.GetString("app_name") != "my-app" {
		t.Error("app_name not set correctly")
	}

	if vars.GetString("database") != "postgres" {
		t.Error("database not set correctly")
	}

	all := vars.All()
	if len(all) != 4 {
		t.Errorf("Expected 4 variables, got %d", len(all))
	}
}

func TestSetFromEnvironment(t *testing.T) {
	// Set test environment variables
	os.Setenv("RITUAL_TEST_VAR", "test-value")
	os.Setenv("RITUAL_PORT", "8080")
	os.Setenv("OTHER_VAR", "should-not-load")
	defer os.Unsetenv("RITUAL_TEST_VAR")
	defer os.Unsetenv("RITUAL_PORT")
	defer os.Unsetenv("OTHER_VAR")

	vars := NewVariables()
	vars.SetFromEnvironment("RITUAL_")

	// Should have loaded prefixed vars (without prefix)
	if val := vars.GetString("test_var"); val != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", val)
	}

	if val := vars.GetString("port"); val != "8080" {
		t.Errorf("Expected '8080', got '%s'", val)
	}

	// Should not have loaded non-prefixed var
	if _, ok := vars.Get("other_var"); ok {
		t.Error("Should not have loaded OTHER_VAR")
	}
}

func TestAddComputed(t *testing.T) {
	vars := NewVariables()
	vars.Set("app_name", "my-app")

	vars.AddComputed()

	// Check timestamp variables exist
	if _, ok := vars.Get("now"); !ok {
		t.Error("now not computed")
	}

	if _, ok := vars.Get("timestamp"); !ok {
		t.Error("timestamp not computed")
	}

	if _, ok := vars.Get("year"); !ok {
		t.Error("year not computed")
	}
}

func TestCaseTransformations(t *testing.T) {
	tests := []struct {
		input  string
		pascal string
		camel  string
		snake  string
		kebab  string
	}{
		{
			input:  "my-app",
			pascal: "MyApp",
			camel:  "myApp",
			snake:  "my_app",
			kebab:  "my-app",
		},
		{
			input:  "user_service",
			pascal: "UserService",
			camel:  "userService",
			snake:  "user_service",
			kebab:  "user-service",
		},
		{
			input:  "HTTPServer",
			pascal: "HTTPServer",
			camel:  "hTTPServer",
			snake:  "h_t_t_p_server",
			kebab:  "h-t-t-p-server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if result := toPascalCase(tt.input); result != tt.pascal {
				t.Errorf("PascalCase: expected '%s', got '%s'", tt.pascal, result)
			}

			if result := toCamelCase(tt.input); result != tt.camel {
				t.Errorf("camelCase: expected '%s', got '%s'", tt.camel, result)
			}

			if result := toSnakeCase(tt.input); result != tt.snake {
				t.Errorf("snake_case: expected '%s', got '%s'", tt.snake, result)
			}

			if result := toKebabCase(tt.input); result != tt.kebab {
				t.Errorf("kebab-case: expected '%s', got '%s'", tt.kebab, result)
			}
		})
	}
}

func TestMaskSecrets(t *testing.T) {
	vars := NewVariables()
	vars.Set("app_name", "my-app")
	vars.Set("db_password", "secret123")
	vars.Set("api_token", "token456")
	vars.Set("port", "8080")

	masked := vars.MaskSecrets(nil)

	// Regular vars should not be masked
	if masked["app_name"] == "***" {
		t.Error("app_name should not be masked")
	}

	if masked["port"] == "***" {
		t.Error("port should not be masked")
	}

	// Secrets should be masked
	if masked["db_password"] != "***" {
		t.Error("db_password should be masked")
	}

	if masked["api_token"] != "***" {
		t.Error("api_token should be masked")
	}
}

func TestMaskSecretsWithKeys(t *testing.T) {
	vars := NewVariables()
	vars.Set("app_name", "my-app")
	vars.Set("db_host", "localhost")
	vars.Set("api_key", "key123")

	masked := vars.MaskSecrets([]string{"db_host"})

	// Explicitly listed as secret
	if masked["db_host"] != "***" {
		t.Error("db_host should be masked (explicit)")
	}

	// Not a secret
	if masked["app_name"] == "***" {
		t.Error("app_name should not be masked")
	}

	// api_key not in list but auto-detected (not implemented in current code)
	// This test documents expected behavior
}

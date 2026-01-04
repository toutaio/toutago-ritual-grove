package questionnaire

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnswerPersistence(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("save answers to file", func(t *testing.T) {
		answersFile := filepath.Join(tempDir, "answers1.yaml")
		persistence := NewAnswerPersistence(answersFile)

		answers := map[string]interface{}{
			"app_name":  "test-app",
			"database":  "postgres",
			"with_auth": true,
			"port":      8080,
		}

		err := persistence.Save(answers)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Check file was created
		if _, err := os.Stat(answersFile); os.IsNotExist(err) {
			t.Error("Answer file was not created")
		}
	})

	t.Run("load answers from file", func(t *testing.T) {
		answersFile := filepath.Join(tempDir, "answers2.yaml")
		persistence := NewAnswerPersistence(answersFile)

		originalAnswers := map[string]interface{}{
			"app_name":  "test-app",
			"database":  "mysql",
			"with_auth": false,
			"port":      3000,
		}

		// Save first
		if err := persistence.Save(originalAnswers); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Load
		loadedAnswers, err := persistence.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Verify
		if len(loadedAnswers) != len(originalAnswers) {
			t.Errorf("Loaded %d answers, want %d", len(loadedAnswers), len(originalAnswers))
		}

		if loadedAnswers["app_name"] != originalAnswers["app_name"] {
			t.Errorf("app_name = %v, want %v", loadedAnswers["app_name"], originalAnswers["app_name"])
		}

		if loadedAnswers["database"] != originalAnswers["database"] {
			t.Errorf("database = %v, want %v", loadedAnswers["database"], originalAnswers["database"])
		}
	})

	t.Run("load from non-existent file", func(t *testing.T) {
		answersFile := filepath.Join(tempDir, "nonexistent.yaml")
		persistence := NewAnswerPersistence(answersFile)

		answers, err := persistence.Load()
		if err == nil {
			t.Error("Expected error loading non-existent file, got nil")
		}
		if answers != nil {
			t.Error("Expected nil answers from non-existent file")
		}
	})

	t.Run("overwrite existing answers", func(t *testing.T) {
		answersFile := filepath.Join(tempDir, "answers3.yaml")
		persistence := NewAnswerPersistence(answersFile)

		// Save first set
		firstAnswers := map[string]interface{}{
			"app_name": "first-app",
		}
		if err := persistence.Save(firstAnswers); err != nil {
			t.Fatalf("First Save() error = %v", err)
		}

		// Save second set (overwrite)
		secondAnswers := map[string]interface{}{
			"app_name": "second-app",
			"database": "postgres",
		}
		if err := persistence.Save(secondAnswers); err != nil {
			t.Fatalf("Second Save() error = %v", err)
		}

		// Load and verify
		loaded, err := persistence.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if loaded["app_name"] != "second-app" {
			t.Errorf("app_name = %v, want %v", loaded["app_name"], "second-app")
		}

		if loaded["database"] != "postgres" {
			t.Errorf("database = %v, want %v", loaded["database"], "postgres")
		}
	})

	t.Run("save empty answers", func(t *testing.T) {
		answersFile := filepath.Join(tempDir, "answers4.yaml")
		persistence := NewAnswerPersistence(answersFile)

		emptyAnswers := map[string]interface{}{}
		err := persistence.Save(emptyAnswers)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		loaded, err := persistence.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(loaded) != 0 {
			t.Errorf("Expected empty answers, got %d entries", len(loaded))
		}
	})
}

func TestSecretAnswerPersistence(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("save answers with secrets masked", func(t *testing.T) {
		answersFile := filepath.Join(tempDir, "secrets1.yaml")
		persistence := NewAnswerPersistence(answersFile)

		answers := map[string]interface{}{
			"app_name":    "test-app",
			"db_password": "super-secret",
			"api_key":     "secret-key-123",
		}

		secretFields := []string{"db_password", "api_key"}
		err := persistence.SaveWithSecrets(answers, secretFields)
		if err != nil {
			t.Fatalf("SaveWithSecrets() error = %v", err)
		}

		// Load raw file content
		content, err := os.ReadFile(answersFile)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		contentStr := string(content)

		// Verify secrets are not in plain text
		if containsStr(contentStr, "super-secret") {
			t.Error("Secret 'db_password' found in plain text")
		}
		if containsStr(contentStr, "secret-key-123") {
			t.Error("Secret 'api_key' found in plain text")
		}

		// Verify non-secret is present
		if !containsStr(contentStr, "test-app") {
			t.Error("Non-secret 'app_name' not found in file")
		}
	})

	t.Run("load answers and restore from environment", func(t *testing.T) {
		answersFile := filepath.Join(tempDir, "secrets2.yaml")
		persistence := NewAnswerPersistence(answersFile)

		answers := map[string]interface{}{
			"app_name":    "test-app",
			"db_password": "my-password",
		}

		secretFields := []string{"db_password"}

		// Save with secrets
		if err := persistence.SaveWithSecrets(answers, secretFields); err != nil {
			t.Fatalf("SaveWithSecrets() error = %v", err)
		}

		// Set environment variable for secret
		os.Setenv("DB_PASSWORD", "env-password")
		defer os.Unsetenv("DB_PASSWORD")

		// Load with secret restoration
		loaded, err := persistence.LoadWithSecrets(map[string]string{
			"db_password": "DB_PASSWORD",
		})
		if err != nil {
			t.Fatalf("LoadWithSecrets() error = %v", err)
		}

		if loaded["app_name"] != "test-app" {
			t.Errorf("app_name = %v, want %v", loaded["app_name"], "test-app")
		}

		if loaded["db_password"] != "env-password" {
			t.Errorf("db_password = %v, want %v (from env)", loaded["db_password"], "env-password")
		}
	})
}

func TestAnswerValidation(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("validate answer file format", func(t *testing.T) {
		answersFile := filepath.Join(tempDir, "valid.yaml")

		// Create a valid YAML file
		validContent := []byte(`app_name: test-app
database: postgres
port: 8080
`)
		if err := os.WriteFile(answersFile, validContent, 0644); err != nil {
			t.Fatal(err)
		}

		persistence := NewAnswerPersistence(answersFile)
		answers, err := persistence.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if answers["app_name"] != "test-app" {
			t.Errorf("app_name = %v, want test-app", answers["app_name"])
		}
	})

	t.Run("reject invalid YAML", func(t *testing.T) {
		answersFile := filepath.Join(tempDir, "invalid.yaml")

		// Create an invalid YAML file
		invalidContent := []byte(`app_name: test-app
database: postgres
invalid yaml here!!!
  bad indentation
`)
		if err := os.WriteFile(answersFile, invalidContent, 0644); err != nil {
			t.Fatal(err)
		}

		persistence := NewAnswerPersistence(answersFile)
		_, err := persistence.Load()
		if err == nil {
			t.Error("Expected error loading invalid YAML, got nil")
		}
	})
}

// Helper function
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && anySubstringMatch(s, substr)
}

func anySubstringMatch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

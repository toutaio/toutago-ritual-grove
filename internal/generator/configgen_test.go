package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigGenerator_GenerateEnvExample(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewConfigGenerator()

	config := AppConfig{
		AppName: "my-app",
		Port:    8080,
		Database: DatabaseConfig{
			Type: "postgres",
			Host: "localhost",
			Port: 5432,
			Name: "myapp_db",
		},
	}

	err := gen.GenerateEnvExample(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateEnvExample() error = %v", err)
	}

	envPath := filepath.Join(tmpDir, ".env.example")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Error(".env.example should be created")
	}

	content, _ := os.ReadFile(envPath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "APP_NAME=my-app") {
		t.Error(".env should contain app name")
	}

	if !strings.Contains(contentStr, "PORT=8080") {
		t.Error(".env should contain port")
	}

	if !strings.Contains(contentStr, "DB_HOST=localhost") {
		t.Error(".env should contain database host")
	}
}

func TestConfigGenerator_GenerateYAMLConfig(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewConfigGenerator()

	config := AppConfig{
		AppName: "test-app",
		Port:    3000,
		Database: DatabaseConfig{
			Type: "mysql",
			Host: "db.example.com",
			Port: 3306,
		},
	}

	err := gen.GenerateYAMLConfig(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateYAMLConfig() error = %v", err)
	}

	configPath := filepath.Join(tmpDir, "config", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.yaml should be created")
	}

	content, _ := os.ReadFile(configPath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "test-app") {
		t.Error("YAML config should contain app name")
	}

	if !strings.Contains(contentStr, "mysql") {
		t.Error("YAML config should contain database type")
	}
}

func TestConfigGenerator_GenerateDockerCompose(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewConfigGenerator()

	config := DockerConfig{
		AppName:  "my-service",
		Port:     8080,
		Database: "postgres",
	}

	err := gen.GenerateDockerCompose(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateDockerCompose() error = %v", err)
	}

	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		t.Error("docker-compose.yml should be created")
	}

	content, _ := os.ReadFile(composePath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "my-service") {
		t.Error("Docker compose should contain service name")
	}

	if !strings.Contains(contentStr, "postgres") {
		t.Error("Docker compose should contain database")
	}

	if !strings.Contains(contentStr, "8080") {
		t.Error("Docker compose should contain port")
	}
}

func TestConfigGenerator_GenerateDockerfile(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewConfigGenerator()

	config := DockerfileConfig{
		GoVersion: "1.21",
		AppName:   "myapp",
		Port:      8080,
	}

	err := gen.GenerateDockerfile(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateDockerfile() error = %v", err)
	}

	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Error("Dockerfile should be created")
	}

	content, _ := os.ReadFile(dockerfilePath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "golang:1.21") {
		t.Error("Dockerfile should contain Go version")
	}

	if !strings.Contains(contentStr, "EXPOSE 8080") {
		t.Error("Dockerfile should expose port")
	}
}

func TestConfigGenerator_GenerateGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewConfigGenerator()

	err := gen.GenerateGitignore(tmpDir)
	if err != nil {
		t.Fatalf("GenerateGitignore() error = %v", err)
	}

	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore should be created")
	}

	content, _ := os.ReadFile(gitignorePath)
	contentStr := string(content)

	expectedPatterns := []string{
		"*.exe",
		".env",
		"tmp/",
		"vendor/",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(contentStr, pattern) {
			t.Errorf(".gitignore should contain %s", pattern)
		}
	}
}

func TestConfigGenerator_GenerateEditorConfig(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewConfigGenerator()

	err := gen.GenerateEditorConfig(tmpDir)
	if err != nil {
		t.Fatalf("GenerateEditorConfig() error = %v", err)
	}

	editorconfigPath := filepath.Join(tmpDir, ".editorconfig")
	if _, err := os.Stat(editorconfigPath); os.IsNotExist(err) {
		t.Error(".editorconfig should be created")
	}

	content, _ := os.ReadFile(editorconfigPath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "indent_style") {
		t.Error(".editorconfig should contain indent style")
	}
}

func TestConfigGenerator_GenerateMakefile(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewConfigGenerator()

	config := MakefileConfig{
		AppName:    "myapp",
		BinaryName: "myapp-bin",
	}

	err := gen.GenerateMakefile(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateMakefile() error = %v", err)
	}

	makefilePath := filepath.Join(tmpDir, "Makefile")
	if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
		t.Error("Makefile should be created")
	}

	content, _ := os.ReadFile(makefilePath)
	contentStr := string(content)

	expectedTargets := []string{
		"build:",
		"test:",
		"run:",
		"clean:",
	}

	for _, target := range expectedTargets {
		if !strings.Contains(contentStr, target) {
			t.Errorf("Makefile should contain %s target", target)
		}
	}
}

func TestConfigGenerator_GenerateAll(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewConfigGenerator()

	fullConfig := FullConfig{
		AppConfig: AppConfig{
			AppName: "test-app",
			Port:    8080,
			Database: DatabaseConfig{
				Type: "postgres",
				Host: "localhost",
				Port: 5432,
			},
		},
		GenerateDocker:       true,
		GenerateGitignore:    true,
		GenerateEditorConfig: true,
		GenerateMakefile:     true,
	}

	err := gen.GenerateAll(tmpDir, fullConfig)
	if err != nil {
		t.Fatalf("GenerateAll() error = %v", err)
	}

	expectedFiles := []string{
		".env.example",
		".gitignore",
		".editorconfig",
		"Makefile",
		"Dockerfile",
		"docker-compose.yml",
		"config/config.yaml",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("File %s should be created", file)
		}
	}
}

func TestConfigGenerator_GenerateForDatabase(t *testing.T) {
	tests := []struct {
		name     string
		database string
		wantPort int
	}{
		{"postgres", "postgres", 5432},
		{"mysql", "mysql", 3306},
		{"mongodb", "mongodb", 27017},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			gen := NewConfigGenerator()

			config := AppConfig{
				AppName: "test",
				Database: DatabaseConfig{
					Type: tt.database,
					Host: "localhost",
					Port: tt.wantPort,
				},
			}

			err := gen.GenerateEnvExample(tmpDir, config)
			if err != nil {
				t.Fatalf("GenerateEnvExample() error = %v", err)
			}

			content, _ := os.ReadFile(filepath.Join(tmpDir, ".env.example"))
			contentStr := string(content)

			if !strings.Contains(contentStr, tt.database) {
				t.Errorf("Should contain database type %s", tt.database)
			}
		})
	}
}

package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigGenerator generates configuration files
type ConfigGenerator struct{}

// NewConfigGenerator creates a new configuration generator
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{}
}

// AppConfig contains application configuration
type AppConfig struct {
	AppName     string
	Port        int
	Environment string
	Database    DatabaseConfig
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Type     string
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

// DockerConfig contains Docker configuration
type DockerConfig struct {
	AppName  string
	Port     int
	Database string
}

// DockerfileConfig contains Dockerfile configuration
type DockerfileConfig struct {
	GoVersion string
	AppName   string
	Port      int
}

// MakefileConfig contains Makefile configuration
type MakefileConfig struct {
	AppName    string
	BinaryName string
}

// FullConfig contains all configuration options
type FullConfig struct {
	AppConfig            AppConfig
	GenerateDocker       bool
	GenerateGitignore    bool
	GenerateEditorConfig bool
	GenerateMakefile     bool
}

// GenerateEnvExample generates a .env.example file
func (g *ConfigGenerator) GenerateEnvExample(targetPath string, config AppConfig) error {
	content := fmt.Sprintf(`# Application Configuration
APP_NAME=%s
PORT=%d
ENVIRONMENT=%s

# Database Configuration
DB_TYPE=%s
DB_HOST=%s
DB_PORT=%d
DB_NAME=%s
DB_USER=%s
DB_PASSWORD=%s

# Security
SECRET_KEY=change-me-in-production
JWT_SECRET=change-me-in-production

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
`,
		config.AppName,
		config.Port,
		getEnvOrDefault(config.Environment, "development"),
		config.Database.Type,
		config.Database.Host,
		config.Database.Port,
		getStrOrDefault(config.Database.Name, config.AppName+"_db"),
		getStrOrDefault(config.Database.User, config.AppName),
		getStrOrDefault(config.Database.Password, ""),
	)
	
	envPath := filepath.Join(targetPath, ".env.example")
	return os.WriteFile(envPath, []byte(content), 0644)
}

// GenerateYAMLConfig generates a config.yaml file
func (g *ConfigGenerator) GenerateYAMLConfig(targetPath string, config AppConfig) error {
	content := fmt.Sprintf(`app:
  name: %s
  port: %d
  environment: %s

database:
  type: %s
  host: %s
  port: %d
  name: %s
  user: %s
  max_connections: 25
  max_idle: 10
  conn_max_lifetime: 3600

server:
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s
  graceful_timeout: 30s

logging:
  level: info
  format: json
  output: stdout
`,
		config.AppName,
		config.Port,
		getEnvOrDefault(config.Environment, "development"),
		config.Database.Type,
		config.Database.Host,
		config.Database.Port,
		getStrOrDefault(config.Database.Name, config.AppName+"_db"),
		getStrOrDefault(config.Database.User, config.AppName),
	)
	
	configDir := filepath.Join(targetPath, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	
	configPath := filepath.Join(configDir, "config.yaml")
	return os.WriteFile(configPath, []byte(content), 0644)
}

// GenerateDockerCompose generates a docker-compose.yml file
func (g *ConfigGenerator) GenerateDockerCompose(targetPath string, config DockerConfig) error {
	var dbService string
	
	switch config.Database {
	case "postgres":
		dbService = `  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ` + config.AppName + `_db
      POSTGRES_USER: ` + config.AppName + `
      POSTGRES_PASSWORD: devpassword
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ` + config.AppName + `"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:`
	
	case "mysql":
		dbService = `  db:
    image: mysql:8
    environment:
      MYSQL_DATABASE: ` + config.AppName + `_db
      MYSQL_USER: ` + config.AppName + `
      MYSQL_PASSWORD: devpassword
      MYSQL_ROOT_PASSWORD: rootpassword
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mysql_data:`
	
	default:
		dbService = ""
	}
	
	content := fmt.Sprintf(`version: '3.8'

services:
  app:
    build: .
    ports:
      - "%d:%d"
    environment:
      - DB_HOST=db
      - DB_PORT=%s
      - DB_NAME=%s_db
      - DB_USER=%s
      - DB_PASSWORD=devpassword
    depends_on:
      - db
    restart: unless-stopped

%s
`,
		config.Port,
		config.Port,
		getDBPort(config.Database),
		config.AppName,
		config.AppName,
		dbService,
	)
	
	composePath := filepath.Join(targetPath, "docker-compose.yml")
	return os.WriteFile(composePath, []byte(content), 0644)
}

// GenerateDockerfile generates a Dockerfile
func (g *ConfigGenerator) GenerateDockerfile(targetPath string, config DockerfileConfig) error {
	content := fmt.Sprintf(`# Build stage
FROM golang:%s AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE %d

# Run the application
CMD ["./main"]
`,
		config.GoVersion,
		config.Port,
	)
	
	dockerfilePath := filepath.Join(targetPath, "Dockerfile")
	return os.WriteFile(dockerfilePath, []byte(content), 0644)
}

// GenerateGitignore generates a .gitignore file
func (g *ConfigGenerator) GenerateGitignore(targetPath string) error {
	content := `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
main
bin/
dist/

# Test binary
*.test
*.out

# Coverage
*.coverprofile
coverage.txt
coverage.html

# Dependencies
vendor/

# Environment
.env
.env.local
.env.*.local

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Temporary files
tmp/
temp/
*.tmp

# Logs
*.log
logs/

# Database
*.db
*.sqlite
*.sqlite3

# Build artifacts
*.tar.gz
*.zip
`
	
	gitignorePath := filepath.Join(targetPath, ".gitignore")
	return os.WriteFile(gitignorePath, []byte(content), 0644)
}

// GenerateEditorConfig generates a .editorconfig file
func (g *ConfigGenerator) GenerateEditorConfig(targetPath string) error {
	content := `# EditorConfig is awesome: https://EditorConfig.org

root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_style = tab
indent_size = 4

[*.{yml,yaml,json}]
indent_style = space
indent_size = 2

[*.md]
trim_trailing_whitespace = false

[Makefile]
indent_style = tab
`
	
	editorconfigPath := filepath.Join(targetPath, ".editorconfig")
	return os.WriteFile(editorconfigPath, []byte(content), 0644)
}

// GenerateMakefile generates a Makefile
func (g *ConfigGenerator) GenerateMakefile(targetPath string, config MakefileConfig) error {
	binaryName := config.BinaryName
	if binaryName == "" {
		binaryName = config.AppName
	}
	
	content := fmt.Sprintf(`.PHONY: build test run clean install lint fmt vet

BINARY_NAME=%s
MAIN_PATH=./cmd/main.go

build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

run:
	go run $(MAIN_PATH)

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out

install:
	go mod download
	go mod verify

lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

docker-build:
	docker build -t $(BINARY_NAME) .

docker-run:
	docker run -p 8080:8080 $(BINARY_NAME)

dev:
	air

.DEFAULT_GOAL := build
`,
		binaryName,
	)
	
	makefilePath := filepath.Join(targetPath, "Makefile")
	return os.WriteFile(makefilePath, []byte(content), 0644)
}

// GenerateAll generates all configuration files
func (g *ConfigGenerator) GenerateAll(targetPath string, config FullConfig) error {
	// Always generate environment example
	if err := g.GenerateEnvExample(targetPath, config.AppConfig); err != nil {
		return fmt.Errorf("failed to generate .env.example: %w", err)
	}
	
	// Always generate YAML config
	if err := g.GenerateYAMLConfig(targetPath, config.AppConfig); err != nil {
		return fmt.Errorf("failed to generate config.yaml: %w", err)
	}
	
	// Optional configurations
	if config.GenerateDocker {
		dockerConfig := DockerConfig{
			AppName:  config.AppConfig.AppName,
			Port:     config.AppConfig.Port,
			Database: config.AppConfig.Database.Type,
		}
		
		if err := g.GenerateDockerCompose(targetPath, dockerConfig); err != nil {
			return fmt.Errorf("failed to generate docker-compose.yml: %w", err)
		}
		
		dockerfileConfig := DockerfileConfig{
			GoVersion: "1.21",
			AppName:   config.AppConfig.AppName,
			Port:      config.AppConfig.Port,
		}
		
		if err := g.GenerateDockerfile(targetPath, dockerfileConfig); err != nil {
			return fmt.Errorf("failed to generate Dockerfile: %w", err)
		}
	}
	
	if config.GenerateGitignore {
		if err := g.GenerateGitignore(targetPath); err != nil {
			return fmt.Errorf("failed to generate .gitignore: %w", err)
		}
	}
	
	if config.GenerateEditorConfig {
		if err := g.GenerateEditorConfig(targetPath); err != nil {
			return fmt.Errorf("failed to generate .editorconfig: %w", err)
		}
	}
	
	if config.GenerateMakefile {
		makefileConfig := MakefileConfig{
			AppName:    config.AppConfig.AppName,
			BinaryName: config.AppConfig.AppName,
		}
		
		if err := g.GenerateMakefile(targetPath, makefileConfig); err != nil {
			return fmt.Errorf("failed to generate Makefile: %w", err)
		}
	}
	
	return nil
}

// Helper functions

func getEnvOrDefault(env, defaultEnv string) string {
	if env == "" {
		return defaultEnv
	}
	return env
}

func getStrOrDefault(str, defaultStr string) string {
	if str == "" {
		return defaultStr
	}
	return str
}

func getDBPort(dbType string) string {
	switch dbType {
	case "postgres":
		return "5432"
	case "mysql":
		return "3306"
	case "mongodb":
		return "27017"
	default:
		return "5432"
	}
}

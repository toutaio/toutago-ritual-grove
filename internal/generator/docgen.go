package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DocGenerator generates project documentation
type DocGenerator struct{}

// NewDocGenerator creates a new documentation generator
func NewDocGenerator() *DocGenerator {
	return &DocGenerator{}
}

// ProjectInfo contains project metadata
type ProjectInfo struct {
	Name        string
	Description string
	Version     string
	Author      string
	Database    string
	License     string
}

// APIEndpoint represents an API endpoint
type APIEndpoint struct {
	Method      string
	Path        string
	Description string
	Request     string
	Response    string
}

// Component represents a system component
type Component struct {
	Name        string
	Description string
}

// DeploymentConfig contains deployment configuration
type DeploymentConfig struct {
	Platform    string
	Database    string
	Environment string
}

// DocConfig configures documentation generation
type DocConfig struct {
	ProjectInfo          ProjectInfo
	APIEndpoints         []APIEndpoint
	Components           []Component
	DeploymentConfig     DeploymentConfig
	GenerateAPI          bool
	GenerateArchitecture bool
	GenerateDeployment   bool
	GenerateChangelog    bool
	GenerateContributing bool
	License              string
}

// GenerateREADME generates a README.md file
func (g *DocGenerator) GenerateREADME(targetPath string, info ProjectInfo) error {
	content := fmt.Sprintf(`# %s

%s

## Version

%s

## Features

- Built with Toutā Framework
- Database: %s
- RESTful API
- Comprehensive documentation

## Installation

`+"```bash"+`
git clone <repository-url>
cd %s
go mod download
`+"```"+`

## Configuration

Copy the example configuration file and edit as needed:

`+"```bash"+`
cp .env.example .env
`+"```"+`

## Running

`+"```bash"+`
go run cmd/main.go
`+"```"+`

## Testing

`+"```bash"+`
go test ./...
`+"```"+`

## Documentation

- [API Documentation](docs/API.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Deployment Guide](docs/DEPLOYMENT.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

%s

## Author

%s
`,
		info.Name,
		info.Description,
		info.Version,
		info.Database,
		info.Name,
		info.License,
		info.Author,
	)

	readmePath := filepath.Join(targetPath, "README.md")
	return os.WriteFile(readmePath, []byte(content), 0600)
}

// GenerateAPIDoc generates API documentation
func (g *DocGenerator) GenerateAPIDoc(targetPath string, endpoints []APIEndpoint) error {
	var sb strings.Builder

	sb.WriteString("# API Documentation\n\n")
	sb.WriteString("## Endpoints\n\n")

	for _, ep := range endpoints {
		sb.WriteString(fmt.Sprintf("### %s %s\n\n", ep.Method, ep.Path))
		sb.WriteString(fmt.Sprintf("%s\n\n", ep.Description))

		if ep.Request != "" {
			sb.WriteString(fmt.Sprintf("**Request Body:** `%s`\n\n", ep.Request))
		}

		if ep.Response != "" {
			sb.WriteString(fmt.Sprintf("**Response:** `%s`\n\n", ep.Response))
		}

		sb.WriteString("---\n\n")
	}

	docsDir := filepath.Join(targetPath, "docs")
	if err := os.MkdirAll(docsDir, 0750); err != nil {
		return err
	}

	apiDocPath := filepath.Join(docsDir, "API.md")
	return os.WriteFile(apiDocPath, []byte(sb.String()), 0600)
}

// GenerateChangelog generates a CHANGELOG.md file
func (g *DocGenerator) GenerateChangelog(targetPath, version string) error {
	content := fmt.Sprintf(`# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [%s] - %s

### Added

- Initial release
- Project scaffolding
- Basic API endpoints
- Database integration
- Documentation

### Changed

### Deprecated

### Removed

### Fixed

### Security
`,
		version,
		time.Now().Format("2006-01-02"),
	)

	changelogPath := filepath.Join(targetPath, "CHANGELOG.md")
	return os.WriteFile(changelogPath, []byte(content), 0600)
}

// GenerateContributing generates a CONTRIBUTING.md file
func (g *DocGenerator) GenerateContributing(targetPath string) error {
	content := `# Contributing

Thank you for considering contributing to this project!

## How to Contribute

1. Fork the repository
2. Create a feature branch (` + "`git checkout -b feature/amazing-feature`" + `)
3. Commit your changes (` + "`git commit -m 'Add amazing feature'`" + `)
4. Push to the branch (` + "`git push origin feature/amazing-feature`" + `)
5. Open a Pull Request

## Development Setup

` + "```bash" + `
git clone <your-fork>
cd <project>
go mod download
` + "```" + `

## Code Style

- Follow Go conventions and best practices
- Run ` + "`gofmt`" + ` before committing
- Write tests for new features
- Update documentation as needed

## Testing

Run the test suite:

` + "```bash" + `
go test ./...
` + "```" + `

## Pull Request Guidelines

- Keep changes focused and atomic
- Write clear commit messages
- Include tests for new functionality
- Update documentation if needed
- Ensure all tests pass before submitting

## Reporting Issues

- Use the issue tracker
- Provide detailed reproduction steps
- Include environment information
- Add relevant code samples or logs

## Code of Conduct

Be respectful and professional in all interactions.
`

	contributingPath := filepath.Join(targetPath, "CONTRIBUTING.md")
	return os.WriteFile(contributingPath, []byte(content), 0600)
}

// GenerateLicense generates a LICENSE file
func (g *DocGenerator) GenerateLicense(targetPath, licenseType, author, year string) error {
	var content string

	switch licenseType {
	case "MIT":
		content = fmt.Sprintf(`MIT License

Copyright (c) %s %s

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`, year, author)

	case "Apache-2.0":
		content = fmt.Sprintf(`Apache License
Version 2.0, January 2004
http://www.apache.org/licenses/

Copyright %s %s

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
`, year, author)

	default:
		return fmt.Errorf("unsupported license type: %s", licenseType)
	}

	licensePath := filepath.Join(targetPath, "LICENSE")
	return os.WriteFile(licensePath, []byte(content), 0600)
}

// GenerateArchitectureDoc generates architecture documentation
func (g *DocGenerator) GenerateArchitectureDoc(targetPath string, components []Component) error {
	var sb strings.Builder

	sb.WriteString("# Architecture\n\n")
	sb.WriteString("## Overview\n\n")
	sb.WriteString("This document describes the system architecture.\n\n")
	sb.WriteString("## Components\n\n")

	for _, comp := range components {
		sb.WriteString(fmt.Sprintf("### %s\n\n", comp.Name))
		sb.WriteString(fmt.Sprintf("%s\n\n", comp.Description))
	}

	sb.WriteString("## Data Flow\n\n")
	sb.WriteString("1. Request received by router\n")
	sb.WriteString("2. Authentication/authorization checked\n")
	sb.WriteString("3. Handler processes request\n")
	sb.WriteString("4. Database operations performed\n")
	sb.WriteString("5. Response returned to client\n\n")

	docsDir := filepath.Join(targetPath, "docs")
	if err := os.MkdirAll(docsDir, 0750); err != nil {
		return err
	}

	archPath := filepath.Join(docsDir, "ARCHITECTURE.md")
	return os.WriteFile(archPath, []byte(sb.String()), 0600)
}

// GenerateDeploymentGuide generates deployment documentation
func (g *DocGenerator) GenerateDeploymentGuide(targetPath string, config DeploymentConfig) error {
	var sb strings.Builder

	sb.WriteString("# Deployment Guide\n\n")
	sb.WriteString(fmt.Sprintf("## Platform: %s\n\n", config.Platform))
	sb.WriteString(fmt.Sprintf("## Database: %s\n\n", config.Database))
	sb.WriteString(fmt.Sprintf("## Environment: %s\n\n", config.Environment))

	if config.Platform == "docker" {
		sb.WriteString("## Docker Deployment\n\n")
		sb.WriteString("```bash\n")
		sb.WriteString("docker build -t myapp .\n")
		sb.WriteString("docker run -p 8080:8080 myapp\n")
		sb.WriteString("```\n\n")
	}

	sb.WriteString("## Environment Variables\n\n")
	sb.WriteString("- `PORT`: Server port (default: 8080)\n")
	sb.WriteString(fmt.Sprintf("- `DB_HOST`: %s host\n", config.Database))
	sb.WriteString("- `DB_PORT`: Database port\n")
	sb.WriteString("- `DB_NAME`: Database name\n")
	sb.WriteString("- `DB_USER`: Database user\n")
	sb.WriteString("- `DB_PASSWORD`: Database password\n\n")

	sb.WriteString("## Health Checks\n\n")
	sb.WriteString("- Health endpoint: `/health`\n")
	sb.WriteString("- Readiness endpoint: `/ready`\n\n")

	docsDir := filepath.Join(targetPath, "docs")
	if err := os.MkdirAll(docsDir, 0750); err != nil {
		return err
	}

	deployPath := filepath.Join(docsDir, "DEPLOYMENT.md")
	return os.WriteFile(deployPath, []byte(sb.String()), 0600)
}

// GenerateAll generates all documentation files
func (g *DocGenerator) GenerateAll(targetPath string, config DocConfig) error {
	// Always generate README
	if err := g.GenerateREADME(targetPath, config.ProjectInfo); err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}

	// Generate optional docs
	if config.GenerateChangelog {
		if err := g.GenerateChangelog(targetPath, config.ProjectInfo.Version); err != nil {
			return fmt.Errorf("failed to generate CHANGELOG: %w", err)
		}
	}

	if config.GenerateContributing {
		if err := g.GenerateContributing(targetPath); err != nil {
			return fmt.Errorf("failed to generate CONTRIBUTING: %w", err)
		}
	}

	if config.License != "" {
		year := time.Now().Format("2006")
		if err := g.GenerateLicense(targetPath, config.License, config.ProjectInfo.Author, year); err != nil {
			return fmt.Errorf("failed to generate LICENSE: %w", err)
		}
	}

	if config.GenerateAPI {
		if err := g.GenerateAPIDoc(targetPath, config.APIEndpoints); err != nil {
			return fmt.Errorf("failed to generate API doc: %w", err)
		}
	}

	if config.GenerateArchitecture {
		if err := g.GenerateArchitectureDoc(targetPath, config.Components); err != nil {
			return fmt.Errorf("failed to generate architecture doc: %w", err)
		}
	}

	if config.GenerateDeployment {
		if err := g.GenerateDeploymentGuide(targetPath, config.DeploymentConfig); err != nil {
			return fmt.Errorf("failed to generate deployment guide: %w", err)
		}
	}

	return nil
}

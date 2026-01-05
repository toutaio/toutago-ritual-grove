# Contributing to Ritual Grove

Thank you for your interest in contributing to Ritual Grove! This document provides guidelines and information for contributors.

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How to Contribute

### Reporting Issues

- Use the GitHub issue tracker
- Check if the issue already exists
- Provide detailed information:
  - Go version
  - Operating system
  - Steps to reproduce
  - Expected vs actual behavior

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Write or update tests
5. Ensure tests pass (`go test ./...`)
6. Ensure code is formatted (`go fmt ./...`)
7. Run linter (`golangci-lint run`)
8. Commit with conventional commit format
9. Push to your fork
10. Open a Pull Request

### Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `perf`: Performance improvement
- `refactor`: Code restructuring
- `test`: Adding or updating tests
- `docs`: Documentation changes
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

**Scopes:**
- `ritual`: Ritual manifest and loading
- `generator`: File generation
- `questionnaire`: Interactive prompts
- `registry`: Ritual registry and discovery
- `deployment`: Deployment features
- `cli`: Command-line interface

### Development Setup

1. Clone the repository:
```bash
git clone https://github.com/toutaio/toutago-ritual-grove.git
cd toutago-ritual-grove
```

2. Install dependencies:
```bash
go mod download
```

3. Run tests:
```bash
go test ./...
```

4. Run linter:
```bash
golangci-lint run
```

## Testing Guidelines

- Write tests for all new features
- Maintain or improve code coverage (target: >70%)
- Run tests with race detector: `go test -race ./...`
- Test on multiple Go versions (1.22+)

## Code Quality

- Follow Go conventions and idioms
- Use `gofmt` and `goimports`
- Add godoc comments for public APIs
- Keep functions focused and maintainable
- Handle errors explicitly

## Documentation

- Update README.md for user-facing changes
- Add godoc comments for public APIs
- Update CHANGELOG.md following [Keep a Changelog](https://keepachangelog.com/)
- Include examples for new features

## Questions?

Feel free to open an issue for questions or clarifications about contributing.

Thank you for contributing to Ritual Grove! 🌳

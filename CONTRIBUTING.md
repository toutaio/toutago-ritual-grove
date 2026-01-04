# Contributing to Ritual Grove

Thank you for your interest in contributing to Ritual Grove!

## Development Setup

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

## Project Structure

- `cmd/ritual/` - CLI entry point
- `internal/` - Internal packages (not for public use)
- `pkg/ritual/` - Public API
- `rituals/` - Built-in ritual definitions
- `examples/` - Example rituals
- `docs/` - Documentation

## Making Changes

1. Create a new branch for your feature
2. Write tests for new functionality
3. Ensure all tests pass
4. Run the linter and fix any issues
5. Update documentation if needed
6. Submit a pull request

## Code Style

- Follow standard Go conventions
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small

## Testing

- Write unit tests for all new functionality
- Aim for >80% code coverage
- Include integration tests for complex features

## Commit Messages

Keep commit messages brief and descriptive:
- Use present tense ("Add feature" not "Added feature")
- Reference issues when applicable
- Keep first line under 72 characters

## Questions?

Open an issue or discussion on GitHub.

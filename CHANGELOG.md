# Changelog

All notable changes to the Toutago Ritual Grove project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Core ritual engine with YAML manifest parsing
- File generator with template rendering (Fíth and Go template support)
- Interactive questionnaire system with multiple question types
- Ritual registry with local and Git-based ritual discovery
- CLI integration as plugin for main touta binary
- Ritual validation with JSON schema support
- Migration system for ritual updates
- Deployment utilities (health checks, rollback, update detection)
- Hook system (pre/post install, update)
- Dependency resolution and circular dependency detection
- Comprehensive test suite with >80% coverage across all packages
- Built-in rituals: minimal, hello-world, basic-site
- Documentation: README, CONTRIBUTING, ritual format guide

### Architecture
- Plugin-based design: ritual-grove integrates into touta binary
- Commands accessible via `touta ritual <command>`
- Supports local rituals in `./rituals/` and `~/.touta/rituals/`
- Extensible template engine interface with Fíth integration

### Commands
- `touta ritual init <name>` - Initialize project from ritual
- `touta ritual list` - List available rituals
- `touta ritual info <name>` - Show ritual details
- `touta ritual validate` - Validate ritual.yaml
- `touta ritual create <name>` - Create new ritual template

### Testing
- Unit tests for all core packages
- Integration tests for end-to-end workflows
- Test coverage tracking with coverage reports
- GitHub Actions CI/CD pipeline

## [0.1.0] - 2026-01-05

### Added
- Initial release of ritual-grove
- Basic ritual system functionality
- Integration with toutago framework


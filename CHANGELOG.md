# Changelog

All notable changes to the Toutago Ritual Grove project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive test coverage for plan command
  - Test coverage improved from 54.9% to 73.4% in commands package
  - Plan command runPlan function coverage increased from 19.4% to 69.4%
  - Added tests for error conditions, invalid states, and command structure
  - Added test for JSON output not implemented error
- Comprehensive test coverage for pkg/cli package
  - Test coverage improved from 34.1% to 74.4%
  - Added tests for listRituals, showRitualInfo, searchRituals, and initRitual functions
  - Added tests for valid and invalid ritual operations
  - Added tests for search with and without results
- Comprehensive test coverage for update command
  - Update command helper functions now have 80%+ coverage
  - Added tests for createBackup, loadNewRitual, saveUpdatedState
  - Added tests for parseVersions with valid and invalid inputs
  - Added tests for displayUpdateInfo with breaking and non-breaking updates

## [0.3.0] - 2026-01-05

### Added
- **Update command handler**: Update projects to new ritual versions
  - Version-to-version migration support
  - Automatic backup creation before updates  
  - Rollback on migration failure
  - Dry-run mode for preview
  - Breaking change detection
- **Search command**: Search for rituals by name, tag, or description
- **Update CLI command**: `touta ritual update --to <version>` for updating projects
- **Migrate CLI command**: `touta ritual migrate --up/--down` for running migrations
- Etymology section in README explaining name origins and philosophy

### Changed
- Improved test coverage for pkg/cli package (16.9% → 34.2%)
- Enhanced command structure validation tests
- **Refactored update handler**: Reduced cyclomatic complexity from 16 to <10
  - Split Execute method into smaller, focused helper methods
  - Improved code maintainability and readability
  - All tests pass without changes to behavior

## [0.2.0] - 2026-01-05

### Added
- **Blog ritual**: Full-featured blog with posts, categories, and comments
  - CRUD operations for posts
  - Category organization  
  - Optional comment system with moderation
  - Optional Markdown support
  - Database migrations included
  - Multi-database support (PostgreSQL/MySQL)
  - RESTful API endpoints
  - Responsive CSS styling
  - Comprehensive documentation

### Changed
- Enhanced ritual catalog with domain-specific examples

## [0.1.0] - 2026-01-05

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
- Built-in rituals: minimal, basic-site, hello-world
- Documentation: README, CONTRIBUTING, ritual format guide
- Environment variable `TOUTA_RITUALS_PATH` for custom ritual paths

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

### Templates
- Updated basic-site ritual to use current Cosan Context API
- Updated basic-site ritual to use current Fíth renderer configuration
- Handlers now use `cosan.Context` instead of `http.ResponseWriter/Request`
- Renderer integrated via `cosan.WithRenderer()` functional option

### Testing
- Unit tests for all core packages
- Integration tests for end-to-end workflows  
- Test coverage >80% across all packages
- GitHub Actions CI/CD pipeline

### Fixed
- Ritual templates now match current Cosan v0.1.0+ API
- Ritual templates now match current Fíth v0.1.0+ API


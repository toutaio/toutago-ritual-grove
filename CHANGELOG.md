# Changelog

All notable changes to the Toutago Ritual Grove project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Comprehensive Task Documentation**
  - Complete hook tasks reference guide covering all 30+ available tasks
  - Usage examples for each task type
  - Error handling and custom task creation guidance
- **System Operation Tasks** (fully tested, 75.2% coverage):
  - wait-for-service - Wait for HTTP services or TCP ports to become available with timeout and retry
  - notify - Send log or webhook notifications during ritual execution
- Full Stack (Inertia + Vue) ritual with complete scaffolding
- Template files for Vue 3 with Composition API
- Vite configuration for fast development
- TypeScript support with type generation
- Layout and page component templates
- Optional authentication scaffolding
- Optional SSR support configuration
- **Hook Tasks System Extensions**
  - template-render task for rendering Go templates to files
  - validate-files task for validating file existence
  - go-mod-download task for downloading Go dependencies
  - go-build task for building Go binaries with custom output paths
  - go-test task for running Go tests with custom arguments
  - go-fmt task for formatting Go code
  - go-run task for running Go programs with arguments and environment variables
  - exec-go task for running arbitrary Go commands
  - **HTTP Operation Tasks** (fully tested):
    - http-get - Send HTTP GET requests with custom headers
    - http-post - Send HTTP POST requests with body and headers
    - http-download - Download files from URLs
    - http-health-check - Perform health checks with configurable retries and delays
  - **Validation Tasks** (fully tested):
    - validate-go-version - Check if Go version meets minimum requirement
    - validate-dependencies - Verify required commands are available
    - validate-config - Check if configuration file exists and is valid
    - env-check - Validate required environment variables are set
    - port-check - Check if a port is available

### Fixed
- Fixed golangci-lint configuration format for latest version compatibility

## [0.3.0] - 2026-01-06

### Added
- **Comprehensive Example Applications**
  - Full blog application with Inertia.js and Vue 3 example
  - Admin panel example with dashboard, user management, and analytics
  - Real-time chat application example with WebSocket support
  - HTMX progressive enhancement example with modern patterns
- **Comprehensive Testing**
  - Blog ritual integration tests for all frontend types (Inertia-Vue, HTMX, traditional)
  - Test coverage for frontend choice question validation
  - Ritual structure validation tests
- **Hook Tasks System**
  - env-set task for managing environment variables in .env files
  - Support for quoted values in environment files
  - Automatic file creation and key updates
- HTMX templates for blog ritual with interactive features
- HTMX app.js with notification system and loading indicators
- HTMX layout templates with navigation
- Frontend build configuration for HTMX projects

### Changed
- **Configuration Flexibility**
  - Inertia middleware now uses configured port and host from ritual variables instead of hardcoded values
  - Makefile docker-run command now uses configured port dynamically
  - File permissions changed from 0644 to 0600 for improved security
- **Build System Enhancements**
  - Enhanced esbuild configuration with production optimization
  - Added SSR build support with separate entry point
  - Improved build output with size analysis and better logging
  - Added code splitting for better performance
  - New npm scripts: `build:ssr`, `watch`, and `types` for TypeScript generation

### Added
- **Documentation**
  - Comprehensive Inertia.js integration guide with examples
  - HTMX integration guide covering patterns and best practices
  - TypeScript type generation documentation
  - SSR setup and troubleshooting guides
  - Frontend migration guide for switching between Traditional, Inertia.js, and HTMX
  - Ritual creation guide with examples and best practices
- **Conditional File Generation**
  - Added condition evaluation support for template and static files
  - Files can be conditionally generated based on ritual variables
  - Support for complex conditions using Go template syntax
  - Comprehensive test coverage for condition evaluation
- **Inertia.js Hook Tasks**
  - SetupInertiaMiddleware task for adding Inertia middleware to main.go
  - AddInertiaHandlers task for generating Inertia-compatible CRUD handlers
  - AddSharedData task for configuring shared data functions
  - GenerateTypeScriptTypes task for auto-generating TypeScript types from Go structs
  - UpdateRoutesForInertia task for adding Inertia routes
  - Complete test coverage for all Inertia hook tasks
- **Inertia.js Frontend Integration**
  - Frontend framework choice question (traditional, inertia-vue, htmx)
  - SSR configuration option for Inertia.js
  - Conditional question support based on frontend type
  - Tests for frontend integration schema and blog ritual
  - Complete Inertia-Vue templates for blog ritual:
    - Vue page components (Home, Posts/Index, Posts/Show, Posts/Edit)
    - Vue layout components (Layout, Header, Footer)
    - Frontend build configuration (esbuild, package.json)
    - Client-side app entry point with SSR support
  - Conditional file generation based on frontend_type
  - Integration with @toutaio/inertia-vue package
- **Declarative Hook Task System** (Cross-Platform)
  - Task interface and context for cross-platform hook execution
  - Task registry system for registering and creating tasks
  - File operation tasks (mkdir, copy, move, remove, chmod)
  - Comprehensive test coverage for task system (100%)
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

### Changed
- Hooks system now supports declarative tasks (replacing script-based hooks)

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


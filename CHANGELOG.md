# Changelog

All notable changes to the Toutago Ritual Grove project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Blog Ritual: Post Management Service (Phase 2.3)** (TDD)
  - **PostService Interface & Implementation**:
    - Full CRUD operations with permission checks
    - Methods: `Create()`, `Update()`, `Delete()`, `Publish()`, `Archive()`
    - Automatic slug generation from titles
    - Author ownership validation
    - Status management (draft, published, archived)
  - **Post DTOs**:
    - `CreatePostDTO` and `UpdatePostDTO` with validation
    - `PostFilters` for querying and pagination
    - `CreateCategoryDTO` and `UpdateCategoryDTO`
  - **Repository Interfaces**:
    - `PostRepository` with comprehensive query methods
    - `CategoryRepository` for category management
    - Support for filtering by author, category, status
  - **Slug Generation**:
    - URL-friendly slug generator utility
    - Automatic uniqueness handling
  - **Comprehensive Testing**:
    - Mock repository for isolated tests
    - Permission-based test scenarios
    - Ownership and access control tests
    - Edge cases coverage

- **Blog Ritual: User Management Service (Phase 2.2)** (TDD)
  - **UserService Interface**:
    - CRUD operations for user management
    - Role-based authorization for all operations
    - Methods: `GetByID()`, `GetByEmail()`, `List()`, `Update()`, `Delete()`, `ChangeRole()`, `ToggleActive()`
  - **UserService Implementation**:
    - Permission checks using PermissionService
    - Prevent self-deletion and self-role-change
    - Email and username uniqueness validation
    - Support for user filtering and pagination
  - **User DTOs**:
    - `UpdateUserDTO` with validation
    - `UserFilters` for querying users
  - **Comprehensive Testing**:
    - Mock repository for isolated testing
    - Test cases for all CRUD operations
    - Permission enforcement tests
    - Edge case coverage (self-deletion, role changes)

- **Blog Ritual: Authorization & RBAC (Phase 2.1)** (TDD)
  - **Domain Models**:
    - `Post` entity with validation, status management, and ownership checks
    - `Category` entity with validation
    - Post status enums: draft, published, archived
    - Domain methods: `CanBeEditedBy()`, `CanBeDeletedBy()`, `IsPublished()`
  - **Permission Service**:
    - `PermissionService` interface for authorization checks
    - Complete implementation with role-based logic
    - Fine-grained permission checks: `Can()`, `CanEditPost()`, `CanDeletePost()`, `CanPublishPost()`
    - Category and user management permission checks
    - Action-based permission system (post.view, post.edit, etc.)
  - **Role Middleware**:
    - `RoleMiddleware()` for dynamic role checking
    - `RequireAdmin()`, `RequireEditor()`, `RequireAuthor()` helpers
    - Proper error responses for unauthorized access
  - **Comprehensive Testing**:
    - Post domain tests (validation, status checks, ownership)
    - Category domain tests
    - Permission service tests (120+ test cases)
    - Role middleware tests
    - All tests follow TDD approach (tests written first)

- **Blog Ritual: Auth UI Templates (Phase 1.4-1.5)**
  - **SSR Templates**:
    - Login page with form validation and error display
    - Register page with password confirmation
    - Setup page for first-user admin creation
    - Consistent styling and mobile-responsive design
  - **Inertia.js Vue Components**:
    - Login.vue with reactive form handling
    - Register.vue with client-side validation hints
    - Setup.vue with admin account creation
    - Integrated with Inertia router for SPA navigation
    - Loading states and error handling
  - **Features**:
    - Client-side form validation (minlength, required)
    - Error message display
    - Processing states (disabled buttons while submitting)
    - Consistent auth card layout
    - Responsive design for all screen sizes

- **Blog Ritual: Database Repositories & Auth Handlers (Phase 1.3 completion)**
  - **PostgreSQL Repositories**:
    - `UserRepository` implementation with full CRUD operations
    - `SessionRepository` implementation with token management
    - Support for user queries by email, username, ID
    - First-user detection for admin setup
    - Session expiration handling
  - **MySQL Repositories**:
    - `UserRepository` implementation (MySQL parameter style)
    - `SessionRepository` implementation (MySQL parameter style)
    - Feature parity with PostgreSQL implementations
  - **Auth Handlers**:
    - `AuthHandler` with login, register, logout, setup methods
    - HTTP-only cookie-based session management
    - Auto-login after registration
    - First-user admin setup flow
    - Proper redirects based on authentication state
    - Support for both SSR and Inertia.js frontends

- **Blog Ritual: Authentication Integration (Phase 1.3)** (TDD)
  - **Domain Models**:
    - `User` domain model implementing Breitheamh auth interfaces
    - `Session` domain model with validation and expiration checking
    - Comprehensive unit tests for User and Session models (100% coverage)
  - **Data Transfer Objects**:
    - `RegisterDTO` with email, username, and password validation
    - `LoginDTO` with credentials and request metadata
    - `PasswordResetDTO` and `PasswordResetRequestDTO`
    - `ChangePasswordDTO` with old/new password validation
    - Complete test coverage for all DTOs
  - **Repository Interfaces**:
    - `UserRepository` interface with CRUD and query methods
    - `SessionRepository` interface with token and expiration management
    - `UserFilters` for advanced user queries
  - **Services**:
    - `AuthService` interface for authentication operations
    - `AuthService` implementation wrapping toutago-breitheamh-auth
    - Support for registration, login, logout, session verification
    - First-user admin auto-promotion
    - Password change functionality
    - Comprehensive unit tests with mocked repositories (90%+ coverage)
  - **Middleware**:
    - `AuthMiddleware` requiring authentication with session cookies
    - `OptionalAuthMiddleware` for mixed public/private routes
    - `GuestMiddleware` redirecting authenticated users from login pages
    - `SetupMiddleware` redirecting to first-user setup if needed
    - Complete test coverage for all middleware functions
  - **Dependencies**:
    - Added `github.com/toutaio/toutago-breitheamh-auth v0.1.0`
    - Added `github.com/google/uuid v1.5.0`
    - Updated go.mod.tmpl with authentication dependencies

- **Blog Ritual: Complete Authentication System (Phase 1.1-1.2)**
  - **Domain Models**:
    - `User` model with authentication fields (email, username, password, role, status, email verification)
    - `Session` model for session management with expiration tracking
    - `VerificationToken` model for email verification and password reset flows
    - `Role` type: Admin, Editor, Author, Reader
    - `UserStatus` type: Active, Inactive, Locked
    - `Permission` constants for granular access control (29 permissions across posts, categories, users, comments, tags, media, settings)
    - `RolePermissions` matrix mapping roles to permissions
    - Comprehensive test coverage (user_test.go, session_test.go, permission_test.go)
  - **Database Migrations**:
    - Migration 001: `users` table with full authentication support
    - Migration 002: `sessions` table with token management
    - Migration 003: `verification_tokens` table for email/password flows
    - Migration 004: Updated `posts` table with author_id FK and soft deletes
    - Migration 005: `tags` and `post_tags` tables for WordPress-style tagging
    - Migration 006: `media` table for S3/cloud storage tracking
    - Migration 007: `webhooks` and `webhook_deliveries` tables for event notifications
    - All migrations include proper indexes, foreign key constraints, and up/down scripts
    - PostgreSQL triggers for automatic `updated_at` timestamps

- **Docker Support**
  - Added shared Docker templates in `rituals/_shared/docker/`
  - Dockerfile.go.tmpl with Air for hot reload
  - docker-compose.yml.tmpl with conditional database services (PostgreSQL/MySQL)
  - .dockerignore.tmpl with common exclusions
  - .air.toml.tmpl for Go hot reload configuration
  - .env.example.tmpl for environment variables
  - wait-for-it.sh script for database readiness
  - Shared frontend templates (package.json.tmpl, esbuild.config.js.tmpl)
  - DOCKER.md.tmpl comprehensive user documentation
  - DATABASE.md.tmpl comprehensive database setup guide
  - Comprehensive test suite (45+ test cases) for Docker functionality
  - Generator support for `_shared:` template prefix
  - Conditional file generation based on ritual answers
  - Docker support added to all rituals: minimal, hello-world, basic-site, blog, wiki, fullstack-inertia-vue
  - Automatic extraction of `_shared` directory from embedded rituals
  - **Automatic .env file generation** - Ready to use without manual copying!

### Changed
- Updated generator to support shared templates via `_shared:` prefix
- Updated scaffolder to evaluate file conditions
- Updated CLI to set rituals base path for shared template access
- Updated registry to extract `_shared` directory alongside rituals
- Fixed template and static file path consistency across all rituals
- Upgraded Docker base image to Go 1.23 for compatibility
- Removed obsolete version field from docker-compose.yml
- Updated Air installation to use air-verse/air (latest fork)

## [0.6.1-dev] - 2026-01-08

### Fixed
- Fixed problems with the templates for blog

## [0.6.0-dev] - 2026-01-08

### Added
- **Inertia.js Integration in Blog Ritual**
  - Handlers now properly use Inertia.js when `frontend_type` is `inertia-vue`
  - Added Category Vue components (Index.vue and Show.vue)
  - Handlers render Inertia pages instead of JSON for Inertia frontend
  - Main.go template initializes Inertia instance and passes to handlers
  - Added CosanAdapter to bridge cosan.Context with inertia.InertiaContext

### Fixed
- **Blog Ritual Inertia.js API Usage**
  - Fixed handler templates to use correct InertiaContext methods
  - Changed from incorrect `h.inertia.Render(w, r, ...)` to proper `ic.Render(...)`
  - Changed from `h.inertia.Location(w, url)` to `ic.Redirect(url)` and `ic.Location(url)`
  - Handlers now create InertiaContext using adapter pattern
  - Supports SSR configuration when enabled

### Fixed
- **Template Path Resolution**
  - Fixed generator to correctly resolve templates from `templates/` subdirectory
  - Ritual templates now load properly after cache cleanup
- **Blog Ritual File Conditions**
  - Fixed file conditions to use `[[ ]]` delimiters instead of `{{ }}`
  - Correctly generates frontend files based on selected framework (inertia-vue, htmx, traditional)
  - Vue/Inertia.js files now properly generated when selected
- **Blog Ritual Templates**
  - Fixed handler templates to properly import nasc package
  - Fixed go.mod template conditional syntax with correct `[[ ]]` delimiters
  - Fixed Vue template syntax to avoid Go template conflicts
  - Removed Makefile templates (not used in Toutā projects)
  - Updated handlers to conditionally use Inertia or JSON based on frontend type
  - Fixed package import paths (removed incorrect `/pkg/` subdirectories)

### Changed
- **Template Delimiter Change**
  - Changed template delimiters from `{{ }}` to `%% %%` for Vue/JS template files
  - Changed general template delimiters from `{{ }}` to `[[ ]]` for other files
  - Prevents conflicts with Vue.js, React, and other JavaScript frameworks
  - All ritual templates updated
  - All tests updated
  - More readable templates when mixing Go templates with frontend frameworks

### Added
- **Improved CLI Questionnaire UX**
  - Numbered menu selection for choice questions
  - Automatic quote stripping from user input
  - Support for both number (1-based) and name selection
  - Retry on invalid input with clear error messages
- **Frontend Build Instructions**
  - Post-initialization message shows npm install/build steps for inertia-vue and htmx
  - README includes frontend-specific setup instructions
  - Project structure documentation adapted for each frontend type
- **Ritual Cache Management**
  - Added `ritual clean` command to clear cached rituals
  - Added `--force` flag to skip confirmation prompt
  - Helps resolve issues when rituals appear outdated after upgrades
  - Comprehensive tests for quote handling and numbered choices
- **Blog Ritual Handlers**
  - Updated handlers to use Cosán Context API
  - PostHandler, CategoryHandler, CommentHandler now use `cosan.Context`
  - All handlers return errors properly
  - Simplified request/response handling with Context methods

### Fixed
- Choice questions now accept quoted values (e.g., `"inertia-vue"`)
- Invalid choice numbers now properly retry instead of failing
- CLI adapter now handles conversion errors gracefully
- **Blog Ritual API Compatibility**
  - Fixed handler templates to use `nasc.Nasc` instead of `nasc.Container`
  - Fixed main.go to use `nasc.New()` instead of `nasc.NewContainer()`
  - Fixed renderer initialization to use `fith.NewWithDir()` instead of `fith.NewRenderer()`
  - Fixed router initialization to use `cosan.New()` instead of `cosan.NewRouter()`
  - Fixed handlers to use `c.Bind()` instead of `c.BindJSON()`
  - Removed non-existent `router.Static()` method, using manual file server instead
  - Blog ritual now generates working, compilable code
- Blog ritual question conditions now use proper structured format instead of template expressions
- Vue template files now correctly separate Go template variables (`[[ ]]`) from Vue interpolations (`{{ }}`)
- All Inertia Vue templates in blog ritual fixed for proper rendering
- SSR question now appears correctly when Inertia.js frontend is selected

### Added (Previous)
- **Cache Management System**
  - `ritual clean` command to clear ritual cache
  - Automatic version checking for embedded rituals
  - Re-extract rituals when embedded version changes
  - Cache size reporting
  - Comprehensive cache management tests

### Fixed (Previous)
- Embedded rituals now automatically update when binary is rebuilt
- Old cached rituals no longer persist after recompilation

### Changed
- Renamed internal registry field from `cache` to `rituals` for clarity

## [0.5.0] - 2026-01-07
- **Semantic Versioning Utilities (Task 8.2.3)**
  - Full semver 2.0 support
  - Parse and compare versions
  - Check version compatibility
  - Detect breaking changes
  - Generate next versions
  - 6 comprehensive test suites (all passing)

- **JSON Output Support (Task 15.4.8.7)**
  - `ritual plan --json` flag
  - Machine-readable structured output
  - Includes all plan details
  - Easy integration with CI/CD
  - 2 comprehensive tests (all passing)

- **Enhanced Ritual Template Creation (Task 19.5)**
  - `ritual create --with-examples` flag
  - Generates example Go template
  - Includes comprehensive .gitignore
  - Migration guide with examples
  - YAML with helpful comments
  - README documenting hooks
  - 7 comprehensive tests (all passing)
  - TDD approach followed

- **Deployment Workflow Example (Task 19.4)**
  - Complete deployment workflow guide
  - Step-by-step update process
  - Rollback procedures and decision matrix
  - Migration patterns (4 common patterns)
  - Production best practices
  - Troubleshooting guide
  - Emergency procedures
  - Real-world deployment scenarios

- **Backup Management CLI (Tasks 14.2, 14.3, 14.7)**
  - `ritual backup list` - List all available backups with metadata
  - `ritual backup create` - Create manual backup with description
  - `ritual backup restore` - Restore from specific backup
  - `ritual backup clean` - Clean old backups with retention policy
  - Automatic backups before updates (already existed)
  - Human-readable size formatting
  - Comprehensive tests (6 test cases, all passing)

- **Comprehensive Documentation (Tasks 18.5, 18.6, 18.10)**
  - Complete CLI command reference (9 commands documented)
  - All flags, options, and usage examples
  - Common workflows and exit codes
  - Best practices guide covering design, security, performance
  - Versioning and testing strategies
  - Anti-patterns to avoid
  - CHANGELOG following Keep a Changelog format

- **Enhanced Ritual Validation (Tasks 17.1-17.3)**
  - File reference validation ensures templates/static files exist
  - Version constraint validation (min < max checks)
  - Question conditional logic validation
  - Circular condition dependency detection
  - Common mistake warnings (unprotected configs, missing tests)
  - Migration reversibility checks
  - 6 new comprehensive test suites (all passing)
  - TDD approach throughout

- **Example Ritual Documentation (Task 19.1)**
  - Comprehensive README for simple-app example
  - Comprehensive README for minimal-app example
  - Shows questionnaire design, conditional questions, templates
  - Includes customization guides and usage examples
  - Comparison table between examples
  - Template syntax reference and best practices

- **State Checkpoint System (Task 9.7)**
  - New `CheckpointManager` for point-in-time state restoration
  - Create checkpoints with custom labels before risky operations
  - Restore project state from any checkpoint
  - Automatic cleanup of old checkpoints (configurable retention)
  - Label-based checkpoint retrieval
  - Comprehensive test coverage (7 test cases, all passing)
  - Checkpoints stored in `.ritual/checkpoints/` as JSON

- **Comprehensive Ritual Integration Tests**
  - New end-to-end tests for all built-in rituals
  - Verifies ritual loading, project generation, compilation, and tests
  - TDD approach identified real issues with import paths and templates
  - Automatic validation of all 6 built-in rituals (minimal, hello-world, basic-site, blog, wiki, fullstack-inertia-vue)
  - Tests verify generated projects compile successfully with `go build`

- **Protected File Management Integration**
  - Integration tests for protected file loading from state and user files
  - Comprehensive pattern matching tests (*.env, config/*.yaml)
  - Documentation of integration with deployment system
  - All protected file tests passing with 72.5% storage coverage

- **Ritual Documentation**
  - Comprehensive README for all 6 built-in rituals
  - Usage examples, configuration options, features, and limitations documented
  - Architecture and customization guides included
  - Clear getting started instructions for each ritual

### Changed
- Tasks 7.10, 7.11, 8.10, 8.11, 9.7, 9.8, 9.9, 9.10, 19.1 marked as complete
- Dry-run mode already fully implemented (marked complete)
- Automatic backups and point-in-time restore verified and marked complete
- Example ritual documentation significantly enhanced
- Overall test coverage maintained at 75%+ across all packages
- Code quality verified with golangci-lint

- **Deployment History Tracking**
  - New `DeploymentHistory` system tracks all deployment attempts
  - Records timestamp, versions, status (success/failure/rollback), errors, warnings
  - Saved to `.ritual/history.yaml` with automatic size limiting (max 100 entries)
  - API methods: `GetLatestSuccessful()`, `GetFailures()`, `GetRollbacks()`
  - Comprehensive test coverage (7 test cases)

- **Protected File Management**
  - New `ProtectedFileManager` for managing files that should not be overwritten
  - Support for exact file paths and glob patterns (`*.env`, `config/*.yaml`)
  - User-defined protected files via `.ritual/protected.txt`
  - Pattern matching with `filepath.Match()` for flexible protection rules
  - API methods: `IsProtected()`, `AddProtectedFile()`, `RemoveProtectedFile()`
  - Comprehensive test coverage (8 test cases)

- **Declarative Task System Integration**
  - Hook executor now supports both shell commands AND JSON task objects
  - Automatic detection and routing of task objects vs shell commands
  - Seamless mixing of shell commands and tasks in hook arrays
  - Comprehensive test suite for task execution (11 new tests)
  - Task validation and error handling
  - Examples:
    - `{"type": "mkdir", "path": "/tmp/dir", "perm": 0755}`
    - `{"type": "go-mod-tidy"}`
    - `{"type": "copy", "src": "file.txt", "dest": "backup.txt"}`
- **go-mod-tidy Task Registration**
  - Added missing go-mod-tidy task to task registry
  - Enables declarative go mod tidy operations in hooks
  
### Changed
- Storage package coverage improved to 72.5%
- Hook executor enhanced to detect JSON task objects automatically
- Task execution integrated into all hook phases (pre/post install/update/deploy)
- Improved hook validation to check both shell commands and task objects

### Fixed
- go-mod-tidy task was defined but not registered in task system
- Hook executor now properly creates TaskContext with working directory and environment

## [0.5.1] - 2026-01-06

### Added
- README documentation for hello-world ritual
- Comprehensive test coverage for dbops package (39.5% → 80.3%)
- Comprehensive test coverage for validationops package (59.1% → 92.9%)

### Changed
- Code quality improvements
  - Fixed error checking in inertia tasks
  - Fixed lint warnings (prealloc, errcheck, lll)
  - Improved function signature formatting
  - Reduced lint issues from 434 to 419
  - Fixed port-check task registration to handle float64 values
  
### Fixed
- Unchecked error in fmt.Sscanf call
- Line length violations in command handlers
- Inefficient slice allocations
- Port validation in task registration system

## [0.5.0] - 2026-01-06

### Added
- **Ritual List Filtering** (`touta ritual list --tag --name --author`)
  - Filter rituals by tags (OR logic - matches any tag)
  - Filter by name pattern (case-insensitive substring match)
  - Filter by author name
  - Display tags in list output
  - Combined filtering support
  - Examples:
    - `touta ritual list --tag web,api`
    - `touta ritual list --name blog`
    - `touta ritual list --tag web --author john`
- **Config File Support** (`touta ritual init --config`)
  - Load answers from YAML or JSON configuration files
  - Supports `.yaml`, `.yml`, and `.json` formats
  - Eliminates need for interactive prompts
  - Example: `touta ritual init blog --config answers.yaml`
- **Git Repository Initialization** (`touta ritual init --git`)
  - Automatically initialize git repository after project creation
  - Creates initial commit with all generated files
  - Gracefully handles missing git or configuration

### Changed
- **Improved Test Coverage**
  - internal/registry: 83.3% → 83.8% (+0.5%)
  - pkg/cli: 65.4% → 68.2% (+2.8%)
  - fileops package: 50.9% → 65.3% (+14.4%)
  - goops package: 58.3% → 64.3% (+6.0%)
  - internal/cli: 80.0% → 79.7% (new features)
  - Overall coverage: 74.7% → 75.2% (+0.5%)
  - Added comprehensive Name() and Validate() tests for all task types

### Summary
Complete CLI enhancement release with filtering, configuration files, and git integration.
All practical CLI features (sections 15.3.x and 15.4.x) are now complete.

## [0.4.0] - 2026-01-06

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

### Metrics
- Overall test coverage: 74.7%
- All core packages above 70% coverage
- Comprehensive testing for all task types

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


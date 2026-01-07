# CLI Command Reference

Complete reference for all Toutago Ritual Grove commands.

## Table of Contents

- [Global Flags](#global-flags)
- [ritual init](#ritual-init) - Initialize project from ritual
- [ritual list](#ritual-list) - List available rituals
- [ritual info](#ritual-info) - Show ritual details
- [ritual validate](#ritual-validate) - Validate ritual
- [ritual create](#ritual-create) - Create new ritual
- [ritual plan](#ritual-plan) - Preview deployment changes
- [ritual search](#ritual-search) - Search for rituals
- [ritual update](#ritual-update) - Update ritual version
- [ritual migrate](#ritual-migrate) - Run migrations

## Global Flags

These flags work with all commands:

- `--help`, `-h` - Show help information
- `--version` - Show version information

## ritual init

Initialize a new project from a ritual template.

### Usage

```bash
ritual init <ritual-name> [flags]
```

### Arguments

- `ritual-name` - Name of the ritual to use (required)

### Flags

- `--output`, `-o` - Output directory (default: current directory)
- `--yes` - Skip questions and use defaults
- `--git` - Initialize git repository after creation
- `--config`, `-c` - Load answers from config file (YAML or JSON)

### Examples

**Basic initialization:**
```bash
ritual init blog
```

**Specify output directory:**
```bash
ritual init blog --output ./my-blog
```

**Skip interactive questions:**
```bash
ritual init blog --yes --output ./quick-blog
```

**Use configuration file:**
```bash
# Create answers.yaml
cat > answers.yaml << EOF
project_name: my-awesome-blog
database: postgres
port: 8080
EOF

ritual init blog --config answers.yaml --output ./my-blog
```

**Initialize with git:**
```bash
ritual init blog --output ./my-blog --git
```

### What It Does

1. Loads the specified ritual
2. Asks questionnaire (unless `--yes` specified)
3. Generates files from templates
4. Installs Go dependencies
5. Runs post-install hooks
6. Initializes git if `--git` specified

### Output

Creates project structure based on ritual configuration:
```
my-blog/
├── .ritual/          # Ritual state and metadata
│   ├── state.yaml    # Current ritual version and state
│   └── ritual.yaml   # Copy of ritual manifest
├── main.go           # Generated application code
├── go.mod            # Go module
├── handlers/         # HTTP handlers
├── models/           # Data models
└── views/            # Templates
```

## ritual list

List all available rituals from configured sources.

### Usage

```bash
ritual list [flags]
```

### Flags

- `--tag` - Filter by tag (can specify multiple)
- `--name` - Filter by name pattern
- `--author` - Filter by author

### Examples

**List all rituals:**
```bash
ritual list
```

**Filter by tag:**
```bash
ritual list --tag blog
ritual list --tag blog --tag cms
```

**Filter by name:**
```bash
ritual list --name blog
```

**Filter by author:**
```bash
ritual list --author "Toutā Team"
```

### Output Format

```
Available Rituals:

  blog (v1.0.0)
    A full-featured blog with posts, comments, and categories
    Tags: blog, content, cms
    Author: Toutā Team

  wiki (v1.0.0)
    Knowledge base with version control and search
    Tags: wiki, knowledge-base, documentation
    Author: Toutā Team

  ...
```

## ritual info

Show detailed information about a specific ritual.

### Usage

```bash
ritual info <ritual-name> [flags]
```

### Arguments

- `ritual-name` - Name of the ritual (required)

### Flags

- `--version` - Show specific version (default: latest)

### Examples

**Show ritual info:**
```bash
ritual info blog
```

**Show specific version:**
```bash
ritual info blog --version 1.0.0
```

### Output

```
Ritual: blog
Version: 1.0.0
Description: A full-featured blog with posts, comments, and categories
Author: Toutā Team
License: MIT
Tags: blog, content, cms

Compatibility:
  Min Toutā Version: 0.1.0
  Go Version: 1.21+

Dependencies:
  - github.com/toutaio/toutago
  - github.com/toutaio/toutago-cosan-router
  - github.com/toutaio/toutago-fith-renderer
  - github.com/toutaio/toutago-datamapper

Questions: 8
  - project_name (text, required)
  - database (choice: postgres, mysql, sqlite, none)
  - port (number, default: 8080)
  ...

Files: 15 templates, 3 static
Migrations: 2
Hooks: 2 post-install
```

## ritual validate

Validate a ritual's structure and configuration.

### Usage

```bash
ritual validate [path] [flags]
```

### Arguments

- `path` - Path to ritual directory (default: current directory)

### Flags

- `--strict` - Enable strict validation mode
- `--check-files` - Verify all referenced files exist

### Examples

**Validate current directory:**
```bash
ritual validate
```

**Validate specific ritual:**
```bash
ritual validate ./my-ritual
```

**Strict validation:**
```bash
ritual validate --strict --check-files
```

### Validation Checks

1. **Schema Validation**
   - Valid `ritual.yaml` structure
   - Required fields present
   - Correct data types

2. **Semantic Validation**
   - Template files exist
   - Version constraints valid
   - No circular dependencies
   - Question conditions valid

3. **Best Practices** (warnings)
   - Config files protected
   - Migrations reversible
   - Tests included

### Output

```
✓ Ritual metadata valid
✓ Compatibility settings valid
✓ All 6 questions valid
✓ All 15 template files found
✓ All 3 static files found
✓ 2 migrations valid

⚠ Warning: config.yaml not in protected files
⚠ Warning: Migration 1.0.0->1.1.0 has no down handler

Validation: PASSED (2 warnings)
```

## ritual create

Create a new ritual template.

### Usage

```bash
ritual create <ritual-name> [flags]
```

### Arguments

- `ritual-name` - Name for the new ritual (required)

### Flags

- `--output`, `-o` - Output directory (default: ./rituals/<name>)
- `--template` - Start from template (minimal, simple, full)

### Examples

**Create minimal ritual:**
```bash
ritual create my-ritual
```

**Use template:**
```bash
ritual create my-ritual --template simple
```

**Specify output:**
```bash
ritual create my-ritual --output ~/rituals/my-ritual
```

### What It Creates

```
my-ritual/
├── ritual.yaml       # Ritual manifest
├── README.md         # Documentation
├── templates/        # Template files
│   └── .gitkeep
└── static/           # Static files
    └── .gitkeep
```

## ritual plan

Preview changes that would be made by updating to a newer ritual version.

### Usage

```bash
ritual plan [flags]
```

### Flags

- `--to-version` - Target version (default: latest)
- `--json` - Output in JSON format

### Examples

**Show update plan:**
```bash
ritual plan
```

**Plan for specific version:**
```bash
ritual plan --to-version 1.2.0
```

**JSON output:**
```bash
ritual plan --json
```

### Output

```
Current ritual: blog v1.0.0
Target ritual:  blog v1.1.0

Changes Summary:
  Files to add:      3
  Files to modify:   5
  Files to delete:   1
  Migrations to run: 1

Detailed Changes:

Files to Add:
  + handlers/api.go
  + views/api/index.html
  + views/api/post.html

Files to Modify:
  ~ main.go
  ~ handlers/post.go
  ~ models/post.go
  ~ go.mod
  ~ README.md

Files to Delete:
  - old_config.yaml

Migrations:
  1.0.0 → 1.1.0: Add API support
    Up: CREATE TABLE api_keys
    Down: DROP TABLE api_keys

Protected Files:
  The following files are protected and won't be overwritten:
    config.yaml
    .env

⚠ Conflicts:
  main.go - Both modified locally and in update
    Resolution: Manual merge required

Use 'ritual update' to apply these changes.
```

## ritual search

Search for rituals by name, tags, or description.

### Usage

```bash
ritual search <query> [flags]
```

### Arguments

- `query` - Search query

### Flags

- `--tag` - Search in tags only
- `--exact` - Exact match only

### Examples

**Search all fields:**
```bash
ritual search blog
```

**Search by tag:**
```bash
ritual search cms --tag
```

**Exact match:**
```bash
ritual search "Full-featured blog" --exact
```

### Output

```
Found 3 matching rituals:

  blog (v1.0.0) ⭐⭐⭐⭐⭐
    A full-featured blog with posts, comments, and categories
    Tags: blog, content, cms

  simple-blog (v1.0.0) ⭐⭐⭐
    Minimal blogging platform
    Tags: blog, minimal

  micro-blog (v0.5.0) ⭐⭐
    Ultra-minimal microblogging
    Tags: blog, micro, twitter-like
```

## ritual update

Update a project to a newer ritual version.

### Usage

```bash
ritual update [flags]
```

### Flags

- `--to-version` - Target version (default: latest)
- `--dry-run` - Show what would change without applying
- `--force` - Skip confirmation prompts
- `--backup` - Create backup before update (default: true)

### Examples

**Update to latest:**
```bash
ritual update
```

**Update to specific version:**
```bash
ritual update --to-version 1.2.0
```

**Dry run:**
```bash
ritual update --dry-run
```

**Force update:**
```bash
ritual update --force
```

### Process

1. Checks current ritual version
2. Finds target version
3. Creates backup
4. Shows change preview
5. Asks for confirmation (unless `--force`)
6. Applies file changes
7. Runs migrations
8. Runs update hooks
9. Updates state

### Safety Features

- Automatic backups before update
- Protected files never overwritten
- Conflict detection and resolution
- Rollback on failure
- Migration reversibility

## ritual migrate

Run ritual migrations manually.

### Usage

```bash
ritual migrate <command> [flags]
```

### Commands

- `up` - Apply pending migrations
- `down` - Rollback last migration
- `status` - Show migration status
- `list` - List all migrations

### Flags

- `--to-version` - Migrate to specific version
- `--dry-run` - Show what would run without executing

### Examples

**Apply pending migrations:**
```bash
ritual migrate up
```

**Rollback last migration:**
```bash
ritual migrate down
```

**Show status:**
```bash
ritual migrate status
```

**List all migrations:**
```bash
ritual migrate list
```

**Dry run:**
```bash
ritual migrate up --dry-run
```

### Output

```
Migration Status:

Applied:
  ✓ 1.0.0 → 1.1.0 (2026-01-05 10:30:00)
  ✓ 1.1.0 → 1.2.0 (2026-01-06 14:20:00)

Pending:
  □ 1.2.0 → 1.3.0 - Add user profiles

Current version: 1.2.0
Latest version:  1.3.0
```

## Common Workflows

### Creating a New Project

```bash
# 1. List available rituals
ritual list

# 2. Get info about a ritual
ritual info blog

# 3. Initialize project
ritual init blog --output my-blog

# 4. Enter project and start development
cd my-blog
go run main.go
```

### Updating an Existing Project

```bash
# 1. Check what would change
ritual plan

# 2. Review the plan

# 3. Apply update
ritual update

# 4. Resolve any conflicts
# (edit conflicting files manually)

# 5. Test the updated project
go test ./...
go run main.go
```

### Creating a Custom Ritual

```bash
# 1. Create ritual structure
ritual create my-custom-ritual

# 2. Edit ritual.yaml
cd my-custom-ritual
vi ritual.yaml

# 3. Add templates
mkdir -p templates
# Add your .tmpl files

# 4. Validate
ritual validate

# 5. Test by using it
ritual init my-custom-ritual --output /tmp/test-project
```

### Validating Before Publishing

```bash
# 1. Validate ritual structure
ritual validate --strict --check-files

# 2. Test generation
ritual init my-ritual --output /tmp/test --yes

# 3. Test generated project compiles
cd /tmp/test
go mod tidy
go build
go test ./...

# 4. Clean up
rm -rf /tmp/test
```

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Invalid arguments
- `3` - Ritual not found
- `4` - Validation failed
- `5` - Conflict detected
- `6` - Migration failed

## Configuration Files

### Answer Configuration (YAML)

```yaml
# answers.yaml
project_name: my-blog
database: postgres
db_host: localhost
db_port: 5432
port: 8080
with_auth: true
```

### Answer Configuration (JSON)

```json
{
  "project_name": "my-blog",
  "database": "postgres",
  "db_host": "localhost",
  "db_port": 5432,
  "port": 8080,
  "with_auth": true
}
```

## Environment Variables

- `RITUAL_PATH` - Additional ritual search path
- `RITUAL_CACHE_DIR` - Cache directory location
- `RITUAL_NO_COLOR` - Disable colored output

## See Also

- [Ritual Format](ritual-format.md) - ritual.yaml specification
- [Creating Rituals](CREATING_RITUALS.md) - Ritual authoring guide
- [Deployment Management](deployment-management.md) - Update and rollback guide
- [README](../README.md) - Project overview

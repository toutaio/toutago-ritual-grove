# Best Practices Guide

Guidelines and best practices for creating and using Toutago rituals.

## Table of Contents

- [Ritual Design Patterns](#ritual-design-patterns)
- [Security Considerations](#security-considerations)
- [Performance Optimization](#performance-optimization)
- [Versioning Strategies](#versioning-strategies)
- [Testing Strategies](#testing-strategies)
- [File Organization](#file-organization)
- [Questionnaire Design](#questionnaire-design)
- [Migration Patterns](#migration-patterns)

## Ritual Design Patterns

### Single Responsibility

Each ritual should serve one clear purpose:

✅ **Good:**
```yaml
# blog ritual - focused on blogging
ritual:
  name: blog
  description: Full-featured blog with posts and comments
```

❌ **Bad:**
```yaml
# kitchen-sink ritual - tries to do everything
ritual:
  name: everything-app
  description: Blog + CRM + E-commerce + Admin panel
```

### Composition Over Inheritance

Build complex rituals by composing simpler ones:

```yaml
# Base ritual
ritual:
  name: web-app-base
  
# Specialized ritual extends base
ritual:
  name: blog
  parent:
    ritual: web-app-base
    version: "^1.0.0"
```

### Default to Secure

Provide secure defaults:

```yaml
questions:
  - name: session_secret
    type: password
    default: ""  # Force user to set
    required: true
    
  - name: csrf_protection
    type: boolean
    default: true  # Secure by default
```

## Security Considerations

### Never Commit Secrets

Always protect sensitive files:

```yaml
files:
  templates:
    - src: .env.template
      dest: .env
      
  protected:
    - .env
    - config/secrets.yaml
    - "*.pem"
    - "*.key"
```

### Validate User Input

Use validation rules:

```yaml
questions:
  - name: database_host
    type: text
    validate:
      pattern: "^[a-zA-Z0-9.-]+$"  # Prevent injection
      
  - name: port
    type: number
    validate:
      min: 1024
      max: 65535
```

### Secure Dependencies

Pin dependency versions:

```yaml
dependencies:
  packages:
    - github.com/toutaio/toutago@v1.0.0  # Pinned
    - github.com/toutaio/toutago-cosan-router@v1.0.5
```

### SQL Injection Prevention

Use parameterized queries in migrations:

✅ **Good:**
```yaml
migrations:
  - from_version: "1.0.0"
    to_version: "1.1.0"
    up:
      sql:
        - "CREATE TABLE users (id SERIAL PRIMARY KEY, email VARCHAR(255) UNIQUE)"
```

❌ **Bad:**
```yaml
# Don't use string interpolation in SQL
up:
  sql:
    - "CREATE TABLE [[ .table_name ]] ..."  # Dangerous!
```

### File Permission Security

Set appropriate permissions:

```yaml
hooks:
  post_install:
    - type: chmod
      path: config/secrets.yaml
      mode: 0600  # Owner read/write only
```

## Performance Optimization

### Minimize Dependencies

Only include what you need:

```yaml
dependencies:
  packages:
    - github.com/toutaio/toutago  # Essential
    # Don't add unused packages
```

### Optimize Templates

**Pre-compile static content:**

```yaml
files:
  static:  # Use static for non-templated files
    - src: logo.png
      dest: static/logo.png
      
  templates:  # Only template files that need variables
    - src: config.yaml.tmpl
      dest: config.yaml
```

### Lazy Loading

Use conditionals to avoid generating unused code:

```yaml
files:
  templates:
    - src: admin/dashboard.go.tmpl
      dest: admin/dashboard.go
      condition: "[[ .with_admin ]]"
```

### Efficient Hooks

Combine related operations:

✅ **Good:**
```yaml
hooks:
  post_install:
    - "go mod tidy && go build"
```

❌ **Bad:**
```yaml
hooks:
  post_install:
    - "go mod download"
    - "go mod verify"  
    - "go mod tidy"
    - "go build"
```

## Versioning Strategies

### Semantic Versioning

Follow [semver](https://semver.org/):

- **MAJOR**: Breaking changes (1.0.0 → 2.0.0)
- **MINOR**: New features, backward compatible (1.0.0 → 1.1.0)
- **PATCH**: Bug fixes (1.0.0 → 1.0.1)

### Version Constraints

Use appropriate constraints:

```yaml
compatibility:
  min_touta_version: "1.0.0"
  max_touta_version: "2.0.0"  # Be explicit
  go_version: ">=1.21"
```

### Migration Paths

Provide clear upgrade paths:

```yaml
migrations:
  # Support multiple upgrade paths
  - from_version: "1.0.0"
    to_version: "1.1.0"
    
  - from_version: "1.1.0"
    to_version: "1.2.0"
    
  # Don't skip versions
  # ❌ 1.0.0 → 1.2.0 (bad - skips 1.1.0)
```

### Breaking Changes

Document breaking changes clearly:

```yaml
ritual:
  version: "2.0.0"
  
migrations:
  - from_version: "1.0.0"
    to_version: "2.0.0"
    description: "BREAKING: Restructures database schema"
    up:
      sql: [...]
    down:
      sql: [...]  # Always provide rollback
```

## Testing Strategies

### Test Generation

Ensure generated projects work:

```yaml
files:
  templates:
    - src: main_test.go.tmpl
      dest: main_test.go
```

### Validation Tests

Validate before generation:

```yaml
hooks:
  pre_install:
    - type: validate
      schema: ritual.schema.json
```

### Smoke Tests

Test basic functionality:

```yaml
hooks:
  post_install:
    - "go test ./..."
    - "go build"
```

### Test Different Configurations

Test various option combinations:

```bash
# Test with postgres
ritual init blog --config answers-postgres.yaml -o /tmp/test-pg

# Test with mysql
ritual init blog --config answers-mysql.yaml -o /tmp/test-mysql

# Test with no database
ritual init blog --config answers-none.yaml -o /tmp/test-none
```

## File Organization

### Logical Structure

Organize files logically:

```
my-ritual/
├── ritual.yaml           # Manifest
├── README.md             # Documentation
├── templates/            # Templates
│   ├── main.go.tmpl
│   ├── config/
│   │   └── config.yaml.tmpl
│   └── handlers/
│       └── post.go.tmpl
├── static/               # Static files
│   └── .gitignore
└── migrations/           # Optional: external migrations
    └── 1_0_0_to_1_1_0.sql
```

### Template Naming

Use clear, descriptive names:

✅ **Good:**
```
templates/
├── main.go.tmpl
├── handlers_post.go.tmpl
├── models_user.go.tmpl
```

❌ **Bad:**
```
templates/
├── 1.tmpl
├── temp.tmpl
├── new_file.tmpl
```

### Separation of Concerns

Keep different file types separate:

```yaml
files:
  templates:  # Generated files
    - src: templates/main.go.tmpl
      dest: main.go
      
  static:     # Static files
    - src: static/.gitignore
      dest: .gitignore
      
  protected:  # Files not to overwrite
    - config.yaml
    - .env
```

## Questionnaire Design

### Progressive Disclosure

Ask simple questions first:

```yaml
questions:
  # Basic questions
  - name: project_name
    prompt: "Project name?"
    required: true
    
  # Advanced questions (conditional)
  - name: advanced_mode
    prompt: "Enable advanced configuration?"
    type: boolean
    default: false
    
  - name: cache_backend
    prompt: "Cache backend?"
    type: choice
    choices: [redis, memcached, in-memory]
    condition:
      field: advanced_mode
      equals: true
```

### Sensible Defaults

Provide good defaults:

```yaml
questions:
  - name: port
    prompt: "HTTP port?"
    type: number
    default: 8080  # Common default
    
  - name: database
    prompt: "Database?"
    type: choice
    choices: [postgres, mysql, sqlite, none]
    default: "postgres"  # Most capable default
```

### Clear Prompts

Write clear, concise prompts:

✅ **Good:**
```yaml
- name: database_host
  prompt: "Database hostname?"
  default: "localhost"
```

❌ **Bad:**
```yaml
- name: db_h
  prompt: "Where is the DB located (host)?"
```

### Validation Feedback

Provide helpful validation:

```yaml
- name: project_name
  prompt: "Project name?"
  validate:
    pattern: "^[a-z][a-z0-9-]*$"
    error: "Must start with lowercase letter, contain only lowercase letters, numbers, and hyphens"
```

## Migration Patterns

### Always Reversible

Provide down migrations:

```yaml
migrations:
  - from_version: "1.0.0"
    to_version: "1.1.0"
    up:
      sql:
        - "CREATE TABLE posts (id SERIAL PRIMARY KEY)"
    down:
      sql:
        - "DROP TABLE posts"
```

### Idempotent Operations

Make migrations safe to re-run:

```yaml
up:
  sql:
    - "CREATE TABLE IF NOT EXISTS users (...)"
    - "ALTER TABLE posts ADD COLUMN IF NOT EXISTS author_id INT"
```

### Data Preservation

Never lose data:

```yaml
# Instead of DROP
down:
  sql:
    - "ALTER TABLE users RENAME TO users_backup_20260107"
```

### Test Migrations

Test both up and down:

```bash
# Test up
ritual migrate up

# Verify database state
psql -c "SELECT * FROM information_schema.tables"

# Test down
ritual migrate down

# Verify rollback
psql -c "SELECT * FROM information_schema.tables"
```

## Common Anti-Patterns to Avoid

### ❌ Hard-coded Paths

```yaml
# Bad
files:
  templates:
    - src: /home/user/rituals/my-ritual/template.go.tmpl
```

### ❌ Absolute URLs

```yaml
# Bad
dependencies:
  packages:
    - "https://github.com/user/private-repo/archive/main.zip"
```

### ❌ Unvalidated Input

```yaml
# Bad - no validation
questions:
  - name: database_name
    type: text
    # Missing validation - could contain ; or other dangerous chars
```

### ❌ Missing Documentation

```yaml
# Bad - no description
ritual:
  name: my-ritual
  version: "1.0.0"
  # description: ???
```

### ❌ Too Many Questions

```yaml
# Bad - too many questions overwhelms users
questions: # 50+ questions
```

## Summary Checklist

Before publishing a ritual:

- [ ] Semantic versioning followed
- [ ] All secrets protected
- [ ] Sensible defaults provided
- [ ] Input validation added
- [ ] Migrations reversible
- [ ] Tests included
- [ ] Documentation complete
- [ ] Validation passing
- [ ] Generated project compiles
- [ ] No hard-coded paths
- [ ] Security reviewed
- [ ] Performance optimized

## See Also

- [Creating Rituals](CREATING_RITUALS.md)
- [Ritual Format](ritual-format.md)
- [CLI Reference](cli-reference.md)
- [Security Guide](security.md) (if exists)

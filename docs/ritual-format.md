# Ritual Format Specification

This document describes the `ritual.yaml` format for defining application templates in Ritual Grove.

## Overview

A ritual is a YAML-based definition that describes how to create, configure, and deploy a complete application. It includes:

- Metadata (name, version, description)
- Compatibility requirements
- Interactive questions for configuration
- File templates and static files
- Package dependencies
- Migration scripts
- Lifecycle hooks

## Basic Structure

```yaml
ritual:
  name: my-app
  version: 1.0.0
  description: My application
  author: Your Name
  template_engine: fith  # or go-template

compatibility:
  min_touta_version: 0.1.0
  min_go_version: 1.22.0

dependencies:
  packages:
    - github.com/toutaio/toutago
  database:
    required: true
    types: [postgres, mysql]

questions:
  - name: app_name
    prompt: Application name?
    type: text
    required: true

files:
  templates:
    - src: templates/main.go.fith
      dest: main.go
  static:
    - src: static/README.md
      dest: README.md
  protected:
    - config/config.yaml

hooks:
  post_install:
    - go mod tidy
```

## Sections

### ritual (required)

Metadata about the ritual.

```yaml
ritual:
  name: blog-app              # Required: ritual identifier
  version: 1.0.0              # Required: semantic version
  description: Blog app       # Recommended
  author: Toutā Team          # Optional
  license: MIT                # Optional
  homepage: https://...       # Optional
  repository: https://...     # Optional
  tags: [blog, cms]           # Optional
  template_engine: fith       # Optional: fith (default), go-template
```

### compatibility (optional)

Version constraints for compatibility.

```yaml
compatibility:
  min_touta_version: 0.1.0
  max_touta_version: 1.0.0
  min_go_version: 1.22.0
  max_go_version: 1.23.0
```

### dependencies (optional)

Required packages and databases.

```yaml
dependencies:
  packages:
    - github.com/toutaio/toutago
    - github.com/lib/pq
  
  rituals:
    - base-app@1.0.0
  
  database:
    required: true
    types: [postgres, mysql, sqlite]
    min_version: "14.0"
```

### questions (optional)

Interactive configuration prompts.

```yaml
questions:
  - name: app_name
    prompt: What is your application name?
    type: text
    required: true
    validate:
      pattern: "^[a-z][a-z0-9-]*$"
      min_len: 3
      max_len: 50
  
  - name: database_type
    prompt: Database type
    type: choice
    choices: [postgres, mysql, sqlite]
    default: postgres
  
  - name: enable_cache
    prompt: Enable caching?
    type: boolean
    default: true
  
  - name: max_connections
    prompt: Max database connections
    type: number
    default: 100
    validate:
      min: 10
      max: 1000
  
  - name: db_host
    prompt: Database host
    type: text
    default: localhost
    condition:
      field: database_type
      not_equals: sqlite
```

#### Question Types

- `text` - Text input
- `password` - Password (hidden input)
- `choice` - Single choice from list
- `multi_choice` - Multiple choices
- `boolean` - Yes/no
- `number` - Numeric input
- `path` - File system path
- `url` - URL
- `email` - Email address

#### Validation

```yaml
validate:
  pattern: "^[a-z]+$"    # Regex pattern
  min: 1                  # Minimum value (number)
  max: 100                # Maximum value (number)
  min_len: 3              # Minimum length (text)
  max_len: 50             # Maximum length (text)
  custom: my_validator    # Custom validator function
```

#### Conditional Questions

```yaml
condition:
  field: some_field
  equals: some_value       # Show if equals
  not_equals: other_value  # Show if not equals
```

#### Question Helpers

```yaml
helper:
  type: db_test          # Test database connection
  config:
    host_field: db_host
    port_field: db_port
```

### files (optional)

File templates and static files.

```yaml
files:
  templates:
    - src: templates/main.go.fith
      dest: main.go
    
    - src: templates/handlers/
      dest: handlers/
      optional: true
      condition: "{{ enable_api }}"
  
  static:
    - src: static/README.md
      dest: README.md
    
    - src: static/LICENSE
      dest: LICENSE
  
  protected:
    - config/config.yaml  # Never overwrite
    - README.md           # Never overwrite
```

### migrations (optional)

Version migration scripts.

```yaml
migrations:
  - from_version: 1.0.0
    to_version: 1.1.0
    description: Add user table
    idempotent: true
    up:
      sql:
        - CREATE TABLE users (id SERIAL PRIMARY KEY)
      script: scripts/migrate-1.1.0.sh
    down:
      sql:
        - DROP TABLE users
```

### hooks (optional)

Lifecycle hooks.

```yaml
hooks:
  pre_install:
    - echo "Installing..."
  
  post_install:
    - go mod tidy
    - go build
  
  pre_update:
    - go test ./...
  
  post_update:
    - go mod tidy
  
  pre_deploy:
    - go test ./...
  
  post_deploy:
    - echo "Deployed!"
```

### multi_tenancy (optional)

Multi-tenant configuration.

```yaml
multi_tenancy:
  enabled: true
  database_mode: separate  # shared or separate
  tenant_id_field: tenant_id
```

### telemetry (optional)

Monitoring configuration.

```yaml
telemetry:
  enabled: true
  metrics:
    - http_requests
    - db_queries
  config:
    endpoint: http://metrics.example.com
```

### parent (optional)

Ritual inheritance.

```yaml
parent:
  name: base-app
  version: 1.0.0
  source: https://github.com/org/base-app.git
```

## Template Variables

In Fíth templates (default):

```jinja2
{{ app_name }}              {# Answer from questions #}
{{ app_name|pascal }}       {# PascalCase #}
{{ app_name|snake }}        {# snake_case #}
{{ app_name|kebab }}        {# kebab-case #}
{{ now() }}                 {# Current timestamp #}
{{ ritual_version }}        {# Ritual version #}
```

In Go templates:

```go
{{ .AppName }}
{{ pascal .AppName }}
{{ snake .AppName }}
{{ now }}
```

## Example: Complete Blog Ritual

```yaml
ritual:
  name: blog
  version: 1.0.0
  description: Complete blog application with posts and comments
  author: Toutā Team
  license: MIT
  template_engine: fith

compatibility:
  min_touta_version: 0.1.0
  min_go_version: 1.22.0

dependencies:
  packages:
    - github.com/toutaio/toutago
    - github.com/lib/pq
  database:
    required: true
    types: [postgres]
    min_version: "14.0"

questions:
  - name: app_name
    prompt: Blog name
    type: text
    required: true
  
  - name: admin_email
    prompt: Admin email
    type: email
    required: true
  
  - name: port
    prompt: HTTP port
    type: number
    default: 8080

files:
  templates:
    - src: templates/main.go.fith
      dest: main.go
    - src: templates/handlers/
      dest: handlers/
  
  static:
    - src: static/README.md
      dest: README.md
  
  protected:
    - config/config.yaml

hooks:
  post_install:
    - go mod tidy
    - go build -o blog

migrations:
  - from_version: 1.0.0
    to_version: 1.1.0
    description: Add comments table
    up:
      sql:
        - CREATE TABLE comments (id SERIAL PRIMARY KEY)
    down:
      sql:
        - DROP TABLE comments
```

## See Also

- [Fíth Template Engine](https://github.com/toutaio/toutago-fith-renderer)
- [Examples](../examples/)

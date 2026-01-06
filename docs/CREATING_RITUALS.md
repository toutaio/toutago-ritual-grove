# Creating Rituals Guide

This guide explains how to create custom rituals for Toutā projects.

## What is a Ritual?

A ritual is a reusable project template that scaffolds a complete Toutā application with:
- Pre-configured project structure
- Database models and migrations
- HTTP handlers and routes
- Frontend templates (Traditional, Inertia.js, or HTMX)
- Tests and documentation
- Build configuration

## Ritual Structure

```
my-ritual/
├── ritual.yaml           # Ritual configuration and metadata
├── templates/            # Go template files
│   ├── main.go.tmpl
│   ├── go.mod.tmpl
│   ├── models/
│   ├── handlers/
│   └── views/
├── frontend/             # Frontend templates (optional)
│   ├── pages/
│   ├── components/
│   └── layouts/
├── README.md            # Ritual documentation
└── .ritual-ignore       # Files to exclude

```

## Creating Your First Ritual

### Step 1: Create ritual.yaml

```yaml
ritual:
  name: my-app
  version: 1.0.0
  description: A sample application
  author: Your Name
  license: MIT
  tags:
    - web
    - api

compatibility:
  min_touta_version: "0.1.0"
  go_version: ">=1.21"

dependencies:
  packages:
    - github.com/toutaio/toutago-cosan-router
    - github.com/toutaio/toutago-datamapper

questions:
  - name: app_name
    type: text
    prompt: "What is your application name?"
    default: "My App"
    required: true
    validate:
      min_len: 1
      max_len: 100

  - name: module_path
    type: text
    prompt: "What is your Go module path?"
    default: "example.com/myapp"
    required: true
    validate:
      pattern: "^[a-z0-9._/-]+$"

  - name: port
    type: number
    prompt: "Which port should the app run on?"
    default: 8080
    validate:
      min: 1024
      max: 65535

  - name: database_type
    type: choice
    prompt: "Select database:"
    choices:
      - postgres
      - mysql
      - sqlite
    default: postgres

files:
  templates:
    - src: templates/main.go.tmpl
      dest: main.go
    
    - src: templates/go.mod.tmpl
      dest: go.mod
    
    - src: templates/README.md.tmpl
      dest: README.md

hooks:
  post_install:
    - "go mod tidy"
    - "go build"
```

### Step 2: Create Templates

Create `templates/main.go.tmpl`:

```go
package main

import (
    "log"
    "{{ .module_path }}/internal/handlers"
    "github.com/toutaio/toutago-cosan-router"
)

func main() {
    router := cosan.New()
    
    // Setup routes
    h := handlers.New()
    router.GET("/", h.Index)
    
    log.Printf("Starting {{ .app_name }} on port {{ .port }}")
    if err := router.Listen(":{{ .port }}"); err != nil {
        log.Fatal(err)
    }
}
```

Create `templates/go.mod.tmpl`:

```
module {{ .module_path }}

go {{ .go_version | default "1.21" }}

require (
{{- range .dependencies.packages }}
    {{ . }} latest
{{- end }}
)
```

### Step 3: Test Your Ritual

```bash
# Install ritual locally
touta ritual install ./my-ritual

# Use the ritual
touta create my-ritual my-project
cd my-project
go run main.go
```

## Question Types

### Text Input

```yaml
- name: app_name
  type: text
  prompt: "Application name?"
  default: "My App"
  required: true
  validate:
    min_len: 1
    max_len: 100
    pattern: "^[a-zA-Z0-9 ]+$"
```

### Number Input

```yaml
- name: port
  type: number
  prompt: "Port number?"
  default: 8080
  validate:
    min: 1024
    max: 65535
```

### Boolean Input

```yaml
- name: enable_auth
  type: boolean
  prompt: "Enable authentication?"
  default: true
```

### Choice Input

```yaml
- name: database
  type: choice
  prompt: "Select database:"
  choices:
    - postgres
    - mysql
    - sqlite
  default: postgres
  required: true
```

### Multi-Select

```yaml
- name: features
  type: multi
  prompt: "Select features:"
  choices:
    - authentication
    - authorization
    - api
    - admin
  defaults:
    - api
```

### Password Input

```yaml
- name: db_password
  type: password
  prompt: "Database password:"
  required: true
```

## Conditional Questions

Show questions based on previous answers:

```yaml
- name: database_type
  type: choice
  prompt: "Select database:"
  choices:
    - postgres
    - mysql
  default: postgres

- name: pg_schema
  type: text
  prompt: "PostgreSQL schema:"
  default: "public"
  condition:
    field: database_type
    equals: postgres

- name: mysql_charset
  type: text
  prompt: "MySQL charset:"
  default: "utf8mb4"
  condition:
    field: database_type
    equals: mysql
```

### Complex Conditions

```yaml
- name: enable_ssr
  type: boolean
  prompt: "Enable SSR?"
  default: false
  condition:
    expression: "{{ eq .frontend_type \"inertia-vue\" }}"
```

## Template Syntax

Toutā uses Go's `text/template` with additional functions.

### Basic Variables

```go
{{ .app_name }}
{{ .module_path }}
{{ .port }}
```

### Conditionals

```go
{{- if .enable_auth }}
import "github.com/toutaio/toutago-breitheamh-auth"
{{- end }}

{{- if eq .database_type "postgres" }}
import "github.com/lib/pq"
{{- else if eq .database_type "mysql" }}
import "github.com/go-sql-driver/mysql"
{{- end }}
```

### Loops

```go
{{- range .features }}
- {{ . }}
{{- end }}
```

### Functions

```go
{{ .app_name | lower }}
{{ .app_name | upper }}
{{ .app_name | title }}
{{ .module_path | base }}
{{ default "default value" .optional_field }}
```

## Frontend Integration

### Traditional (Fíth Templates)

```yaml
files:
  templates:
    - src: templates/views/index.fith.tmpl
      dest: views/index.fith
```

### Inertia.js

```yaml
questions:
  - name: frontend_type
    type: choice
    prompt: "Select frontend:"
    choices:
      - traditional
      - inertia-vue
      - htmx
    default: traditional

files:
  templates:
    # Backend
    - src: templates/handlers/post_handler.go.tmpl
      dest: internal/handlers/post_handler.go
    
    # Frontend (conditional)
    - src: templates/frontend/pages/Posts/Index.vue.tmpl
      dest: frontend/pages/Posts/Index.vue
      condition:
        expression: "{{ eq .frontend_type \"inertia-vue\" }}"
    
    - src: templates/frontend/app.js.tmpl
      dest: frontend/app.js
      condition:
        expression: "{{ eq .frontend_type \"inertia-vue\" }}"
```

### HTMX

```yaml
files:
  templates:
    - src: templates/views/posts/index.fith.tmpl
      dest: views/posts/index.fith
      condition:
        expression: "{{ eq .frontend_type \"htmx\" }}"
    
    - src: templates/views/posts/_list.fith.tmpl
      dest: views/posts/_list.fith
      condition:
        expression: "{{ eq .frontend_type \"htmx\" }}"
```

## Hooks

Hooks run commands at specific lifecycle points.

### Available Hooks

```yaml
hooks:
  pre_install:
    - "echo 'Starting installation...'"
  
  post_install:
    - "go mod tidy"
    - "go build"
  
  pre_generate:
    - "mkdir -p internal/models"
  
  post_generate:
    - "gofmt -w ."
    - "go test ./..."
```

### Hook Tasks

Use built-in tasks instead of shell commands for cross-platform compatibility:

```yaml
hooks:
  post_install:
    - task: go-mod-tidy
    
    - task: npm-install
      config:
        directory: frontend
    
    - task: npm-build
      config:
        directory: frontend
        script: build
    
    - task: directory-create
      config:
        path: internal/models
        mode: 0755
    
    - task: file-copy
      config:
        src: config/example.env
        dest: .env
        mode: 0600
```

### Inertia-Specific Tasks

```yaml
hooks:
  post_install:
    - task: setup-inertia-middleware
      config:
        root_view: app
        ssr: false
    
    - task: generate-typescript-types
      config:
        output: frontend/types/models.ts
        models:
          - internal/models
```

## Database Migrations

Include migrations in your ritual:

```yaml
files:
  templates:
    - src: templates/migrations/001_create_users.sql.tmpl
      dest: migrations/001_create_users.sql
```

Create `templates/migrations/001_create_users.sql.tmpl`:

```sql
-- +migrate Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE users;
```

## Testing Your Ritual

### Unit Tests

Create tests for generated code:

```yaml
files:
  templates:
    - src: templates/models/user_test.go.tmpl
      dest: internal/models/user_test.go
```

### Integration Tests

```yaml
hooks:
  post_install:
    - "go test ./..."
```

## Publishing Your Ritual

### Option 1: Git Repository

```bash
# Create git repository
git init
git add .
git commit -m "Initial ritual"
git remote add origin git@github.com:user/my-ritual.git
git push -u origin main
```

Users can install with:
```bash
touta ritual install git@github.com:user/my-ritual.git
```

### Option 2: Tarball

```bash
tar -czf my-ritual.tar.gz my-ritual/
```

Users can install with:
```bash
touta ritual install ./my-ritual.tar.gz
```

### Option 3: Ritual Registry (Future)

```bash
touta ritual publish
```

## Best Practices

### 1. Provide Sensible Defaults

```yaml
- name: port
  default: 8080  # Good default for development
```

### 2. Use Validation

```yaml
- name: email
  validate:
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
```

### 3. Document Your Ritual

Include comprehensive README:

```markdown
# My Ritual

## What It Creates

- REST API with authentication
- PostgreSQL database
- Vue.js frontend with Inertia

## Requirements

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+

## Usage

\`\`\`bash
touta create my-ritual my-project
\`\`\`

## Configuration

...
```

### 4. Include Examples

```yaml
files:
  templates:
    - src: templates/examples/handler_example.go.tmpl
      dest: examples/handler_example.go
```

### 5. Use Semantic Versioning

Update version for:
- Major: Breaking changes
- Minor: New features
- Patch: Bug fixes

### 6. Test Across Platforms

Test your ritual on:
- Linux
- macOS  
- Windows

### 7. Keep Dependencies Minimal

Only include necessary packages.

### 8. Provide Migration Paths

Help users upgrade between versions.

## Advanced Features

### Custom Validators

```yaml
questions:
  - name: slug
    type: text
    prompt: "URL slug:"
    validate:
      pattern: "^[a-z0-9-]+$"
      custom: "validateSlug"  # Custom Go function
```

### Dynamic File Generation

Generate files based on user input:

```yaml
files:
  dynamic:
    - pattern: "internal/models/{{.entity}}.go"
      template: templates/model.go.tmpl
      foreach: entities
```

### Multi-Language Support

```yaml
ritual:
  name: my-ritual
  locales:
    - en
    - es
    - fr

questions:
  - name: app_name
    prompt:
      en: "What is your application name?"
      es: "¿Cuál es el nombre de su aplicación?"
      fr: "Quel est le nom de votre application?"
```

## Troubleshooting

### Common Issues

**Problem:** Template syntax errors
**Solution:** Test templates with sample data

**Problem:** Hooks not executing
**Solution:** Check hook syntax and ensure commands are cross-platform

**Problem:** Conditional files not generating
**Solution:** Verify condition expressions match user input

## Examples

Check these rituals for reference:
- `blog` - Full-featured blog with multiple frontend options
- `api` - REST API template
- `admin` - Admin panel with CRUD operations

## Resources

- Template Functions: https://pkg.go.dev/text/template
- YAML Specification: https://yaml.org/spec/
- Semantic Versioning: https://semver.org/

## Support

- GitHub: https://github.com/toutaio/toutago-ritual-grove
- Discord: https://discord.gg/touta

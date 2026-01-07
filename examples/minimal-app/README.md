# Minimal App Ritual Example

The most basic example ritual showing the minimum required structure for a working ritual.

## Purpose

This is the simplest possible ritual that demonstrates:
- Minimal `ritual.yaml` configuration
- Go template engine usage
- Basic project scaffolding
- Essential file generation
- Single post-generation hook

## Structure

```
minimal-app/
├── ritual.yaml          # Minimal ritual configuration
├── README.md           # This file
└── templates/          # Template files
    ├── main.go.tmpl
    ├── go.mod.tmpl
    ├── README.md.tmpl
    ├── .env.example.tmpl
    └── handlers/
        └── hello.go.tmpl
```

## What Gets Generated

When you use this ritual, it creates:

```
my-app/
├── main.go              # HTTP server entry point
├── go.mod               # Go module definition
├── README.md            # Project documentation
├── .env.example         # Environment variables template
├── .gitignore           # Git ignore patterns
└── handlers/            # HTTP handlers (if enabled)
    └── hello.go
```

## Questionnaire

Only 4 questions - the absolute minimum:

1. **app_name**: Your application name (validated, lowercase-hyphenated)
2. **module_name**: Go module path (defaults to github.com/example/{{app_name}})
3. **port**: HTTP server port (default: 8080, range: 1024-65535)
4. **with_example_handler**: Include example handler? (boolean, default: true)

## Features

### 1. Template Engine

Uses **Go's text/template** engine (instead of Fíth):

```yaml
ritual:
  template_engine: go-template
```

Variables are accessed with: `{{ .variable_name }}`

### 2. Computed Default

The `module_name` question uses the `app_name` answer as a default:

```yaml
- name: module_name
  default: "github.com/example/{{ .app_name }}"
```

### 3. Conditional File Generation

The example handler is only generated if requested:

```yaml
files:
  templates:
    - src: "handlers/hello.go.tmpl"
      dest: "handlers/hello.go"
      condition: "{{ .with_example_handler }}"
```

### 4. Static Files

The `.gitignore` is copied without template processing:

```yaml
files:
  static:
    - src: ".gitignore"
      dest: ".gitignore"
```

### 5. Single Hook

Only one post-generation hook to download dependencies:

```yaml
hooks:
  post_generate:
    - "go mod tidy"
```

## Usage

### Generate a Project

```bash
touta ritual init minimal-app --output my-new-app
```

### Answer Questions

```
Application name? my-cool-api
Go module name? [github.com/example/my-cool-api]
HTTP port? [8080]
Include example handler? [Y/n]
```

### Run Generated Project

```bash
cd my-new-app
go run main.go
```

Visit `http://localhost:8080` to see your app!

## Customization Guide

### 1. Add More Questions

```yaml
questions:
  - name: author_name
    prompt: "Your name?"
    type: text
    required: true
```

### 2. Add More Templates

Create a new template file:

```yaml
files:
  templates:
    - src: "config.yaml.tmpl"
      dest: "config/config.yaml"
```

### 3. Add Dependencies

```yaml
dependencies:
  packages:
    - github.com/toutaio/toutago-fith-renderer
```

### 4. Add More Hooks

```yaml
hooks:
  post_generate:
    - "go mod tidy"
    - "go fmt ./..."
    - "go build -o {{.app_name}}"
```

## Comparison with simple-app

| Feature | minimal-app | simple-app |
|---------|-------------|------------|
| Template Engine | Go templates | Fíth |
| Questions | 4 | 6 |
| Conditional Questions | No | Yes |
| Database Support | No | Yes |
| Complexity | Minimal | Simple |
| Best For | Learning | Starting projects |

## When to Use This

Use `minimal-app` when:
- ✅ Learning ritual creation
- ✅ Creating a quick prototype
- ✅ Building a simple HTTP service
- ✅ Testing ritual features

Use `simple-app` or `blog` when:
- ❌ Need database support
- ❌ Building a full application
- ❌ Want more features out of the box

## Template Syntax Reference

### Variables

```go
{{ .app_name }}           // String variable
{{ .port }}               // Number variable
{{ .with_example_handler }} // Boolean variable
```

### Conditionals

```go
{{ if .with_example_handler }}
  // This code is included
{{ end }}

{{ if eq .database "postgres" }}
  // PostgreSQL-specific code
{{ else }}
  // Other database code
{{ end }}
```

### Iteration

```go
{{ range .items }}
  Item: {{ . }}
{{ end }}
```

## Next Steps

After using this example:

1. **Try simple-app** for more features
2. **Read the docs** in `../../docs/`
3. **Create your own ritual** based on this template
4. **Explore built-in rituals** in `../../rituals/`

## Files Included

- `ritual.yaml` - Complete ritual configuration
- `templates/main.go.tmpl` - HTTP server with Cosan router
- `templates/go.mod.tmpl` - Go module with Toutā dependencies
- `templates/README.md.tmpl` - Generated project documentation
- `templates/.env.example.tmpl` - Environment variables
- `templates/handlers/hello.go.tmpl` - Example HTTP handler

## Learn More

- [Ritual Format](../../docs/ritual-format.md)
- [Template Guide](../../docs/templates.md)
- [Go Templates](https://pkg.go.dev/text/template)
- [Toutā Framework](https://github.com/toutaio/toutago)

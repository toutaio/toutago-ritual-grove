# Ritual Grove

**Application recipe system for Toutā framework** - Create, manage, and deploy complete applications from templates.

## Overview

Ritual Grove is a powerful system for building production-ready applications using **rituals** (recipes/templates). Think of it as a sophisticated project generator that goes beyond simple scaffolding - it handles the complete lifecycle from creation to deployment and updates.

## Features

- 🎯 **Create Complete Applications** - Generate production-ready apps from rituals (blog, CRM, wiki, API server, etc.)
- 📦 **Package Management** - Automatic dependency resolution and package installation
- 🔄 **Lifecycle Management** - Deploy, update, and rollback applications
- 🧩 **Mixin System** - Add features to existing projects (auth, comments, admin, etc.)
- 🏢 **Multi-Tenancy** - Built-in support for multi-tenant applications
- 📝 **Interactive Setup** - Smart questionnaires with validation and helpers
- 🔌 **Pluggable Templates** - Fíth (default), Go templates, or custom engines
- 📚 **Multi-Source Loading** - Built-in rituals, Git repos, or local tarballs
- 🔐 **File Protection** - Preserves user modifications during updates
- 🔒 **Lock Files** - Reproducible builds with `ritual.lock`

## Installation

```bash
go install github.com/toutaio/toutago-ritual-grove/cmd/ritual@latest
```

## Quick Start

```bash
# List available rituals
ritual list

# Create a new blog application
ritual create blog-app --ritual blog

# Add authentication to existing project
ritual mixin add auth

# Deploy to production
ritual deploy

# Update to newer ritual version
ritual update
```

## Ritual Format

A ritual is a YAML-based definition with templates and logic:

```yaml
ritual:
  name: blog-app
  version: 1.0.0
  description: A production-ready blog application
  template_engine: fith  # Options: fith, go-template

questions:
  - name: app_name
    prompt: Application name
    type: text
    required: true
  
  - name: database
    prompt: Database type
    type: choice
    choices: [postgres, mysql, sqlite]

templates:
  - src: templates/main.go.fith
    dest: main.go
  
  - src: templates/handlers/
    dest: handlers/

packages:
  - github.com/toutaio/toutago
  - github.com/lib/pq

mixins:
  - name: auth
    description: User authentication
  - name: comments
    description: Comment system
```

## Architecture

```
toutago-ritual-grove/
├── cmd/ritual/          # CLI entry point
├── internal/
│   ├── ritual/          # Core ritual engine
│   ├── registry/        # Ritual discovery
│   ├── questionnaire/   # Interactive prompts
│   ├── generator/       # Code generation
│   ├── deployment/      # Update/deploy logic
│   ├── storage/         # State management
│   └── validator/       # Ritual validation
├── pkg/ritual/          # Public API
├── rituals/             # Built-in rituals
│   ├── blog/
│   ├── wiki/
│   ├── api-server/
│   └── microservice/
├── examples/            # Example rituals
└── docs/                # Documentation
```

## Development Status

This project is under active development.

## Documentation

See [docs/](docs/) for detailed documentation on ritual format, examples, and usage.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [toutago](https://github.com/toutaio/toutago) - Main Toutā framework
- [toutago-fith-renderer](https://github.com/toutaio/toutago-fith-renderer) - Template engine (Jinja2-style)
- [toutago-nasc-dependency-injector](https://github.com/toutaio/toutago-nasc-dependency-injector) - Dependency injection
- [toutago-scela-bus](https://github.com/toutaio/toutago-scela-bus) - Message bus

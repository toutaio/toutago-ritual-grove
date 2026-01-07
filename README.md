# Ritual Grove - Application Recipe System for Toutā

[![CI](https://github.com/toutaio/toutago-ritual-grove/actions/workflows/ci.yml/badge.svg)](https://github.com/toutaio/toutago-ritual-grove/actions/workflows/ci.yml)
[![Lint](https://github.com/toutaio/toutago-ritual-grove/actions/workflows/lint.yml/badge.svg)](https://github.com/toutaio/toutago-ritual-grove/actions/workflows/lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/toutaio/toutago-ritual-grove.svg)](https://pkg.go.dev/github.com/toutaio/toutago-ritual-grove)
[![Go Report Card](https://goreportcard.com/badge/github.com/toutaio/toutago-ritual-grove)](https://goreportcard.com/report/github.com/toutaio/toutago-ritual-grove)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> **Ritual Grove** - Create, manage, and deploy complete applications from templates and recipes.

## Status

✅ **Production Ready** - Stable v1.0+ releases  
📦 [View Releases](https://github.com/toutaio/toutago-ritual-grove/releases) for the latest version  
📖 [Changelog](CHANGELOG.md) - Full version history

## Overview

Ritual Grove is a powerful system for building production-ready applications using **rituals** (recipes/templates). Think of it as a sophisticated project generator that goes beyond simple scaffolding - it handles the complete lifecycle from creation to deployment and updates.

**Architecture:** Ritual Grove is a **library integrated into the main `touta` CLI**, not a standalone tool. All commands are accessed through the main `touta` binary.

## Features

- 🎯 **Create Complete Applications** - Generate production-ready apps from rituals (blog, CRM, wiki, API server, etc.)
- 📦 **Package Management** - Automatic dependency resolution and package installation
- 🔄 **Lifecycle Management** - Deploy, update, and rollback applications
- 📊 **Deployment History** - Track all deployments with timestamps, status, and error logs
- 🛡️ **Protected Files** - Glob patterns to prevent overwriting user customizations
- 🧩 **Mixin System** - Add features to existing projects (auth, comments, admin, etc.)
- 🏢 **Multi-Tenancy** - Built-in support for multi-tenant applications
- 📝 **Interactive Setup** - Smart questionnaires with validation and helpers
- 🔌 **Pluggable Templates** - Fíth (default), Go templates, or custom engines
- 📚 **Multi-Source Loading** - Built-in rituals, Git repos, or local tarballs
- 🔐 **File Protection** - Preserves user modifications during updates (`.ritual/protected.txt`)
- 🔒 **Lock Files** - Reproducible builds with `ritual.lock`
- 📋 **Declarative Tasks** - 30+ built-in tasks for hooks (file ops, Go ops, HTTP, validation)

## Installation

Ritual Grove is integrated into the main Toutā CLI. Install or build the `touta` binary:

```bash
# From the toutago repository
cd toutago
go build -o touta cmd/touta/main.go
sudo mv touta /usr/local/bin/
```

## Quick Start

All ritual commands are accessed through the `touta` binary:

```bash
# List available rituals
touta ritual list

# Create a new blog application
touta ritual init blog-app --ritual blog

# Add authentication to existing project
touta ritual mixin add auth

# Deploy to production
touta ritual deploy

# Update to newer ritual version
touta ritual update

# Clean ritual cache (useful after upgrading touta)
touta ritual clean --force
```

## Troubleshooting

### Rituals Not Updating After Rebuild

If you rebuild the `touta` binary but rituals still show old content, you need to clean the ritual cache:

```bash
# Clear the ritual cache
touta ritual clean --force

# Now rituals will be re-extracted from the new binary
touta ritual list
```

The cache is stored in `~/.toutago/ritual-cache/`. Embedded rituals are extracted once and cached for performance. After rebuilding `touta`, the cache may contain outdated versions.

## Configuration

### Ritual Search Paths

Ritual Grove searches for rituals in the following locations (in order):

1. **Environment variable:** `TOUTA_RITUALS_PATH` - Custom path for development
2. **Built-in:** `<executable-dir>/rituals/` - Bundled with touta binary
3. **Current directory:** `./rituals/` - Project-local rituals
4. **Current directory:** `./.ritual/` - Alternative local path
5. **User home:** `~/.toutago/rituals/` - User-installed rituals

### Development Setup

For development, set the `TOUTA_RITUALS_PATH` environment variable to point to your ritual-grove repository:

```bash
export TOUTA_RITUALS_PATH=/path/to/toutago-ritual-grove/rituals
touta ritual list  # Now finds rituals from development directory
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

Ritual Grove is designed as a **library that integrates into the main `touta` CLI**:

```
toutago-ritual-grove/
├── pkg/cli/             # Exported CLI commands (for touta integration)
│   └── ritual.go        # RitualCommand() for cobra integration
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

**Integration:** The `touta` binary imports `pkg/cli.RitualCommand()` and adds it as a subcommand.

## Development Status

This project is under active development.

## Etymology

**Ritual Grove** combines two meaningful concepts:

- **Ritual**: A ceremonial act or series of acts performed according to a prescribed order. In software development, rituals are established patterns and procedures for creating applications—blueprints that encode best practices and architectural decisions.

- **Grove**: A sacred space in Celtic tradition where druids gathered for ceremonies and knowledge sharing. In our context, it represents a curated collection of application templates and recipes, a sanctuary of proven patterns where developers can find and share project archetypes.

Together, **Ritual Grove** symbolizes a sacred space where application creation rituals (templates) are preserved, shared, and performed—a garden of recipes for growing robust software projects.

## Documentation

### Core Guides
- [Deployment Management](docs/deployment-management.md) - History tracking and protected files
- [Hook Tasks Reference](docs/hook-tasks-reference.md) - All available declarative tasks
- [Ritual Format](docs/ritual-format.md) - Complete ritual.yaml specification

### Built-in Rituals
- [Minimal](rituals/minimal/README.md) - Empty starting point
- [Hello World](rituals/hello-world/README.md) - Simple HTTP server
- [Basic Site](rituals/basic-site/README.md) - Multi-page website with routing
- [Blog](rituals/blog/README.md) - Blog with posts, categories, comments
- [Wiki](rituals/wiki/README.md) - Wiki with pages and search
- [Fullstack Inertia Vue](rituals/fullstack-inertia-vue/README.md) - Modern SPA with Inertia.js

See [docs/](docs/) for more detailed documentation.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [toutago](https://github.com/toutaio/toutago) - Main Toutā framework
- [toutago-fith-renderer](https://github.com/toutaio/toutago-fith-renderer) - Template engine (Jinja2-style)
- [toutago-nasc-dependency-injector](https://github.com/toutaio/toutago-nasc-dependency-injector) - Dependency injection
- [toutago-scela-bus](https://github.com/toutaio/toutago-scela-bus) - Message bus

The CI should now pass with commit 052ed8a which includes all lint fixes.

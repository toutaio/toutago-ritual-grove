# Ritual Grove

**Application recipe system for Toutā framework** - Create, manage, and deploy complete applications from templates.

## Overview

Ritual Grove is a powerful system for building production-ready applications using **rituals** (recipes/templates). Think of it as a sophisticated project generator that goes beyond simple scaffolding - it handles the complete lifecycle from creation to deployment and updates.

**Architecture:** Ritual Grove is a **library integrated into the main `touta` CLI**, not a standalone tool. All commands are accessed through the main `touta` binary.

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
```

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

The CI should now pass with commit 052ed8a which includes all lint fixes.

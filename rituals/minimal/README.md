# Minimal Ritual

A minimal Toutā application ritual that creates a basic project structure with essential files.

## What it Creates

This ritual generates:

- **Main application** (`cmd/{project-name}/main.go`) with basic HTTP server
- **Go module** (`go.mod`) with Toutā dependency
- **README** with project documentation
- **.gitignore** with common patterns

## Features

- Simple HTTP server with health check endpoint
- Configurable port number
- Clean project structure following Go conventions
- Automatic `go mod tidy` after installation

## Usage

```bash
ritual create minimal
```

You'll be prompted for:

- **Project name** - The name of your project (lowercase with hyphens)
- **Module name** - Your Go module path (e.g., github.com/user/project)
- **Port** - The HTTP server port (default: 8080)

## Generated Endpoints

- `GET /` - Welcome message
- `GET /health` - Health check (returns "OK")

## After Installation

The ritual automatically runs:
1. `go mod tidy` - Downloads dependencies
2. `go fmt ./...` - Formats code

## Running Your Project

```bash
cd {project-name}
go run cmd/{project-name}/main.go
```

Then visit `http://localhost:{port}` in your browser.

## Next Steps

After creating your minimal project, you can:

1. Add more endpoints in `main.go`
2. Create additional packages for handlers, models, etc.
3. Apply other rituals to add features (auth, database, etc.)

## License

MIT

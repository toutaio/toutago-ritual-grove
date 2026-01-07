# Hello World Ritual

A minimal Toutā application that demonstrates the basics of building a web service.

## Description

This ritual creates a simple "Hello World" web application using the Toutā framework. It's perfect for:
- Learning Toutā basics
- Quick prototyping
- Testing your Toutā installation
- Understanding ritual structure

## What's Generated

The ritual generates a minimal Toutā application with:
- **Main entry point** (`main.go`) - Server initialization and routing setup
- **Hello handler** (`internal/handlers/hello.go`) - Simple HTTP handler
- **Go module** (`go.mod`) - Dependency management
- **README** - Project documentation

## Usage

Initialize a new hello-world project:

```bash
touta ritual init hello-world
```

Or with custom configuration:

```bash
touta ritual init hello-world --config answers.yaml
```

### Interactive Questions

The ritual will ask you:
1. **Application name** - Name for your application (default: "hello-world")
2. **Port** - HTTP server port (default: 8080, range: 1024-65535)

### Configuration File Example

Create an `answers.yaml` file to skip interactive prompts:

```yaml
app_name: my-hello-app
port: 3000
```

Then run:

```bash
touta ritual init hello-world --config answers.yaml
```

## Running the Generated Application

After initialization:

```bash
cd <your-app-name>
go run main.go
```

Visit `http://localhost:8080` (or your configured port) to see the greeting.

## Structure

```
<your-app-name>/
├── main.go                   # Application entry point
├── internal/
│   └── handlers/
│       └── hello.go          # Hello handler
├── go.mod                    # Go dependencies
└── README.md                 # Project documentation
```

## Dependencies

- `github.com/toutaio/toutago` - Toutā core framework
- `github.com/toutaio/toutago-cosan-router` - HTTP router

## Requirements

- **Toutā**: >= 0.1.0
- **Go**: >= 1.21

## Customization

After generation, you can:
- Add more handlers in `internal/handlers/`
- Extend routing in `main.go`
- Add middleware
- Integrate databases
- Add templates

## Next Steps

After mastering this ritual, try:
- **minimal** - Bare-bones Toutā project
- **basic-site** - Static site with templates
- **blog** - Blog with posts and categories
- **fullstack-inertia-vue** - Full-stack SPA with Vue.js

## License

Generated projects inherit your license choice. The ritual itself is MIT licensed.

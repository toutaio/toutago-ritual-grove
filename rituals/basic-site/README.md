# Basic Site Ritual

A simple starter ritual for creating a basic Toutā website with homepage.

## Features

- ✅ Clean project structure
- ✅ Cosan Router for routing  
- ✅ Fíth Renderer for templating
- ✅ Homepage handler and view
- ✅ Static file serving setup
- ✅ Go module configuration

## Usage

Initialize a new project:

```bash
touta ritual init basic-site
```

Or initialize with all defaults (skip questions):

```bash
touta ritual init basic-site --yes
```

Or initialize to a specific directory:

```bash
touta ritual init basic-site --output ./my-site
```

## Questions

The ritual will ask you:

1. **Site name** - The display name for your site (default: "My Toutā Site")
2. **Port** - Server port number (default: 8080, range: 1024-65535)
3. **Enable database** - Whether to include database support (default: false)
4. **Go version** - Go version for go.mod (default: "1.22")

## Generated Structure

```
your-project/
├── go.mod                 # Go module file
├── main.go               # Application entry point
├── handlers/             # Request handlers
│   └── home.go          # Homepage handler
├── views/               # HTML templates
│   └── home.html       # Homepage template
└── static/             # Static assets (CSS, JS, images)
```

## Next Steps

After initialization:

1. Navigate to your project:
   ```bash
   cd your-project
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the development server:
   ```bash
   touta serve
   ```
   
   Or use `go run`:
   ```bash
   go run main.go
   ```

4. Open your browser to `http://localhost:8080` (or your configured port)

## Customization

### Adding new pages

1. Create a new handler in `handlers/`
2. Create a new template in `views/`
3. Register the route in `main.go`

### Adding static assets

Place CSS, JavaScript, and images in the `static/` directory. They'll be served at `/static/*`

### Enabling database

Set `enable_database: true` when initializing, or manually add database configuration later.

## Requirements

- Toutā v0.2.0+
- Go 1.21+

## License

MIT

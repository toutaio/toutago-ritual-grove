# Full Stack (Inertia + Vue) Ritual

This ritual creates a complete full-stack web application with:

- **Backend**: Go with Toutago framework
- **Frontend**: Vue 3 with Composition API
- **Bridge**: Inertia.js for SPA-like experience
- **Tooling**: Vite for fast development and builds
- **Type Safety**: TypeScript support

## Features

- Server-side routing with client-side navigation
- Hot module replacement (HMR)
- TypeScript type generation from Go structs
- Optional authentication scaffolding
- Optional SSR support
- Database integration (PostgreSQL or MySQL)

## Usage

```bash
toutago ritual run fullstack-inertia-vue
```

You'll be prompted for:
- Project name
- Go module path
- Database driver (postgres/mysql)
- Server port
- Authentication (yes/no)
- SSR support (yes/no)

## Development

After generation:

1. Start the Vite dev server:
   ```bash
   npm run dev
   ```

2. In another terminal, start the Go server:
   ```bash
   go run main.go
   ```

3. Open http://localhost:8080 (or your configured port)

## Project Structure

```
your-project/
├── main.go                 # Go application entry point
├── frontend/
│   ├── app.js             # Inertia app setup
│   ├── pages/             # Vue page components
│   ├── layouts/           # Layout components
│   └── components/        # Reusable components
├── public/
│   └── dist/              # Built assets (generated)
├── vite.config.js         # Vite configuration
├── tsconfig.json          # TypeScript configuration
└── .env                   # Environment variables
```

## Next Steps

1. Add more routes in `main.go`
2. Create new pages in `frontend/pages/`
3. Build reusable components in `frontend/components/`
4. Configure your database connection
5. Implement authentication if enabled
6. Deploy your application

## Documentation

- [Toutago Documentation](https://github.com/toutaio/toutago)
- [Inertia.js Guide](https://inertiajs.com)
- [Vue 3 Documentation](https://vuejs.org)
- [Vite Guide](https://vitejs.dev)

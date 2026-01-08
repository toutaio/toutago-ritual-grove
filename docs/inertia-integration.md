# Inertia.js Integration Guide

This guide explains how to use Inertia.js with Toutago Ritual Grove for building modern, reactive frontends with server-side routing.

## Overview

Toutago's Inertia.js integration allows you to build single-page applications using classic server-side routing and controllers. You get the benefits of a modern SPA without the complexity of building an API.

## Choosing Inertia.js

When creating a new ritual project, you'll be asked to choose a frontend framework:

```
Select frontend framework:
  > traditional  (Server-rendered HTML templates)
    inertia-vue  (Inertia.js with Vue 3)
    htmx         (HTMX-enhanced templates)
```

Select `inertia-vue` to use Inertia.js with Vue 3.

### Enabling SSR

If you choose Inertia.js, you'll be prompted about Server-Side Rendering:

```
Enable Server-Side Rendering (SSR)? (y/N)
```

SSR improves initial page load performance and SEO but requires a Node.js runtime.

## Project Structure

An Inertia-enabled project includes:

```
your-project/
├── frontend/                    # Frontend source code
│   ├── app.js                  # Client-side entry point
│   ├── ssr.js                  # SSR entry point (if enabled)
│   ├── pages/                  # Vue page components
│   │   ├── Home.vue
│   │   └── Posts/
│   │       ├── Index.vue
│   │       ├── Show.vue
│   │       └── Edit.vue
│   ├── components/             # Reusable Vue components
│   │   ├── Header.vue
│   │   └── Footer.vue
│   └── layouts/                # Layout components
│       └── Layout.vue
├── types/                      # Auto-generated TypeScript types
│   └── models.ts
├── internal/
│   └── handlers/               # Go handlers using Inertia
│       └── post.go
├── public/
│   └── build/                  # Compiled assets
├── esbuild.config.js           # Build configuration
└── package.json                # Frontend dependencies
```

## Backend Integration

### Basic Usage

In your Go handlers, use the Inertia middleware to render pages:

```go
package handlers

import (
    "net/http"
    "github.com/toutaio/toutago-inertia"
)

type PostHandler struct {
    repo *repositories.PostRepository
}

func (h *PostHandler) Index(w http.ResponseWriter, r *http.Request) {
    posts, err := h.repo.FindAll()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Render the Posts/Index.vue component with props
    inertia.Render(w, r, "Posts/Index", map[string]interface{}{
        "posts": posts,
    })
}
```

### Shared Data

Add shared data that's available to all pages:

```go
// In main.go
middleware := inertia.NewMiddleware(inertia.Config{
    Version: "1.0",
    RootView: "app.html",
})

// Add shared data function
middleware.ShareData(func(r *http.Request) map[string]interface{} {
    return map[string]interface{}{
        "auth": map[string]interface{}{
            "user": getCurrentUser(r),
        },
        "flash": getFlashMessages(r),
    }
})
```

### TypeScript Type Generation

Generate TypeScript types from your Go structs:

```bash
npm run types
```

This creates `types/models.ts` with TypeScript interfaces matching your Go models:

```typescript
// Generated from Go structs
export interface Post {
    id: number
    title: string
    content: string
    author: User
    createdAt: string
}
```

## Frontend Development

### Running the Dev Server

Start the frontend build watcher:

```bash
npm run dev
```

This watches for changes and rebuilds automatically.

### Building for Production

```bash
npm run build
```

For SSR-enabled projects:

```bash
npm run build:ssr
```

### Using Inertia in Vue Components

```vue
<template>
  <Layout title="Posts">
    <div class="posts">
      <h1>All Posts</h1>
      <div v-for="post in posts" :key="post.id" class="post-card">
        <Link :href="`/posts/${post.id}`">
          <h2>[[ post.title ]]</h2>
        </Link>
        <p>[[ post.excerpt ]]</p>
      </div>
    </div>
  </Layout>
</template>

<script setup>
import { Link } from '@toutaio/inertia-vue'
import Layout from '../layouts/Layout.vue'

// Props passed from the Go handler
const props = defineProps({
  posts: Array
})
</script>
```

### Navigation

Use the `Link` component for client-side navigation:

```vue
<Link href="/posts" method="get">All Posts</Link>
<Link href="/posts/create" method="get">Create Post</Link>
```

### Forms

Use the `useForm` composable for forms:

```vue
<script setup>
import { useForm } from '@toutaio/inertia-vue'

const form = useForm({
  title: '',
  content: '',
})

function submit() {
  form.post('/posts', {
    onSuccess: () => {
      form.reset()
    },
  })
}
</script>

<template>
  <form @submit.prevent="submit">
    <input v-model="form.title" type="text" />
    <textarea v-model="form.content"></textarea>
    <button type="submit" :disabled="form.processing">
      Save
    </button>
    <div v-if="form.errors.title">[[ form.errors.title ]]</div>
  </form>
</template>
```

## Advanced Features

### Lazy Data Evaluation

For expensive computations, use lazy evaluation:

```go
inertia.Render(w, r, "Dashboard", map[string]interface{}{
    "stats": inertia.Lazy(func() interface{} {
        return computeExpensiveStats()
    }),
})
```

The data is only computed if the component requests it.

### Partial Reloads

Request only specific props:

```javascript
router.reload({ only: ['posts'] })
```

### Asset Versioning

The middleware automatically handles asset versioning for cache busting. Update the version in your config when assets change:

```go
middleware := inertia.NewMiddleware(inertia.Config{
    Version: assetVersion(), // Function that returns current asset hash
})
```

### Error Handling

Inertia automatically handles validation errors:

```go
func (h *PostHandler) Store(w http.ResponseWriter, r *http.Request) {
    var req CreatePostRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // Return validation errors
        inertia.RenderError(w, r, map[string]string{
            "title": "Title is required",
            "content": "Content must be at least 10 characters",
        }, http.StatusUnprocessableEntity)
        return
    }
    
    // ... create post
}
```

## Best Practices

1. **Component Organization**: Keep pages in `pages/`, reusable components in `components/`, and layouts in `layouts/`

2. **Type Safety**: Regularly regenerate TypeScript types with `npm run types` after changing Go models

3. **Performance**: Use lazy data evaluation for expensive operations that aren't always needed

4. **SEO**: Enable SSR for public-facing pages that need good SEO

5. **Asset Management**: Use asset versioning to ensure users get fresh assets after deployments

## Troubleshooting

### Build Errors

If you see build errors:

```bash
# Clear the build directory
rm -rf public/build

# Reinstall dependencies
rm -rf node_modules package-lock.json
npm install

# Rebuild
npm run build
```

### SSR Issues

If SSR isn't working:

1. Check that `ssr.js` exists in `frontend/`
2. Verify the SSR build completed: `npm run build:ssr`
3. Check Go server logs for SSR errors

### Type Generation Issues

If types aren't generating:

1. Ensure your Go models have proper struct tags
2. Check that the types directory exists
3. Run with verbose logging: `go run cmd/main.go generate-types -v`

## Migration from Traditional Templates

To migrate an existing project from traditional templates to Inertia:

1. Install frontend dependencies:
   ```bash
   npm install @toutaio/inertia-vue vue
   npm install -D esbuild esbuild-plugin-vue3
   ```

2. Create the frontend directory structure

3. Convert template handlers to Inertia renders:
   ```go
   // Before
   tmpl.Execute(w, data)
   
   // After
   inertia.Render(w, r, "PageName", data)
   ```

4. Create Vue components from HTML templates

5. Update routes to use the Inertia middleware

6. Set up the build system

See the complete migration guide for detailed steps.

## Resources

- [Inertia.js Documentation](https://inertiajs.com/)
- [Vue 3 Documentation](https://vuejs.org/)
- [toutago-inertia API Reference](https://pkg.go.dev/github.com/toutaio/toutago-inertia)
- [Example Projects](../examples/)

## Getting Help

- GitHub Issues: [toutago-inertia issues](https://github.com/toutaio/toutago-inertia/issues)
- GitHub Discussions: [toutago discussions](https://github.com/toutaio/toutago/discussions)
- Community Discord: [Join our Discord](https://discord.gg/toutago)

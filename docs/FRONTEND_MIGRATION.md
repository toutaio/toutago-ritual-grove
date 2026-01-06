# Frontend Migration Guide

This guide helps you migrate between different frontend approaches in Toutā projects.

## Overview

Toutā supports three frontend approaches:
- **Traditional**: Server-rendered templates using Fíth
- **Inertia.js**: SPA-like experience with Vue.js (React and Svelte coming soon)
- **HTMX**: Modern hypermedia-driven applications

## Migration Paths

### From Traditional to Inertia.js

#### 1. Install Dependencies

```bash
# Add Inertia adapter
go get github.com/toutaio/toutago-inertia

# Initialize npm project (if not already done)
cd your-project
npm init -y

# Install Inertia Vue adapter
npm install @toutaio/inertia-vue vue
npm install --save-dev esbuild @vitejs/plugin-vue
```

#### 2. Update Go Dependencies

Add to your `go.mod`:
```go
require (
    github.com/toutaio/toutago-inertia v0.2.0
)
```

#### 3. Setup Inertia Middleware

In your `main.go` or router setup:

```go
import (
    "github.com/toutaio/toutago-inertia"
    "github.com/toutaio/toutago-cosan-router"
)

func setupRouter() *cosan.Router {
    router := cosan.New()
    
    // Setup Inertia
    inertiaAdapter := inertia.New(inertia.Config{
        RootView:    "app",
        Version:     "1.0.0",
        SSR:         false,
    })
    
    // Add middleware
    router.Use(inertiaAdapter.Middleware())
    
    return router
}
```

#### 4. Convert Handlers

**Before (Traditional):**
```go
func (h *Handler) ListPosts(ctx *cosan.Context) error {
    posts, err := h.postService.List()
    if err != nil {
        return err
    }
    
    return ctx.Render("posts/index", map[string]interface{}{
        "posts": posts,
    })
}
```

**After (Inertia):**
```go
func (h *Handler) ListPosts(ctx *cosan.Context) error {
    posts, err := h.postService.List()
    if err != nil {
        return err
    }
    
    return ctx.Inertia("Posts/Index", map[string]interface{}{
        "posts": posts,
    })
}
```

#### 5. Create Vue Components

Create `frontend/pages/Posts/Index.vue`:

```vue
<template>
  <div>
    <h1>Posts</h1>
    <div v-for="post in posts" :key="post.id">
      <h2>{{ post.title }}</h2>
      <p>{{ post.content }}</p>
    </div>
  </div>
</template>

<script setup>
defineProps({
  posts: Array
})
</script>
```

#### 6. Setup Frontend Build

Create `esbuild.config.js`:

```javascript
const esbuild = require('esbuild')
const vue = require('esbuild-plugin-vue3')

esbuild.build({
  entryPoints: ['frontend/app.js'],
  bundle: true,
  outfile: 'public/js/app.js',
  plugins: [vue()],
  loader: {
    '.js': 'jsx',
  },
}).catch(() => process.exit(1))
```

Create `frontend/app.js`:

```javascript
import { createApp, h } from 'vue'
import { createInertiaApp } from '@toutaio/inertia-vue'

createInertiaApp({
  resolve: name => {
    const pages = import.meta.glob('./pages/**/*.vue', { eager: true })
    return pages[`./pages/${name}.vue`]
  },
  setup({ el, App, props, plugin }) {
    createApp({ render: () => h(App, props) })
      .use(plugin)
      .mount(el)
  },
})
```

#### 7. Update Root Template

Create or update `views/app.fith`:

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .page.props.title || "My App" }}</title>
    <script src="/js/app.js" defer></script>
</head>
<body>
    <div id="app" data-page="{{ .page }}"></div>
</body>
</html>
```

#### 8. Build and Run

```bash
# Build frontend
npm run build

# Run Go server
go run main.go
```

### From Traditional to HTMX

#### 1. Add HTMX

Add to your base template:

```html
<script src="https://unpkg.com/htmx.org@1.9.10"></script>
```

#### 2. Update Templates for Partials

Create partial templates that return HTML fragments:

`views/posts/_post.fith`:
```html
<div class="post" id="post-{{ .id }}">
    <h2>{{ .title }}</h2>
    <p>{{ .content }}</p>
</div>
```

#### 3. Update Handlers

**Before:**
```go
func (h *Handler) ListPosts(ctx *cosan.Context) error {
    posts, err := h.postService.List()
    if err != nil {
        return err
    }
    
    return ctx.Render("posts/index", map[string]interface{}{
        "posts": posts,
    })
}
```

**After:**
```go
func (h *Handler) ListPosts(ctx *cosan.Context) error {
    posts, err := h.postService.List()
    if err != nil {
        return err
    }
    
    // Check if HTMX request
    if ctx.Request.Header.Get("HX-Request") == "true" {
        // Return partial
        return ctx.Render("posts/_list", map[string]interface{}{
            "posts": posts,
        })
    }
    
    // Return full page
    return ctx.Render("posts/index", map[string]interface{}{
        "posts": posts,
    })
}
```

#### 4. Add HTMX Attributes

Update your HTML:

```html
<div id="posts-container">
    <button hx-get="/posts" 
            hx-target="#posts-container" 
            hx-swap="innerHTML">
        Load Posts
    </button>
</div>
```

#### 5. Add Loading States

```html
<div hx-get="/posts" 
     hx-trigger="load"
     hx-indicator=".spinner">
    <div class="spinner" style="display:none;">Loading...</div>
</div>
```

### From Inertia.js to HTMX

#### 1. Remove NPM Dependencies

```bash
npm uninstall @toutaio/inertia-vue vue
```

#### 2. Update Handlers

Replace `ctx.Inertia()` calls with conditional `ctx.Render()` as shown in the Traditional to HTMX section.

#### 3. Convert Vue Components to Templates

**Vue Component:**
```vue
<template>
  <div>
    <h1>{{ post.title }}</h1>
    <p>{{ post.content }}</p>
  </div>
</template>
```

**Fíth Template:**
```html
<div>
    <h1>{{ .post.title }}</h1>
    <p>{{ .post.content }}</p>
</div>
```

#### 4. Add HTMX

Follow steps from "Traditional to HTMX" section above.

## Best Practices

### TypeScript Support

When using Inertia, generate TypeScript types:

```bash
# Add to your hooks
go run cmd/codegen/main.go
```

### Shared Data

Set up shared data that's available on all pages:

```go
inertiaAdapter.Share("auth", func(ctx *cosan.Context) interface{} {
    return ctx.Get("user")
})

inertiaAdapter.Share("flash", func(ctx *cosan.Context) interface{} {
    return ctx.Session.Flash()
})
```

### Error Handling

**Inertia:**
```go
func (h *Handler) CreatePost(ctx *cosan.Context) error {
    var input CreatePostInput
    if err := ctx.Bind(&input); err != nil {
        return ctx.InertiaValidationError(err)
    }
    
    post, err := h.postService.Create(input)
    if err != nil {
        return ctx.InertiaError("Failed to create post", 500)
    }
    
    return ctx.InertiaRedirect("/posts/" + post.ID)
}
```

**HTMX:**
```go
func (h *Handler) CreatePost(ctx *cosan.Context) error {
    var input CreatePostInput
    if err := ctx.Bind(&input); err != nil {
        ctx.Response.Header().Set("HX-Reswap", "innerHTML")
        ctx.Response.Header().Set("HX-Retarget", "#errors")
        return ctx.Render("_errors", map[string]interface{}{
            "errors": err,
        })
    }
    
    post, err := h.postService.Create(input)
    if err != nil {
        return err
    }
    
    ctx.Response.Header().Set("HX-Redirect", "/posts/" + post.ID)
    return nil
}
```

## Performance Considerations

### Inertia.js
- Enable SSR for better initial load times
- Use lazy loading for heavy components
- Implement code splitting
- Cache assets with versioning

### HTMX
- Keep partials small and focused
- Use `hx-boost` for progressive enhancement
- Implement caching headers
- Consider using `hx-swap-oob` for out-of-band updates

## Testing

### Inertia
Test both server and client:

```go
// Server test
func TestListPosts(t *testing.T) {
    req := httptest.NewRequest("GET", "/posts", nil)
    req.Header.Set("X-Inertia", "true")
    
    // assertions...
}
```

```javascript
// Client test (Vitest)
import { mount } from '@vue/test-utils'
import PostsIndex from './PostsIndex.vue'

test('renders posts', () => {
    const wrapper = mount(PostsIndex, {
        props: { posts: [...] }
    })
    expect(wrapper.text()).toContain('Post Title')
})
```

### HTMX
Test with standard HTTP tests:

```go
func TestListPostsPartial(t *testing.T) {
    req := httptest.NewRequest("GET", "/posts", nil)
    req.Header.Set("HX-Request", "true")
    
    // assertions...
}
```

## Troubleshooting

### Inertia Issues

**Problem:** "Inertia asset version mismatch"
**Solution:** Ensure your asset version is consistent and update when deploying

**Problem:** "Component not found"
**Solution:** Check component path in `resolve` function matches file structure

### HTMX Issues

**Problem:** "Updates not triggering"
**Solution:** Check `hx-trigger` attribute and ensure target element exists

**Problem:** "Forms submitting as full page"
**Solution:** Add `hx-boost="true"` or specific `hx-post` attributes

## Resources

- [Inertia.js Official Docs](https://inertiajs.com/)
- [HTMX Official Docs](https://htmx.org/)
- [Toutā Inertia Adapter](https://github.com/toutaio/toutago-inertia)
- [Toutā Fíth Renderer](https://github.com/toutaio/toutago-fith-renderer)

## Support

For questions or issues:
- GitHub Issues: https://github.com/toutaio/toutago/issues
- Discord: https://discord.gg/touta

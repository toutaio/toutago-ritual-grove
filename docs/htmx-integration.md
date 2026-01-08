# HTMX Integration Guide

This guide explains how to use HTMX with Toutago Ritual Grove for building interactive web applications with minimal JavaScript.

## Overview

HTMX allows you to access modern browser features directly from HTML, enabling you to build dynamic user interfaces without writing JavaScript. Toutago's HTMX integration provides enhanced server-rendered templates with progressive enhancement.

## Choosing HTMX

When creating a new ritual project, select `htmx` as your frontend framework:

```
Select frontend framework:
    traditional  (Server-rendered HTML templates)
    inertia-vue  (Inertia.js with Vue 3)
  > htmx         (HTMX-enhanced templates)
```

## Project Structure

An HTMX-enabled project includes:

```
your-project/
├── views/                      # HTML templates
│   ├── layout.html            # Base layout with HTMX
│   ├── partials/              # Partial templates for HTMX
│   │   ├── post-card.html
│   │   └── comment-form.html
│   └── pages/
│       ├── posts/
│       │   ├── index.html
│       │   ├── show.html
│       │   └── edit.html
│       └── home.html
├── static/
│   ├── css/
│   └── js/
│       └── htmx.min.js
└── internal/
    └── handlers/
        ├── post.go
        └── htmx.go             # HTMX-specific handlers
```

## Backend Integration

### Basic HTMX Handler

```go
package handlers

import (
    "net/http"
    "github.com/toutaio/toutago-fith-renderer"
)

type PostHandler struct {
    renderer *fith.Renderer
    repo     *repositories.PostRepository
}

func (h *PostHandler) Index(w http.ResponseWriter, r *http.Request) {
    posts, err := h.repo.FindAll()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Check if this is an HTMX request
    if r.Header.Get("HX-Request") == "true" {
        // Return partial HTML
        h.renderer.RenderPartial(w, "partials/post-list", map[string]interface{}{
            "posts": posts,
        })
        return
    }

    // Return full page
    h.renderer.Render(w, "pages/posts/index", map[string]interface{}{
        "posts": posts,
    })
}
```

### Detecting HTMX Requests

HTMX includes several headers you can check:

```go
func isHTMX(r *http.Request) bool {
    return r.Header.Get("HX-Request") == "true"
}

func isBoosted(r *http.Request) bool {
    return r.Header.Get("HX-Boosted") == "true"
}

func getTarget(r *http.Request) string {
    return r.Header.Get("HX-Target")
}
```

### Responding with HTMX Headers

Set response headers to control HTMX behavior:

```go
func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    
    if err := h.repo.Delete(id); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Tell HTMX to refresh the page
    w.Header().Set("HX-Refresh", "true")
    w.WriteHeader(http.StatusOK)
}
```

## Frontend Templates

### Including HTMX

In your base layout:

```html
<!DOCTYPE html>
<html>
<head>
    <title>[[ .title ]]</title>
    <script src="/static/js/htmx.min.js"></script>
</head>
<body>
    [[ template "content" . ]]
</body>
</html>
```

### Basic HTMX Attributes

**Load content on click:**

```html
<button hx-get="/posts/[[ .id ]]" 
        hx-target="#post-detail" 
        hx-swap="innerHTML">
    View Post
</button>

<div id="post-detail">
    <!-- Content will be loaded here -->
</div>
```

**Submit forms with AJAX:**

```html
<form hx-post="/posts" 
      hx-target="#post-list" 
      hx-swap="afterbegin">
    <input type="text" name="title" placeholder="Post title" />
    <textarea name="content" placeholder="Content"></textarea>
    <button type="submit">Create Post</button>
</form>
```

**Delete with confirmation:**

```html
<button hx-delete="/posts/[[ .id ]]" 
        hx-confirm="Are you sure you want to delete this post?"
        hx-target="closest .post-card"
        hx-swap="outerHTML swap:1s">
    Delete
</button>
```

### Advanced Features

**Infinite scroll:**

```html
<div id="post-list">
    [[ range .posts ]]
        [[ template "partials/post-card" . ]]
    [[ end ]]
</div>

[[ if .hasMore ]]
<div hx-get="/posts?page=[[ .nextPage ]]" 
     hx-trigger="revealed" 
     hx-swap="afterend"
     hx-indicator="#spinner">
    <div id="spinner" class="htmx-indicator">Loading...</div>
</div>
[[ end ]]
```

**Live search:**

```html
<input type="search" 
       name="q" 
       hx-get="/search" 
       hx-trigger="keyup changed delay:500ms" 
       hx-target="#search-results"
       placeholder="Search posts..." />

<div id="search-results">
    <!-- Results appear here as you type -->
</div>
```

**Polling for updates:**

```html
<div hx-get="/notifications/count" 
     hx-trigger="every 30s" 
     hx-swap="innerHTML">
    0 notifications
</div>
```

**Optimistic UI:**

```html
<button hx-post="/likes/[[ .id ]]" 
        hx-swap="outerHTML" 
        class="like-button">
    ♡ Like ([[ .likeCount ]])
</button>
```

## Partial Templates

Create reusable partials for HTMX responses:

```html
{{/* partials/post-card.html */}}
<div class="post-card" id="post-[[ .id ]]">
    <h2>[[ .title ]]</h2>
    <p>[[ .excerpt ]]</p>
    <div class="actions">
        <button hx-get="/posts/[[ .id ]]" 
                hx-target="#modal" 
                hx-swap="innerHTML">
            Read More
        </button>
        <button hx-delete="/posts/[[ .id ]]" 
                hx-target="closest .post-card" 
                hx-swap="outerHTML">
            Delete
        </button>
    </div>
</div>
```

## Best Practices

### 1. Progressive Enhancement

Ensure forms work without JavaScript:

```html
<form action="/posts" method="POST" 
      hx-post="/posts" 
      hx-target="#posts" 
      hx-swap="afterbegin">
    <!-- Form fields -->
    <button type="submit">Create Post</button>
</form>
```

### 2. Use Semantic HTML

```html
<!-- Good: Semantic and accessible -->
<article class="post" id="post-[[ .id ]]">
    <header>
        <h2>[[ .title ]]</h2>
    </header>
    <section class="content">
        [[ .content ]]
    </section>
</article>

<!-- Avoid: Non-semantic divs -->
<div class="post">
    <div class="title">[[ .title ]]</div>
    <div>[[ .content ]]</div>
</div>
```

### 3. Handle Errors Gracefully

```go
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
    var post Post
    if err := parseForm(r, &post); err != nil {
        // Return error partial
        w.WriteHeader(http.StatusUnprocessableEntity)
        h.renderer.RenderPartial(w, "partials/error", map[string]interface{}{
            "message": "Invalid post data",
            "errors": err,
        })
        return
    }
    
    // ... create post
}
```

### 4. Use Indicators for Loading States

```html
<style>
    .htmx-indicator {
        display: none;
    }
    .htmx-request .htmx-indicator,
    .htmx-request.htmx-indicator {
        display: inline;
    }
</style>

<button hx-post="/posts" hx-indicator="#spinner">
    Submit
    <span id="spinner" class="htmx-indicator">⏳</span>
</button>
```

### 5. Validate on Both Sides

Always validate on the server, even with client-side validation:

```html
<form hx-post="/posts">
    <input type="email" 
           name="email" 
           required 
           pattern="[^@]+@[^@]+\.[^@]+" />
</form>
```

```go
func validateEmail(email string) error {
    // Server-side validation
    if !regexp.MustCompile(`^[^@]+@[^@]+\.[^@]+$`).MatchString(email) {
        return errors.New("invalid email format")
    }
    return nil
}
```

## Common Patterns

### Modal Dialogs

```html
<!-- Trigger -->
<button hx-get="/posts/[[ .id ]]/edit" 
        hx-target="#modal" 
        hx-swap="innerHTML">
    Edit
</button>

<!-- Modal container -->
<div id="modal" class="modal hidden">
    <!-- Content loaded here -->
</div>
```

### Tabs

```html
<div class="tabs">
    <button hx-get="/posts" 
            hx-target="#content" 
            class="active">Posts</button>
    <button hx-get="/comments" 
            hx-target="#content">Comments</button>
</div>

<div id="content">
    <!-- Tab content -->
</div>
```

### Inline Editing

```html
<div id="post-title-[[ .id ]]">
    <span>[[ .title ]]</span>
    <button hx-get="/posts/[[ .id ]]/edit/title" 
            hx-target="#post-title-[[ .id ]]" 
            hx-swap="outerHTML">
        Edit
    </button>
</div>
```

When editing:

```html
<form hx-put="/posts/[[ .id ]]/title" 
      hx-target="#post-title-[[ .id ]]" 
      hx-swap="outerHTML">
    <input type="text" name="title" value="[[ .title ]]" />
    <button type="submit">Save</button>
    <button hx-get="/posts/[[ .id ]]/title" 
            hx-target="#post-title-[[ .id ]]" 
            hx-swap="outerHTML">
        Cancel
    </button>
</form>
```

## Extensions

HTMX supports several extensions:

### Class Tools

Add/remove/toggle CSS classes:

```html
<script src="https://unpkg.com/htmx.org/dist/ext/class-tools.js"></script>

<div hx-ext="class-tools">
    <button hx-post="/toggle" 
            classes="add highlight:1s">
        Click me
    </button>
</div>
```

### Morphdom (DOM Morphing)

Smooth updates by morphing the DOM:

```html
<script src="https://unpkg.com/htmx.org/dist/ext/morphdom-swap.js"></script>

<div hx-ext="morphdom-swap" 
     hx-get="/updates" 
     hx-swap="morphdom">
    <!-- Content -->
</div>
```

## Debugging

Enable HTMX logging:

```html
<script>
    htmx.logger = function(elt, event, data) {
        if(console) {
            console.log("HTMX:", event, elt, data);
        }
    }
</script>
```

## Resources

- [HTMX Documentation](https://htmx.org/)
- [HTMX Examples](https://htmx.org/examples/)
- [Hypermedia Systems Book](https://hypermedia.systems/)
- [Example Projects](../examples/)

## Getting Help

- GitHub Issues: [toutago-ritual-grove issues](https://github.com/toutaio/toutago-ritual-grove/issues)
- GitHub Discussions: [toutago discussions](https://github.com/toutaio/toutago/discussions)
- Community Discord: [Join our Discord](https://discord.gg/toutago)

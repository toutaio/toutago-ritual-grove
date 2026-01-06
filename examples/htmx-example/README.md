# HTMX Example

This example demonstrates building a modern web application with Toutā and HTMX for progressive enhancement.

## Features

- 🔄 **SPA-like Experience** without JavaScript frameworks
- 📡 **Partial Page Updates** with HTMX
- 🎯 **Progressive Enhancement** - works without JS
- ⚡ **Fast** - minimal JavaScript payload
- 📱 **Mobile Friendly**
- 🎨 **Server-Side Rendering** with Fíth templates
- 🔍 **SEO Friendly**
- ♿ **Accessible** by default

## Quick Start

```bash
# Generate from ritual
touta ritual init blog

# Choose:
# - Frontend: htmx
# - Templates: fith
# - Database: postgres
```

## Project Structure

```
htmx-example/
├── handlers/
│   ├── posts.go              # Blog post handlers
│   ├── comments.go           # Comment handlers (partial renders)
│   └── search.go             # Search with live results
├── views/
│   ├── layouts/
│   │   └── base.html         # Base layout
│   ├── posts/
│   │   ├── index.html        # Post list
│   │   ├── show.html         # Single post
│   │   ├── _post_card.html   # Partial: post card
│   │   └── _comment.html     # Partial: comment
│   ├── partials/
│   │   ├── _search.html      # Search results partial
│   │   └── _pagination.html  # Pagination partial
│   └── components/
│       ├── _header.html
│       └── _footer.html
└── static/
    ├── css/
    │   └── app.css           # Tailwind CSS
    └── js/
        └── htmx.min.js       # HTMX library
```

## Key Implementations

### 1. Post List with Infinite Scroll

```html
<!-- views/posts/index.html -->
<% extends "layouts/base.html" %>

<% block content %>
<div class="container mx-auto px-4">
  <h1 class="text-3xl font-bold mb-8">Blog Posts</h1>
  
  <!-- Search -->
  <div class="mb-6">
    <input
      type="search"
      name="q"
      placeholder="Search posts..."
      class="w-full p-3 border rounded"
      hx-post="/search"
      hx-trigger="keyup changed delay:500ms"
      hx-target="#search-results"
      hx-indicator=".spinner"
    />
    <div class="spinner htmx-indicator">Searching...</div>
  </div>
  
  <div id="search-results">
    <!-- Posts list -->
    <div id="posts-list">
      <% for post in posts %>
        <% include "posts/_post_card.html" with post=post %>
      <% endfor %>
    </div>
    
    <!-- Infinite scroll trigger -->
    <% if has_more %>
      <div
        hx-get="/posts?page=<%= next_page %>"
        hx-trigger="revealed"
        hx-swap="afterend"
        class="text-center py-4"
      >
        <div class="spinner">Loading more...</div>
      </div>
    <% endif %>
  </div>
</div>
<% endblock %>
```

```go
// handlers/posts.go
func (h *PostHandler) Index(c *cosan.Context) error {
    page := c.QueryInt("page", 1)
    perPage := 10
    
    posts, total, err := h.db.GetPosts(page, perPage)
    if err != nil {
        return err
    }
    
    hasMore := page * perPage < total
    
    // Return partial for HTMX requests
    if c.Header("HX-Request") != "" {
        return c.Render("posts/_post_list.html", map[string]interface{}{
            "posts": posts,
            "has_more": hasMore,
            "next_page": page + 1,
        })
    }
    
    // Return full page for normal requests
    return c.Render("posts/index.html", map[string]interface{}{
        "posts": posts,
        "has_more": hasMore,
        "next_page": page + 1,
    })
}
```

### 2. Post Card Partial

```html
<!-- views/posts/_post_card.html -->
<article class="bg-white rounded-lg shadow p-6 mb-4 hover:shadow-lg transition">
  <h2 class="text-2xl font-bold mb-2">
    <a href="/posts/<%= post.id %>" class="hover:text-blue-600">
      <%= post.title %>
    </a>
  </h2>
  
  <p class="text-gray-600 mb-4"><%= post.excerpt %></p>
  
  <div class="flex items-center justify-between">
    <span class="text-sm text-gray-500">
      <%= post.author %> • <%= post.created_at | date:"Jan 2, 2006" %>
    </span>
    
    <!-- Like button with optimistic UI -->
    <button
      hx-post="/posts/<%= post.id %>/like"
      hx-target="this"
      hx-swap="outerHTML"
      class="flex items-center gap-2 text-gray-600 hover:text-red-600"
    >
      <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
        <path d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z"/>
      </svg>
      <span><%= post.likes %></span>
    </button>
  </div>
</article>
```

### 3. Live Search

```go
// handlers/search.go
func (h *SearchHandler) Search(c *cosan.Context) error {
    query := c.FormValue("q")
    
    if query == "" {
        // Return empty results
        return c.Render("partials/_search.html", map[string]interface{}{
            "results": []Post{},
        })
    }
    
    results, err := h.db.SearchPosts(query)
    if err != nil {
        return err
    }
    
    return c.Render("partials/_search.html", map[string]interface{}{
        "results": results,
        "query": query,
    })
}
```

```html
<!-- views/partials/_search.html -->
<% if len(results) > 0 %>
  <div class="bg-white rounded shadow-lg max-h-96 overflow-y-auto">
    <% for result in results %>
      <a
        href="/posts/<%= result.id %>"
        class="block p-4 hover:bg-gray-50 border-b"
      >
        <h3 class="font-semibold"><%= result.title %></h3>
        <p class="text-sm text-gray-600"><%= result.excerpt %></p>
      </a>
    <% endfor %>
  </div>
<% else if query %>
  <div class="p-4 text-gray-600">
    No results found for "<%= query %>"
  </div>
<% endif %>
```

### 4. Comments with HTMX

```html
<!-- views/posts/show.html -->
<% extends "layouts/base.html" %>

<% block content %>
<article class="container mx-auto px-4 max-w-3xl">
  <h1 class="text-4xl font-bold mb-4"><%= post.title %></h1>
  <div class="prose max-w-none mb-8">
    <%= post.content | safe %>
  </div>
  
  <!-- Comments Section -->
  <section id="comments" class="mt-12">
    <h2 class="text-2xl font-bold mb-6">
      Comments (<span id="comment-count"><%= len(comments) %></span>)
    </h2>
    
    <!-- Comment Form -->
    <form
      hx-post="/posts/<%= post.id %>/comments"
      hx-target="#comments-list"
      hx-swap="afterbegin"
      hx-on::after-request="this.reset()"
      class="mb-8 bg-gray-50 p-4 rounded"
    >
      <textarea
        name="content"
        placeholder="Add a comment..."
        class="w-full p-3 border rounded mb-2"
        rows="3"
        required
      ></textarea>
      <button
        type="submit"
        class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
      >
        Post Comment
      </button>
    </form>
    
    <!-- Comments List -->
    <div id="comments-list">
      <% for comment in comments %>
        <% include "posts/_comment.html" with comment=comment %>
      <% endfor %>
    </div>
  </section>
</article>
<% endblock %>
```

```go
// handlers/comments.go
func (h *CommentHandler) Create(c *cosan.Context) error {
    postID := c.ParamInt("postID")
    content := c.FormValue("content")
    user := c.Get("user").(*models.User)
    
    comment, err := h.db.CreateComment(postID, user.ID, content)
    if err != nil {
        return err
    }
    
    // Return just the comment partial
    return c.Render("posts/_comment.html", map[string]interface{}{
        "comment": comment,
    })
}

func (h *CommentHandler) Delete(c *cosan.Context) error {
    commentID := c.ParamInt("id")
    
    err := h.db.DeleteComment(commentID)
    if err != nil {
        return err
    }
    
    // Return empty response to remove element
    c.Response().Header().Set("HX-Trigger", "commentDeleted")
    return c.NoContent(200)
}
```

```html
<!-- views/posts/_comment.html -->
<div id="comment-<%= comment.id %>" class="border-l-4 border-blue-500 pl-4 mb-4">
  <div class="flex items-start justify-between">
    <div>
      <strong><%= comment.author %></strong>
      <span class="text-sm text-gray-500 ml-2">
        <%= comment.created_at | timeago %>
      </span>
    </div>
    
    <% if current_user and current_user.id == comment.user_id %>
      <button
        hx-delete="/comments/<%= comment.id %>"
        hx-target="#comment-<%= comment.id %>"
        hx-swap="outerHTML swap:1s"
        hx-confirm="Delete this comment?"
        class="text-red-600 text-sm"
      >
        Delete
      </button>
    <% endif %>
  </div>
  
  <p class="mt-2"><%= comment.content %></p>
  
  <!-- Reply button -->
  <button
    hx-get="/comments/<%= comment.id %>/reply-form"
    hx-target="#comment-<%= comment.id %>"
    hx-swap="afterend"
    class="text-sm text-blue-600 mt-2"
  >
    Reply
  </button>
</div>
```

### 5. Modal with HTMX

```html
<!-- Edit post modal trigger -->
<button
  hx-get="/posts/<%= post.id %>/edit"
  hx-target="#modal"
  hx-swap="innerHTML"
  class="btn-primary"
>
  Edit Post
</button>

<!-- Modal container -->
<div id="modal"></div>
```

```go
// handlers/posts.go
func (h *PostHandler) Edit(c *cosan.Context) error {
    postID := c.ParamInt("id")
    post, err := h.db.GetPost(postID)
    if err != nil {
        return err
    }
    
    return c.Render("posts/_edit_modal.html", map[string]interface{}{
        "post": post,
    })
}
```

```html
<!-- views/posts/_edit_modal.html -->
<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
  <div class="bg-white rounded-lg p-6 w-full max-w-2xl">
    <h2 class="text-2xl font-bold mb-4">Edit Post</h2>
    
    <form
      hx-put="/posts/<%= post.id %>"
      hx-target="#post-<%= post.id %>"
      hx-swap="outerHTML"
    >
      <input
        type="text"
        name="title"
        value="<%= post.title %>"
        class="w-full p-3 border rounded mb-4"
        required
      />
      
      <textarea
        name="content"
        class="w-full p-3 border rounded mb-4"
        rows="10"
        required
      ><%= post.content %></textarea>
      
      <div class="flex gap-2">
        <button type="submit" class="btn-primary">
          Save Changes
        </button>
        <button
          type="button"
          onclick="this.closest('#modal').innerHTML=''"
          class="btn-secondary"
        >
          Cancel
        </button>
      </div>
    </form>
  </div>
</div>
```

### 6. Base Layout with HTMX

```html
<!-- views/layouts/base.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title><% block title %><%= site_name %><% endblock %></title>
  <link rel="stylesheet" href="/static/css/app.css">
  <script src="/static/js/htmx.min.js"></script>
</head>
<body>
  <% include "components/_header.html" %>
  
  <main>
    <!-- Flash messages -->
    <div id="flash-messages" class="container mx-auto px-4 mt-4">
      <% if flash_success %>
        <div class="bg-green-100 border-green-500 text-green-700 p-4 rounded mb-4">
          <%= flash_success %>
        </div>
      <% endif %>
    </div>
    
    <% block content %><% endblock %>
  </main>
  
  <% include "components/_footer.html" %>
  
  <!-- Global HTMX config -->
  <script>
    // Show loading indicator
    document.body.addEventListener('htmx:beforeRequest', () => {
      document.body.classList.add('loading')
    })
    
    document.body.addEventListener('htmx:afterRequest', () => {
      document.body.classList.remove('loading')
    })
    
    // Handle errors
    document.body.addEventListener('htmx:responseError', (e) => {
      alert('An error occurred. Please try again.')
    })
  </script>
</body>
</html>
```

## HTMX Features Used

### Attributes
- `hx-get`, `hx-post`, `hx-put`, `hx-delete` - HTTP methods
- `hx-target` - Where to put response
- `hx-swap` - How to swap content
- `hx-trigger` - When to trigger request
- `hx-indicator` - Loading indicator
- `hx-confirm` - Confirmation dialog

### Advanced Features
- Infinite scroll with `hx-trigger="revealed"`
- Live search with `delay:500ms`
- Optimistic UI updates
- Out-of-band swaps with `hx-swap-oob`
- Polling with `hx-trigger="every 2s"`

## Progressive Enhancement

Works without JavaScript:
```html
<form action="/posts" method="POST" hx-post="/posts" hx-target="#posts-list">
  <!-- Falls back to traditional form submission if HTMX fails -->
</form>
```

## Running the Application

```bash
# Install dependencies
go mod tidy

# Start server
go run main.go
```

Visit http://localhost:3000

## Benefits of HTMX

- 📦 **Small**: ~14KB gzipped
- 🚀 **Fast**: Minimal JavaScript
- ♿ **Accessible**: Server-rendered HTML
- 🔍 **SEO**: Full HTML responses
- 💪 **Simple**: No build step needed
- 🎯 **Progressive**: Works without JS

## License

MIT

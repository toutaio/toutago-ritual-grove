# Blog with Inertia.js Example

This example demonstrates a full-featured blog application using Toutā with Inertia.js and Vue 3.

## Features

- 📝 **Full CRUD** for blog posts
- 🏷️ **Categories** and tags
- 💬 **Comments** system
- 👤 **User authentication** with sessions
- 🔍 **Search** functionality
- 📱 **Responsive** Vue components
- ⚡ **SPA experience** with server-side routing
- 🎨 **Modern UI** with Tailwind CSS

## Generated with

```bash
touta ritual init blog
```

Answer the questionnaire:
- Frontend: `inertia-vue`
- Enable SSR: `yes`
- Database: `postgres`
- Authentication: `yes`

## Project Structure

```
blog-inertia/
├── main.go                    # Application entry point
├── go.mod                     # Go dependencies
├── handlers/                  # HTTP handlers
│   ├── posts.go              # Blog post CRUD
│   ├── comments.go           # Comment management
│   └── auth.go               # Authentication
├── models/                    # Data models
│   ├── post.go
│   ├── comment.go
│   └── user.go
├── migrations/                # Database migrations
│   ├── 001_create_posts.sql
│   ├── 002_create_comments.sql
│   └── 003_create_users.sql
├── resources/                 # Frontend resources
│   ├── js/
│   │   ├── app.js            # Inertia app entry
│   │   ├── Pages/            # Vue pages
│   │   │   ├── Home.vue
│   │   │   ├── Posts/
│   │   │   │   ├── Index.vue
│   │   │   │   ├── Show.vue
│   │   │   │   ├── Create.vue
│   │   │   │   └── Edit.vue
│   │   │   └── Auth/
│   │   │       ├── Login.vue
│   │   │       └── Register.vue
│   │   └── Components/
│   │       ├── Layout.vue
│   │       ├── Header.vue
│   │       └── CommentList.vue
│   └── css/
│       └── app.css           # Tailwind CSS
├── package.json              # npm dependencies
├── esbuild.config.js         # Build configuration
└── public/                   # Static assets
    └── dist/                 # Built assets
```

## Getting Started

### 1. Install Dependencies

```bash
# Go dependencies
go mod tidy

# Node dependencies
npm install
```

### 2. Setup Database

```bash
# Create database
createdb myblog

# Run migrations
touta migrate up
```

### 3. Build Frontend Assets

```bash
# Development mode with watch
npm run dev

# Production build
npm run build
```

### 4. Start the Server

```bash
# Development mode
go run main.go

# Or using touta CLI
touta serve
```

Visit http://localhost:3000

## Key Code Examples

### Inertia Handler (Go)

```go
func (h *PostHandler) Index(c *cosan.Context) error {
    posts, err := h.db.GetAllPosts()
    if err != nil {
        return err
    }
    
    return c.Inertia("Posts/Index", map[string]interface{}{
        "posts": posts,
        "user": c.Get("user"),
    })
}
```

### Vue Page Component

```vue
<template>
  <Layout>
    <h1>Blog Posts</h1>
    <div v-for="post in posts" :key="post.id">
      <Link :href="`/posts/${post.id}`">
        <h2>{{ post.title }}</h2>
      </Link>
      <p>{{ post.excerpt }}</p>
    </div>
  </Layout>
</template>

<script setup>
import { Link } from '@toutaio/inertia-vue'
import Layout from '../Components/Layout.vue'

defineProps({
  posts: Array
})
</script>
```

### Form Handling

```vue
<template>
  <form @submit.prevent="submit">
    <input v-model="form.title" type="text" />
    <textarea v-model="form.content"></textarea>
    <button type="submit" :disabled="form.processing">
      Publish
    </button>
  </form>
</template>

<script setup>
import { useForm } from '@toutaio/inertia-vue'

const form = useForm({
  title: '',
  content: ''
})

const submit = () => {
  form.post('/posts', {
    onSuccess: () => {
      form.reset()
    }
  })
}
</script>
```

## Features in Detail

### Server-Side Rendering (SSR)

SSR is enabled, providing:
- ✅ SEO-friendly pages
- ✅ Faster initial page load
- ✅ Social media previews
- ✅ Progressive enhancement

### Authentication

Uses session-based auth with:
- Login/Register pages
- Password hashing with bcrypt
- CSRF protection
- Session middleware

### Database

PostgreSQL with:
- Migration system
- Prepared statements
- Connection pooling
- Transaction support

### TypeScript Support

TypeScript types are generated from Go structs:

```bash
# Generate types
touta inertia generate-types
```

Creates `resources/js/types.ts`:

```typescript
export interface Post {
  id: number
  title: string
  content: string
  author: User
  createdAt: string
}
```

## Development Workflow

### 1. Make changes to Go handlers

```go
// handlers/posts.go
func (h *PostHandler) Show(c *cosan.Context) error {
    // Add new data
    relatedPosts, _ := h.db.GetRelatedPosts(postID)
    
    return c.Inertia("Posts/Show", map[string]interface{}{
        "post": post,
        "related": relatedPosts, // New data
    })
}
```

### 2. Update Vue component

```vue
<script setup>
defineProps({
  post: Object,
  related: Array // New prop
})
</script>
```

### 3. Generate TypeScript types

```bash
touta inertia generate-types
```

### 4. Hot reload picks up changes

Both Go (via air) and Vue (via esbuild watch) auto-reload!

## Testing

```bash
# Run Go tests
go test ./...

# Run Vue component tests
npm test

# Run E2E tests
npm run test:e2e
```

## Deployment

### Build for production

```bash
# Build frontend
npm run build

# Build Go binary
go build -o blog main.go
```

### Environment variables

```bash
export DATABASE_URL=postgres://user:pass@localhost/myblog
export SESSION_SECRET=your-secret-key
export PORT=3000
```

### Run

```bash
./blog
```

## Learn More

- [Toutā Documentation](https://github.com/toutaio/toutago)
- [Inertia.js Guide](https://inertiajs.com)
- [Vue 3 Documentation](https://vuejs.org)
- [Cosan Router](https://github.com/toutaio/toutago-cosan-router)

## License

MIT

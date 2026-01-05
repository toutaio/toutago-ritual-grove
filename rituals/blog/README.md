# Blog Ritual

A comprehensive blog ritual for Toutā that generates a full-featured blog application with posts, categories, and optional comments.

## Features

- **Post Management**: Full CRUD operations for blog posts
- **Categories**: Organize posts by category
- **Comments**: Optional comment system with moderation
- **Markdown Support**: Optional Markdown rendering for posts
- **RESTful API**: Clean API endpoints for all operations
- **Responsive Design**: Mobile-friendly CSS included
- **Database Migrations**: Schema migrations included
- **Multi-database**: Supports PostgreSQL and MySQL

## Usage

Initialize a new blog project:

```bash
touta ritual init blog
```

You'll be prompted for:
- Blog name
- Go module path
- Server port
- Database configuration (type, host, port, name, credentials)
- Comment system (enable/disable)
- Posts per page
- Markdown support (enable/disable)

## Generated Structure

```
myblog/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── internal/
│   ├── models/
│   │   ├── post.go           # Post model
│   │   ├── category.go       # Category model
│   │   └── comment.go        # Comment model (if enabled)
│   └── handlers/
│       ├── post.go           # Post handlers
│       ├── category.go       # Category handlers
│       └── comment.go        # Comment handlers (if enabled)
├── views/
│   ├── layout.html           # Base HTML layout
│   ├── post_list.html        # Posts listing
│   ├── post_detail.html      # Single post view
│   └── category_list.html    # Categories listing
├── public/
│   └── css/
│       └── style.css         # Blog styles
├── migrations/
│   ├── 001_initial_schema.sql       # Up migration
│   └── 001_initial_schema_down.sql  # Down migration
├── .env.example              # Environment variables template
└── README.md                 # Project documentation
```

## API Endpoints

### Posts
- `GET /` - Homepage (posts list)
- `GET /posts` - List all posts
- `GET /posts/:id` - Get a specific post
- `POST /posts` - Create a new post
- `PUT /posts/:id` - Update a post
- `DELETE /posts/:id` - Delete a post

### Categories
- `GET /categories` - List all categories
- `GET /categories/:id` - Get a specific category

### Comments (if enabled)
- `POST /posts/:id/comments` - Add a comment to a post
- `DELETE /comments/:id` - Delete a comment

## Database Schema

### posts
- `id` - Primary key
- `title` - Post title
- `slug` - URL-friendly slug
- `content` - Post content
- `excerpt` - Short excerpt
- `category_id` - Foreign key to categories
- `author_id` - Author identifier
- `status` - draft | published | archived
- `published_at` - Publication timestamp
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

### categories
- `id` - Primary key
- `name` - Category name
- `slug` - URL-friendly slug
- `description` - Category description
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

### comments (optional)
- `id` - Primary key
- `post_id` - Foreign key to posts
- `author` - Commenter name
- `email` - Commenter email
- `content` - Comment content
- `status` - pending | approved | spam
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

## Configuration

The ritual generates a `.env.example` file with all configuration options:

```env
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=blog_db
DB_USER=blog_user
DB_PASSWORD=your_password

PORT=8080
APP_ENV=development

POSTS_PER_PAGE=10
ENABLE_COMMENTS=true
ENABLE_MARKDOWN=true
```

## Post-Installation Steps

1. Copy `.env.example` to `.env` and configure your database
2. Run the migration: Execute `migrations/001_initial_schema.sql`
3. Install dependencies: `go mod tidy`
4. Start the server: `go run main.go`

## Customization

The generated code provides a solid foundation. Common customizations:

- **Add authentication**: Integrate toutago-breitheamh-auth for user management
- **Enhance models**: Add more fields like tags, featured images, SEO metadata
- **Add search**: Implement full-text search for posts
- **RSS feed**: Add RSS/Atom feed generation
- **Admin panel**: Create an admin interface for content management
- **Rich editor**: Integrate a WYSIWYG editor
- **Image upload**: Add support for post images

## Requirements

- Go 1.21 or higher
- PostgreSQL or MySQL database
- Toutā framework components:
  - toutago-cosan-router
  - toutago-fith-renderer
  - toutago-datamapper
  - toutago-nasc-dependency-injector

## License

MIT

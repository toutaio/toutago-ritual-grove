# Blog Ritual for Toutā

A full-featured, production-ready blog system ritual for the Toutā Go framework.

## Features

### Core Functionality
- ✅ **Complete Authentication System**
  - First-user admin setup
  - Login/Register/Logout
  - Session management with Breitheamh
  - Password hashing and security

- ✅ **Role-Based Access Control (RBAC)**
  - 4 user roles: Admin, Editor, Author, User
  - Granular permissions system
  - Resource-level access control
  - UI elements adapt to user permissions

- ✅ **Post Management**
  - Create, Read, Update, Delete (CRUD)
  - Draft, Published, Archived statuses
  - Markdown editor (EasyMDE)
  - Featured images
  - Category assignment
  - Slug generation
  - Bulk actions (delete, publish, archive)

- ✅ **Category Management**
  - Full CRUD operations
  - Slug auto-generation
  - Post count tracking
  - Public category archives

- ✅ **User Management**
  - Admin user list
  - Role assignment
  - User activation/deactivation
  - Profile editing

- ✅ **Admin Dashboard**
  - Statistics overview (posts, users, categories)
  - Quick actions
  - Activity monitoring
  - Role-based interface

### Technical Features
- ✅ **Multi-Database Support** - PostgreSQL and MySQL
- ✅ **Comprehensive Testing** - 120+ test cases (TDD)
- ✅ **Error Handling** - Professional error pages (404, 403, 500)
- ✅ **Security** - Multi-layer permission checks, input validation
- ✅ **Modern UI** - Responsive design, professional admin interface

## Installation

### Prerequisites
- Go 1.21 or higher
- PostgreSQL or MySQL database
- Toutā framework installed

### Quick Start

1. **Create a new blog project:**
```bash
touta perform blog
```

2. **Answer the setup questions:**
- Blog name
- Go module path
- Port number
- Database type
- Enable Docker

3. **Set up and run:**
```bash
cd myblog
go run main.go migrate up
go run main.go serve
```

4. **Access the setup page:**
Open `http://localhost:8080/auth/setup` and create your first admin user.

## User Roles & Permissions

| Role | Create Post | Edit Own | Edit Any | Delete Own | Delete Any | Publish | Manage Users | Manage Categories |
|------|------------|----------|----------|------------|------------|---------|--------------|-------------------|
| **Admin** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Editor** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| **Author** | ✅ | ✅ | ❌ | ✅ | ❌ | Own only | ❌ | ❌ |
| **User** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

## Project Structure

```
myblog/
├── internal/
│   ├── domain/          # Domain models (User, Post, Category)
│   ├── dto/             # Data Transfer Objects with validation
│   ├── services/        # Business logic layer
│   ├── repositories/    # Data access layer
│   ├── handlers/        # HTTP handlers
│   │   └── admin/       # Admin-specific handlers
│   └── helpers/         # Template helpers
├── views/               # HTML templates
│   ├── auth/           # Authentication pages
│   ├── admin/          # Admin interface
│   ├── posts/          # Public blog views
│   └── errors/         # Error pages
├── public/             # Static assets
├── migrations/         # Database migrations
└── main.go            # Application entry point
```

## Usage

### Creating a Post

1. Login to your account
2. Navigate to **Admin → Posts → Create New Post**
3. Fill in the details (title, content, category, status)
4. Click **Save Post**

### Bulk Actions

1. Go to **Admin → Posts**
2. Select multiple posts using checkboxes
3. Choose an action (Publish/Archive/Delete)
4. Confirm the action

## Development

### Running Tests

```bash
go test ./...              # Run all tests
go test -cover ./...       # With coverage
go test -v ./...           # Verbose output
```

### Database Migrations

```bash
go run main.go migrate up      # Run migrations
go run main.go migrate down    # Rollback migrations
```

## Architecture

The blog follows Clean Architecture principles:

```
Handlers → Services → Repositories → Domain
```

- **Handlers**: HTTP request/response
- **Services**: Business logic & permissions
- **Repositories**: Database access
- **Domain**: Core entities

## License

MIT License

## Support

- **Documentation**: https://touta.io/docs
- **Issues**: https://github.com/toutaio/toutago-ritual-grove/issues

---

Built with ❤️ by the Toutā Team

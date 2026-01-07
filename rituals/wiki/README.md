# Wiki Ritual

A complete wiki/knowledge base application with version control, markdown support, and full-text search.

## Features

- **Markdown Support**: Write pages using markdown syntax
- **Version Control**: Track all changes with complete revision history
- **Full-Text Search**: Find pages quickly with PostgreSQL full-text search
- **Tagging System**: Organize pages with tags (optional)
- **File Attachments**: Upload files to wiki pages (optional)
- **Revision Limits**: Configure maximum revisions to keep per page

## What's Included

- Page models with slug-based URLs
- Revision tracking for version control
- Tag support for categorization
- Markdown rendering with Goldmark
- Full-text search capabilities
- Clean, responsive UI
- Auto-save draft functionality

## Usage

Initialize a new wiki project:

```bash
touta ritual init wiki
```

Answer the interactive prompts:
- **Wiki Name**: The name of your wiki
- **Module Path**: Go module path (e.g., github.com/yourorg/wiki)
- **Port**: Server port (default: 8080)
- **Database**: Choose postgres or mysql
- **Enable Search**: Full-text search feature
- **Enable Tags**: Page tagging system
- **Enable Attachments**: File upload support
- **Max Revisions**: Maximum revisions to keep (0 = unlimited)

## Database Setup

The wiki requires a database. Create one before starting:

```bash
# PostgreSQL
createdb mywiki

# MySQL
mysql -e "CREATE DATABASE mywiki"
```

Set database connection in `.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mywiki
DB_USER=postgres
DB_PASSWORD=yourpassword
```

## Running

```bash
go run main.go
```

Visit `http://localhost:8080` to access your wiki.

## Creating Pages

1. Click "New Page"
2. Enter title and content in Markdown
3. Add your name as author
4. Optionally add a change comment
5. Click "Create Page"

## Editing Pages

1. Navigate to a page
2. Click "Edit"
3. Modify content
4. Add change comment
5. Click "Update Page"

All changes are saved in the revision history.

## Viewing History

Click "History" on any page to see all revisions. Each revision shows:
- Version number
- Author
- Timestamp
- Change comment

## Search

{{if .enable_search}}
Use the search box in the header to find pages by title or content.
{{else}}
Search is disabled. Enable it during ritual initialization to use this feature.
{{end}}

## Customization

- **CSS**: Edit `static/wiki.css` for styling
- **JavaScript**: Edit `static/wiki.js` for client-side features
- **Templates**: Modify files in `views/` directory
- **Models**: Extend models in `models/` directory

## Architecture

- **Toutā Framework**: Web framework and routing
- **Cosan Router**: HTTP routing
- **Fíth Renderer**: Template rendering
- **Datamapper**: Database ORM
- **Goldmark**: Markdown parser

## License

Generated from toutago-ritual-grove wiki ritual.

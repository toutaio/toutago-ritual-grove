package generator

import (
	"fmt"
	"strings"
)

// DatabaseGenerator handles database-specific code generation
type DatabaseGenerator struct{}

// NewDatabaseGenerator creates a new database code generator
func NewDatabaseGenerator() *DatabaseGenerator {
	return &DatabaseGenerator{}
}

// DBConnectionConfig holds database connection configuration
type DBConnectionConfig struct {
	Type     string // "mysql" or "postgres"
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string // For PostgreSQL
}

// TableSchema defines a database table structure
type TableSchema struct {
	Name    string
	Columns []ColumnSchema
}

// ColumnSchema defines a table column
type ColumnSchema struct {
	Name          string
	Type          string // "int", "string", "timestamp", "text", "bool"
	Size          int    // For string types
	PrimaryKey    bool
	AutoIncrement bool
	NotNull       bool
	Unique        bool
	Default       string
}

// QuerySpec defines a database query
type QuerySpec struct {
	Operation string // "select", "insert", "update", "delete"
	Table     string
	Columns   []string
	Where     map[string]string
	OrderBy   string
	Limit     int
}

// ModelSpec defines a data model
type ModelSpec struct {
	Name   string
	Table  string
	Fields []FieldSpec
}

// FieldSpec defines a model field
type FieldSpec struct {
	Name     string
	Type     string
	DBColumn string
	Tags     string
}

// GenerateConnectionCode generates database connection code
func (g *DatabaseGenerator) GenerateConnectionCode(config DBConnectionConfig) string {
	var code strings.Builder

	code.WriteString("import (\n")
	code.WriteString("\t\"database/sql\"\n")
	code.WriteString("\t\"fmt\"\n\n")

	switch config.Type {
	case "mysql":
		code.WriteString("\t_ \"github.com/go-sql-driver/mysql\"\n")
	case "postgres":
		code.WriteString("\t_ \"github.com/lib/pq\"\n")
	}

	code.WriteString(")\n\n")
	code.WriteString("func NewDatabase() (*sql.DB, error) {\n")

	switch config.Type {
	case "mysql":
		code.WriteString(fmt.Sprintf("\tdsn := fmt.Sprintf(\"%%s:%%s@tcp(%%s:%d)/%%s?parseTime=true&charset=utf8mb4\",\n", config.Port))
		code.WriteString("\t\tos.Getenv(\"DB_USER\"),\n")
		code.WriteString("\t\tos.Getenv(\"DB_PASSWORD\"),\n")
		code.WriteString(fmt.Sprintf("\t\t\"%s\",\n", config.Host))
		code.WriteString(fmt.Sprintf("\t\t\"%s\",\n", config.Database))
		code.WriteString("\t)\n")
		code.WriteString("\treturn sql.Open(\"mysql\", dsn)\n")

	case "postgres":
		sslMode := config.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		code.WriteString(fmt.Sprintf("\tdsn := fmt.Sprintf(\"host=%%s port=%d user=%%s password=%%s dbname=%%s sslmode=%s\",\n", config.Port, sslMode))
		code.WriteString(fmt.Sprintf("\t\t\"%s\",\n", config.Host))
		code.WriteString("\t\tos.Getenv(\"DB_USER\"),\n")
		code.WriteString("\t\tos.Getenv(\"DB_PASSWORD\"),\n")
		code.WriteString(fmt.Sprintf("\t\t\"%s\",\n", config.Database))
		code.WriteString("\t)\n")
		code.WriteString("\treturn sql.Open(\"postgres\", dsn)\n")
	}

	code.WriteString("}\n")

	return code.String()
}

// GenerateMigrationSQL generates SQL for creating a table
func (g *DatabaseGenerator) GenerateMigrationSQL(dbType string, schema TableSchema) string {
	var sql strings.Builder

	sql.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", schema.Name))

	for i, col := range schema.Columns {
		if i > 0 {
			sql.WriteString(",\n")
		}
		sql.WriteString("\t")
		sql.WriteString(col.Name)
		sql.WriteString(" ")

		// Type mapping
		sqlType := g.mapTypeToSQL(dbType, col)
		sql.WriteString(sqlType)

		// Constraints
		if col.NotNull {
			sql.WriteString(" NOT NULL")
		}
		if col.Unique && !col.PrimaryKey {
			sql.WriteString(" UNIQUE")
		}
		if col.Default != "" {
			sql.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}
		if col.PrimaryKey {
			sql.WriteString(" PRIMARY KEY")
		}
	}

	sql.WriteString("\n);\n")

	return sql.String()
}

func (g *DatabaseGenerator) mapTypeToSQL(dbType string, col ColumnSchema) string {
	switch col.Type {
	case "int":
		if col.AutoIncrement {
			if dbType == "mysql" {
				return "INT AUTO_INCREMENT"
			}
			return "SERIAL"
		}
		return "INT"
	case "string":
		size := col.Size
		if size == 0 {
			size = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", size)
	case "text":
		return "TEXT"
	case "bool":
		if dbType == "mysql" {
			return "TINYINT(1)"
		}
		return "BOOLEAN"
	case "timestamp":
		if dbType == "mysql" {
			return "TIMESTAMP"
		}
		return "TIMESTAMP"
	default:
		return "VARCHAR(255)"
	}
}

// GenerateQueryCode generates Go code for a query
func (g *DatabaseGenerator) GenerateQueryCode(dbType string, query QuerySpec) string {
	var code strings.Builder

	switch query.Operation {
	case "select":
		code.WriteString("query := `SELECT ")
		code.WriteString(strings.Join(query.Columns, ", "))
		code.WriteString(" FROM ")
		code.WriteString(query.Table)

		if len(query.Where) > 0 {
			code.WriteString(" WHERE ")
			first := true
			for col, placeholder := range query.Where {
				if !first {
					code.WriteString(" AND ")
				}
				code.WriteString(col)
				code.WriteString(" = ")
				code.WriteString(placeholder)
				first = false
			}
		}

		if query.OrderBy != "" {
			code.WriteString(" ORDER BY ")
			code.WriteString(query.OrderBy)
		}

		if query.Limit > 0 {
			code.WriteString(fmt.Sprintf(" LIMIT %d", query.Limit))
		}

		code.WriteString("`\n")
	}

	return code.String()
}

// GenerateRepositoryCode generates a repository interface and implementation
func (g *DatabaseGenerator) GenerateRepositoryCode(model ModelSpec, dbType string) string {
	var code strings.Builder

	// Generate interface
	code.WriteString(fmt.Sprintf("type %sRepository interface {\n", model.Name))
	code.WriteString(fmt.Sprintf("\tFindByID(ctx context.Context, id int) (*%s, error)\n", model.Name))
	code.WriteString(fmt.Sprintf("\tFindAll(ctx context.Context) ([]*%s, error)\n", model.Name))
	code.WriteString(fmt.Sprintf("\tCreate(ctx context.Context, m *%s) error\n", model.Name))
	code.WriteString(fmt.Sprintf("\tUpdate(ctx context.Context, m *%s) error\n", model.Name))
	code.WriteString("\tDelete(ctx context.Context, id int) error\n")
	code.WriteString("}\n\n")

	// Generate implementation stub
	code.WriteString(fmt.Sprintf("type %sRepositoryImpl struct {\n", strings.ToLower(model.Name)))
	code.WriteString("\tdb *sql.DB\n")
	code.WriteString("}\n\n")

	code.WriteString(fmt.Sprintf("func New%sRepository(db *sql.DB) %sRepository {\n", model.Name, model.Name))
	code.WriteString(fmt.Sprintf("\treturn &%sRepositoryImpl{db: db}\n", strings.ToLower(model.Name)))
	code.WriteString("}\n")

	return code.String()
}

// GenerateAbstractionLayer generates a database abstraction interface
func (g *DatabaseGenerator) GenerateAbstractionLayer() string {
	var code strings.Builder

	code.WriteString("// Database provides a database abstraction layer\n")
	code.WriteString("type Database interface {\n")
	code.WriteString("\tQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)\n")
	code.WriteString("\tQueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row\n")
	code.WriteString("\tExec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)\n")
	code.WriteString("\tBegin(ctx context.Context) (*sql.Tx, error)\n")
	code.WriteString("\tClose() error\n")
	code.WriteString("}\n")

	return code.String()
}

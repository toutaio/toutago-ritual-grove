package generator

import (
	"testing"
)

func TestDatabaseGenerator_MySQL(t *testing.T) {
	gen := NewDatabaseGenerator()

	config := DBConnectionConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "secret",
	}

	code := gen.GenerateConnectionCode(config)

	// Verify MySQL-specific DSN format
	if !contains(code, "mysql") {
		t.Error("Expected MySQL driver reference in code")
	}
	if !contains(code, "parseTime=true") {
		t.Error("Expected parseTime parameter for MySQL")
	}
	if !contains(code, "3306") {
		t.Error("Expected MySQL port in connection string")
	}
}

func TestDatabaseGenerator_PostgreSQL(t *testing.T) {
	gen := NewDatabaseGenerator()

	config := DBConnectionConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		Password: "secret",
	}

	code := gen.GenerateConnectionCode(config)

	// Verify PostgreSQL-specific format
	if !contains(code, "postgres") {
		t.Error("Expected PostgreSQL driver reference in code")
	}
	if !contains(code, "sslmode") {
		t.Error("Expected sslmode parameter for PostgreSQL")
	}
	if !contains(code, "5432") {
		t.Error("Expected PostgreSQL port in connection string")
	}
}

func TestDatabaseGenerator_MigrationSQL_MySQL(t *testing.T) {
	gen := NewDatabaseGenerator()

	schema := TableSchema{
		Name: "users",
		Columns: []ColumnSchema{
			{Name: "id", Type: "int", PrimaryKey: true, AutoIncrement: true},
			{Name: "email", Type: "string", Size: 255, Unique: true, NotNull: true},
			{Name: "created_at", Type: "timestamp", NotNull: true},
		},
	}

	sql := gen.GenerateMigrationSQL("mysql", schema)

	// Verify MySQL-specific syntax
	if !contains(sql, "AUTO_INCREMENT") {
		t.Error("Expected AUTO_INCREMENT for MySQL")
	}
	if !contains(sql, "PRIMARY KEY") {
		t.Error("Expected PRIMARY KEY declaration")
	}
	if !contains(sql, "UNIQUE") {
		t.Error("Expected UNIQUE constraint")
	}
	if !contains(sql, "VARCHAR(255)") {
		t.Error("Expected VARCHAR type for string")
	}
}

func TestDatabaseGenerator_MigrationSQL_PostgreSQL(t *testing.T) {
	gen := NewDatabaseGenerator()

	schema := TableSchema{
		Name: "users",
		Columns: []ColumnSchema{
			{Name: "id", Type: "int", PrimaryKey: true, AutoIncrement: true},
			{Name: "email", Type: "string", Size: 255, Unique: true, NotNull: true},
			{Name: "created_at", Type: "timestamp", NotNull: true},
		},
	}

	sql := gen.GenerateMigrationSQL("postgres", schema)

	t.Logf("Generated SQL:\n%s", sql)

	// Verify PostgreSQL-specific syntax
	if !contains(sql, "SERIAL") && !contains(sql, "GENERATED") {
		t.Error("Expected SERIAL or GENERATED for auto-increment in PostgreSQL")
	}
	if !contains(sql, "PRIMARY KEY") {
		t.Error("Expected PRIMARY KEY declaration")
	}
}

func TestDatabaseGenerator_QueryCode_MySQL(t *testing.T) {
	gen := NewDatabaseGenerator()

	query := QuerySpec{
		Operation: "select",
		Table:     "users",
		Columns:   []string{"id", "email", "name"},
		Where:     map[string]string{"id": "?"},
	}

	code := gen.GenerateQueryCode("mysql", query)

	if !contains(code, "SELECT") {
		t.Error("Expected SELECT in generated query")
	}
	if !contains(code, "FROM users") {
		t.Error("Expected FROM users in query")
	}
	if !contains(code, "WHERE") {
		t.Error("Expected WHERE clause in query")
	}
	// MySQL uses ? for placeholders
	if !contains(code, "?") {
		t.Error("Expected ? placeholder for MySQL")
	}
}

func TestDatabaseGenerator_QueryCode_PostgreSQL(t *testing.T) {
	gen := NewDatabaseGenerator()

	query := QuerySpec{
		Operation: "select",
		Table:     "users",
		Columns:   []string{"id", "email", "name"},
		Where:     map[string]string{"id": "$1"},
	}

	code := gen.GenerateQueryCode("postgres", query)

	// PostgreSQL uses $1, $2, etc. for placeholders
	if !contains(code, "$1") && !contains(code, "?") {
		t.Error("Expected $1 or ? placeholder for PostgreSQL")
	}
}

func TestDatabaseGenerator_RepositoryCode(t *testing.T) {
	gen := NewDatabaseGenerator()

	model := ModelSpec{
		Name:  "User",
		Table: "users",
		Fields: []FieldSpec{
			{Name: "ID", Type: "int", DBColumn: "id"},
			{Name: "Email", Type: "string", DBColumn: "email"},
			{Name: "Name", Type: "string", DBColumn: "name"},
		},
	}

	code := gen.GenerateRepositoryCode(model, "mysql")

	if !contains(code, "type UserRepository") {
		t.Error("Expected UserRepository interface definition")
	}
	if !contains(code, "FindByID") {
		t.Error("Expected FindByID method")
	}
	if !contains(code, "Create") {
		t.Error("Expected Create method")
	}
	if !contains(code, "Update") {
		t.Error("Expected Update method")
	}
	if !contains(code, "Delete") {
		t.Error("Expected Delete method")
	}
}

func TestDatabaseGenerator_MultiDatabaseSupport(t *testing.T) {
	gen := NewDatabaseGenerator()

	// Should support generating code that works with both databases
	code := gen.GenerateAbstractionLayer()

	if !contains(code, "interface") {
		t.Error("Expected interface for database abstraction")
	}
	if !contains(code, "Database") {
		t.Error("Expected Database interface or type")
	}
}

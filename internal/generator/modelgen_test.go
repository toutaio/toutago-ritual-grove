package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModelGenerator_GenerateModel(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewModelGenerator()
	
	config := ModelConfig{
		Name:    "User",
		Package: "models",
		Fields: []Field{
			{Name: "ID", Type: "uint", Tags: `json:"id" db:"id"`},
			{Name: "Name", Type: "string", Tags: `json:"name" db:"name" validate:"required"`},
			{Name: "Email", Type: "string", Tags: `json:"email" db:"email" validate:"required,email"`},
		},
	}
	
	err := gen.GenerateModel(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateModel() error = %v", err)
	}
	
	modelPath := filepath.Join(tmpDir, "internal", "models", "user.go")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Error("Model file should be created")
	}
	
	content, _ := os.ReadFile(modelPath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "type User struct") {
		t.Error("Model should contain struct definition")
	}
	
	if !strings.Contains(contentStr, "ID uint") {
		t.Error("Model should contain ID field")
	}
	
	if !strings.Contains(contentStr, "Email string") {
		t.Error("Model should contain Email field")
	}
	
	if !strings.Contains(contentStr, `validate:"required,email"`) {
		t.Error("Model should contain validation tags")
	}
}

func TestModelGenerator_GenerateWithTimestamps(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewModelGenerator()
	
	config := ModelConfig{
		Name:       "Article",
		Timestamps: true,
		Fields: []Field{
			{Name: "Title", Type: "string", Tags: `json:"title"`},
		},
	}
	
	err := gen.GenerateModel(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateModel() error = %v", err)
	}
	
	modelPath := filepath.Join(tmpDir, "internal", "models", "article.go")
	content, _ := os.ReadFile(modelPath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "CreatedAt") {
		t.Error("Model should contain CreatedAt field")
	}
	
	if !strings.Contains(contentStr, "UpdatedAt") {
		t.Error("Model should contain UpdatedAt field")
	}
}

func TestModelGenerator_GenerateWithSoftDelete(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewModelGenerator()
	
	config := ModelConfig{
		Name:       "Post",
		SoftDelete: true,
		Fields: []Field{
			{Name: "Title", Type: "string", Tags: `json:"title"`},
		},
	}
	
	err := gen.GenerateModel(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateModel() error = %v", err)
	}
	
	modelPath := filepath.Join(tmpDir, "internal", "models", "post.go")
	content, _ := os.ReadFile(modelPath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "DeletedAt") {
		t.Error("Model should contain DeletedAt field")
	}
}

func TestModelGenerator_GenerateRepository(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewModelGenerator()
	
	config := ModelConfig{
		Name:               "Product",
		GenerateRepository: true,
		Fields: []Field{
			{Name: "Name", Type: "string"},
		},
	}
	
	err := gen.GenerateRepository(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateRepository() error = %v", err)
	}
	
	repoPath := filepath.Join(tmpDir, "internal", "repository", "product_repository.go")
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Error("Repository file should be created")
	}
	
	content, _ := os.ReadFile(repoPath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "ProductRepository") {
		t.Error("Repository should contain interface definition")
	}
	
	expectedMethods := []string{
		"Create",
		"GetByID",
		"List",
		"Update",
		"Delete",
	}
	
	for _, method := range expectedMethods {
		if !strings.Contains(contentStr, method) {
			t.Errorf("Repository should contain %s method", method)
		}
	}
}

func TestModelGenerator_GenerateWithRelationships(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewModelGenerator()
	
	config := ModelConfig{
		Name: "Comment",
		Fields: []Field{
			{Name: "Text", Type: "string"},
		},
		Relationships: []Relationship{
			{Name: "User", Type: "BelongsTo", Model: "User"},
			{Name: "Post", Type: "BelongsTo", Model: "Post"},
		},
	}
	
	err := gen.GenerateModel(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateModel() error = %v", err)
	}
	
	modelPath := filepath.Join(tmpDir, "internal", "models", "comment.go")
	content, _ := os.ReadFile(modelPath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "UserID") {
		t.Error("Model should contain foreign key UserID")
	}
	
	if !strings.Contains(contentStr, "PostID") {
		t.Error("Model should contain foreign key PostID")
	}
}

func TestModelGenerator_GenerateValidationMethods(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewModelGenerator()
	
	config := ModelConfig{
		Name:       "Account",
		Validation: true,
		Fields: []Field{
			{Name: "Email", Type: "string", Tags: `validate:"required,email"`},
		},
	}
	
	err := gen.GenerateModel(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateModel() error = %v", err)
	}
	
	modelPath := filepath.Join(tmpDir, "internal", "models", "account.go")
	content, _ := os.ReadFile(modelPath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "Validate()") {
		t.Error("Model should contain Validate method")
	}
}

func TestModelGenerator_GenerateMultipleModels(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewModelGenerator()
	
	configs := []ModelConfig{
		{Name: "Author", Fields: []Field{{Name: "Name", Type: "string"}}},
		{Name: "Book", Fields: []Field{{Name: "Title", Type: "string"}}},
		{Name: "Publisher", Fields: []Field{{Name: "Name", Type: "string"}}},
	}
	
	err := gen.GenerateMultiple(tmpDir, configs)
	if err != nil {
		t.Fatalf("GenerateMultiple() error = %v", err)
	}
	
	expectedFiles := []string{
		"internal/models/author.go",
		"internal/models/book.go",
		"internal/models/publisher.go",
	}
	
	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("File %s should be created", file)
		}
	}
}

func TestModelGenerator_GenerateWithJSONMethods(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewModelGenerator()
	
	config := ModelConfig{
		Name: "Config",
		Fields: []Field{
			{Name: "Settings", Type: "map[string]interface{}", Tags: `json:"settings"`},
		},
		JSONMethods: true,
	}
	
	err := gen.GenerateModel(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateModel() error = %v", err)
	}
	
	modelPath := filepath.Join(tmpDir, "internal", "models", "config.go")
	content, _ := os.ReadFile(modelPath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "MarshalJSON") || !strings.Contains(contentStr, "UnmarshalJSON") {
		t.Error("Model should contain JSON marshal/unmarshal methods")
	}
}

package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlerGenerator_GenerateHandler(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewHandlerGenerator()
	
	config := HandlerConfig{
		Name:       "User",
		Package:    "handlers",
		Operations: []string{"Create", "Get", "List", "Update", "Delete"},
		Model:      "User",
	}
	
	err := gen.GenerateHandler(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateHandler() error = %v", err)
	}
	
	handlerPath := filepath.Join(tmpDir, "internal", "handlers", "user_handler.go")
	if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
		t.Error("Handler file should be created")
	}
	
	content, _ := os.ReadFile(handlerPath)
	contentStr := string(content)
	
	expectedMethods := []string{
		"CreateUser",
		"GetUser",
		"ListUsers",
		"UpdateUser",
		"DeleteUser",
	}
	
	for _, method := range expectedMethods {
		if !strings.Contains(contentStr, method) {
			t.Errorf("Handler should contain method %s", method)
		}
	}
}

func TestHandlerGenerator_GenerateCRUDHandler(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewHandlerGenerator()
	
	config := HandlerConfig{
		Name:    "Product",
		Package: "handlers",
		Model:   "Product",
		CRUD:    true,
	}
	
	err := gen.GenerateHandler(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateHandler() error = %v", err)
	}
	
	handlerPath := filepath.Join(tmpDir, "internal", "handlers", "product_handler.go")
	content, _ := os.ReadFile(handlerPath)
	contentStr := string(content)
	
	expectedMethods := []string{
		"CreateProduct",
		"GetProduct",
		"ListProducts",
		"UpdateProduct",
		"DeleteProduct",
	}
	
	for _, method := range expectedMethods {
		if !strings.Contains(contentStr, method) {
			t.Errorf("CRUD handler should contain method %s", method)
		}
	}
}

func TestHandlerGenerator_GenerateHandlerTests(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewHandlerGenerator()
	
	config := HandlerConfig{
		Name:  "Order",
		Model: "Order",
		CRUD:  true,
	}
	
	err := gen.GenerateHandlerTests(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateHandlerTests() error = %v", err)
	}
	
	testPath := filepath.Join(tmpDir, "internal", "handlers", "order_handler_test.go")
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Error("Handler test file should be created")
	}
	
	content, _ := os.ReadFile(testPath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "TestOrderHandler") {
		t.Error("Test should contain handler test functions")
	}
	
	if !strings.Contains(contentStr, "httptest") {
		t.Error("Test should use httptest package")
	}
}

func TestHandlerGenerator_GenerateWithValidation(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewHandlerGenerator()
	
	config := HandlerConfig{
		Name:       "Article",
		Model:      "Article",
		CRUD:       true,
		Validation: true,
	}
	
	err := gen.GenerateHandler(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateHandler() error = %v", err)
	}
	
	handlerPath := filepath.Join(tmpDir, "internal", "handlers", "article_handler.go")
	content, _ := os.ReadFile(handlerPath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "Validate") {
		t.Error("Handler should include validation")
	}
}

func TestHandlerGenerator_GenerateWithRepository(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewHandlerGenerator()
	
	config := HandlerConfig{
		Name:       "Comment",
		Model:      "Comment",
		CRUD:       true,
		Repository: "CommentRepository",
	}
	
	err := gen.GenerateHandler(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateHandler() error = %v", err)
	}
	
	handlerPath := filepath.Join(tmpDir, "internal", "handlers", "comment_handler.go")
	content, _ := os.ReadFile(handlerPath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "CommentRepository") {
		t.Error("Handler should reference repository")
	}
}

func TestHandlerGenerator_GenerateMultipleHandlers(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewHandlerGenerator()
	
	configs := []HandlerConfig{
		{Name: "User", Model: "User", CRUD: true},
		{Name: "Post", Model: "Post", CRUD: true},
		{Name: "Tag", Model: "Tag", CRUD: true},
	}
	
	err := gen.GenerateMultiple(tmpDir, configs)
	if err != nil {
		t.Fatalf("GenerateMultiple() error = %v", err)
	}
	
	expectedFiles := []string{
		"internal/handlers/user_handler.go",
		"internal/handlers/post_handler.go",
		"internal/handlers/tag_handler.go",
	}
	
	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("File %s should be created", file)
		}
	}
}

func TestHandlerGenerator_GenerateWithCustomLogic(t *testing.T) {
	tmpDir := t.TempDir()
	
	gen := NewHandlerGenerator()
	
	config := HandlerConfig{
		Name:         "Payment",
		Model:        "Payment",
		Operations:   []string{"Process", "Refund"},
		CustomLogic:  true,
	}
	
	err := gen.GenerateHandler(tmpDir, config)
	if err != nil {
		t.Fatalf("GenerateHandler() error = %v", err)
	}
	
	handlerPath := filepath.Join(tmpDir, "internal", "handlers", "payment_handler.go")
	content, _ := os.ReadFile(handlerPath)
	contentStr := string(content)
	
	if !strings.Contains(contentStr, "ProcessPayment") {
		t.Error("Handler should contain ProcessPayment method")
	}
	
	if !strings.Contains(contentStr, "RefundPayment") {
		t.Error("Handler should contain RefundPayment method")
	}
	
	if !strings.Contains(contentStr, "// TODO: Implement") {
		t.Error("Custom logic handlers should have TODO comments")
	}
}

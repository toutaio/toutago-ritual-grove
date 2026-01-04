package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTestGenerator_GenerateUnitTests(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")
	
	// Create a handler file first
	handlersDir := filepath.Join(projectPath, "internal", "handlers")
	if err := os.MkdirAll(handlersDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	handlerContent := `package handlers

import (
	"net/http"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}
`
	handlerFile := filepath.Join(handlersDir, "hello.go")
	if err := os.WriteFile(handlerFile, []byte(handlerContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	generator := NewTestGenerator()
	err := generator.GenerateUnitTests(projectPath, "internal/handlers", "hello.go")
	if err != nil {
		t.Fatalf("GenerateUnitTests() error = %v", err)
	}
	
	// Verify test file was created
	testFile := filepath.Join(handlersDir, "hello_test.go")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Test file was not created")
	}
	
	// Read and verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	
	contentStr := string(content)
	if !contains(contentStr, "package handlers") {
		t.Error("Test file should have correct package")
	}
	if !contains(contentStr, "func TestHelloHandler") {
		t.Error("Test file should contain test function")
	}
	if !contains(contentStr, "testing.T") {
		t.Error("Test file should import testing package")
	}
}

func TestTestGenerator_GenerateHandlerTest(t *testing.T) {
	generator := NewTestGenerator()
	
	testSpec := HandlerTestSpec{
		PackageName:  "handlers",
		HandlerName:  "UserHandler",
		HTTPMethod:   "GET",
		Path:         "/users",
		ExpectedCode: 200,
		CheckBody:    true,
		BodyContains: "users",
	}
	
	content, err := generator.GenerateHandlerTest(testSpec)
	if err != nil {
		t.Fatalf("GenerateHandlerTest() error = %v", err)
	}
	
	if !contains(content, "package handlers") {
		t.Error("Should contain package declaration")
	}
	if !contains(content, "func TestUserHandler") {
		t.Error("Should contain test function")
	}
	if !contains(content, "GET") {
		t.Error("Should contain HTTP method")
	}
	if !contains(content, "/users") {
		t.Error("Should contain path")
	}
	if !contains(content, "200") {
		t.Error("Should contain expected status code")
	}
}

func TestTestGenerator_GenerateTableDrivenTest(t *testing.T) {
	generator := NewTestGenerator()
	
	testSpec := TableDrivenTestSpec{
		PackageName:  "models",
		FunctionName: "ValidateEmail",
		TestCases: []TestCase{
			{
				Name:     "valid email",
				Input:    map[string]interface{}{"email": "test@example.com"},
				Expected: true,
			},
			{
				Name:     "invalid email",
				Input:    map[string]interface{}{"email": "invalid"},
				Expected: false,
			},
		},
	}
	
	content, err := generator.GenerateTableDrivenTest(testSpec)
	if err != nil {
		t.Fatalf("GenerateTableDrivenTest() error = %v", err)
	}
	
	if !contains(content, "package models") {
		t.Error("Should contain package declaration")
	}
	if !contains(content, "func TestValidateEmail") {
		t.Error("Should contain test function")
	}
	if !contains(content, "tests := []struct") {
		t.Error("Should contain table structure")
	}
	if !contains(content, "valid email") {
		t.Error("Should contain test case names")
	}
	if !contains(content, "t.Run") {
		t.Error("Should use subtests")
	}
}

func TestTestGenerator_GenerateIntegrationTest(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")
	
	if err := os.MkdirAll(filepath.Join(projectPath, "test"), 0755); err != nil {
		t.Fatal(err)
	}
	
	generator := NewTestGenerator()
	
	integrationSpec := IntegrationTestSpec{
		Name:        "UserAPI",
		Description: "Test user API endpoints",
		SetupDB:     true,
		Endpoints: []EndpointTest{
			{
				Name:         "CreateUser",
				Method:       "POST",
				Path:         "/api/users",
				Body:         `{"name":"John"}`,
				ExpectedCode: 201,
			},
			{
				Name:         "GetUsers",
				Method:       "GET",
				Path:         "/api/users",
				ExpectedCode: 200,
			},
		},
	}
	
	err := generator.GenerateIntegrationTest(projectPath, integrationSpec)
	if err != nil {
		t.Fatalf("GenerateIntegrationTest() error = %v", err)
	}
	
	// Verify test file was created
	testFile := filepath.Join(projectPath, "test", "user_api_test.go")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Integration test file was not created")
	}
	
	// Read and verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	
	contentStr := string(content)
	if !contains(contentStr, "TestUserAPI") {
		t.Error("Should contain main test function")
	}
	if !contains(contentStr, "POST") && !contains(contentStr, "GET") {
		t.Error("Should contain HTTP methods")
	}
}

func TestTestGenerator_GenerateTestFixture(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")
	
	if err := os.MkdirAll(filepath.Join(projectPath, "test", "fixtures"), 0755); err != nil {
		t.Fatal(err)
	}
	
	generator := NewTestGenerator()
	
	fixture := TestFixture{
		Name: "users",
		Data: []map[string]interface{}{
			{
				"id":    1,
				"name":  "John Doe",
				"email": "john@example.com",
			},
			{
				"id":    2,
				"name":  "Jane Doe",
				"email": "jane@example.com",
			},
		},
	}
	
	err := generator.GenerateTestFixture(projectPath, fixture)
	if err != nil {
		t.Fatalf("GenerateTestFixture() error = %v", err)
	}
	
	// Verify fixture file was created
	fixtureFile := filepath.Join(projectPath, "test", "fixtures", "users.json")
	if _, err := os.Stat(fixtureFile); os.IsNotExist(err) {
		t.Error("Fixture file was not created")
	}
	
	// Read and verify content
	content, err := os.ReadFile(fixtureFile)
	if err != nil {
		t.Fatal(err)
	}
	
	contentStr := string(content)
	if !contains(contentStr, "John Doe") {
		t.Error("Fixture should contain test data")
	}
	if !contains(contentStr, "jane@example.com") {
		t.Error("Fixture should contain all entries")
	}
}

func TestTestGenerator_GenerateBenchmark(t *testing.T) {
	generator := NewTestGenerator()
	
	benchSpec := BenchmarkSpec{
		PackageName:  "services",
		FunctionName: "ProcessData",
		Setup:        "data := generateTestData()",
	}
	
	content, err := generator.GenerateBenchmark(benchSpec)
	if err != nil {
		t.Fatalf("GenerateBenchmark() error = %v", err)
	}
	
	if !contains(content, "func BenchmarkProcessData") {
		t.Error("Should contain benchmark function")
	}
	if !contains(content, "b.N") {
		t.Error("Should use b.N for iterations")
	}
	if !contains(content, "generateTestData") {
		t.Error("Should include setup code")
	}
}

func TestTestGenerator_GenerateMockInterface(t *testing.T) {
	generator := NewTestGenerator()
	
	mockSpec := MockSpec{
		PackageName:   "mocks",
		InterfaceName: "UserRepository",
		Methods: []MockMethod{
			{
				Name:       "GetByID",
				Parameters: []string{"id int"},
				Returns:    []string{"*User", "error"},
			},
			{
				Name:       "Create",
				Parameters: []string{"user *User"},
				Returns:    []string{"error"},
			},
		},
	}
	
	content, err := generator.GenerateMockInterface(mockSpec)
	if err != nil {
		t.Fatalf("GenerateMockInterface() error = %v", err)
	}
	
	if !contains(content, "type MockUserRepository") {
		t.Error("Should contain mock struct")
	}
	if !contains(content, "func (m *MockUserRepository) GetByID") {
		t.Error("Should contain mock method implementations")
	}
	if !contains(content, "func (m *MockUserRepository) Create") {
		t.Error("Should contain all methods")
	}
}

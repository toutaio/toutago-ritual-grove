package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TestGenerator generates test files
type TestGenerator struct{}

// NewTestGenerator creates a new test generator
func NewTestGenerator() *TestGenerator {
	return &TestGenerator{}
}

// HandlerTestSpec specifies a handler test
type HandlerTestSpec struct {
	PackageName  string
	HandlerName  string
	HTTPMethod   string
	Path         string
	ExpectedCode int
	CheckBody    bool
	BodyContains string
}

// TableDrivenTestSpec specifies a table-driven test
type TableDrivenTestSpec struct {
	PackageName  string
	FunctionName string
	TestCases    []TestCase
}

// TestCase represents a single test case
type TestCase struct {
	Name     string
	Input    map[string]interface{}
	Expected interface{}
	WantErr  bool
}

// IntegrationTestSpec specifies an integration test
type IntegrationTestSpec struct {
	Name        string
	Description string
	SetupDB     bool
	Endpoints   []EndpointTest
}

// EndpointTest specifies an endpoint to test
type EndpointTest struct {
	Name         string
	Method       string
	Path         string
	Body         string
	ExpectedCode int
}

// TestFixture represents test data
type TestFixture struct {
	Name string
	Data []map[string]interface{}
}

// BenchmarkSpec specifies a benchmark test
type BenchmarkSpec struct {
	PackageName  string
	FunctionName string
	Setup        string
}

// MockSpec specifies a mock interface
type MockSpec struct {
	PackageName   string
	InterfaceName string
	Methods       []MockMethod
}

// MockMethod represents a method in a mock interface
type MockMethod struct {
	Name       string
	Parameters []string
	Returns    []string
}

// GenerateUnitTests generates unit tests for a source file
func (g *TestGenerator) GenerateUnitTests(projectPath, packagePath, sourceFile string) error {
	// Parse source file name
	baseName := strings.TrimSuffix(sourceFile, ".go")
	testFileName := baseName + "_test.go"

	// Determine package name from path
	packageName := filepath.Base(packagePath)

	// Convert to title case for function name (e.g., hello -> Hello)
	functionName := strings.Title(baseName)
	// Handle snake_case or kebab-case
	functionName = strings.ReplaceAll(functionName, "_", "")
	functionName = strings.ReplaceAll(functionName, "-", "")

	// Create test file content
	content := fmt.Sprintf(`package %s

import (
	"testing"
)

// TODO: Add tests for %s
func Test%sHandler(t *testing.T) {
	tests := []struct {
		name string
		want interface{}
	}{
		{
			name: "test case 1",
			want: nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Implement test
		})
	}
}
`, packageName, sourceFile, functionName)

	// Write test file
	testPath := filepath.Join(projectPath, packagePath, testFileName)
	return os.WriteFile(testPath, []byte(content), 0644)
}

// GenerateHandlerTest generates a test for an HTTP handler
func (g *TestGenerator) GenerateHandlerTest(spec HandlerTestSpec) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s\n\n", spec.PackageName))
	sb.WriteString("import (\n")
	sb.WriteString("\t\"net/http\"\n")
	sb.WriteString("\t\"net/http/httptest\"\n")
	sb.WriteString("\t\"testing\"\n")
	if spec.CheckBody {
		sb.WriteString("\t\"strings\"\n")
	}
	sb.WriteString(")\n\n")

	sb.WriteString(fmt.Sprintf("func Test%s(t *testing.T) {\n", spec.HandlerName))
	sb.WriteString("\treq := httptest.NewRequest(\"")
	sb.WriteString(spec.HTTPMethod)
	sb.WriteString("\", \"")
	sb.WriteString(spec.Path)
	sb.WriteString("\", nil)\n")
	sb.WriteString("\tw := httptest.NewRecorder()\n\n")

	sb.WriteString(fmt.Sprintf("\t%s(w, req)\n\n", spec.HandlerName))

	sb.WriteString("\tresp := w.Result()\n")
	sb.WriteString(fmt.Sprintf("\tif resp.StatusCode != %d {\n", spec.ExpectedCode))
	sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Expected status %d, got %%d\", resp.StatusCode)\n", spec.ExpectedCode))
	sb.WriteString("\t}\n")

	if spec.CheckBody && spec.BodyContains != "" {
		sb.WriteString("\n\tbody := w.Body.String()\n")
		sb.WriteString(fmt.Sprintf("\tif !strings.Contains(body, \"%s\") {\n", spec.BodyContains))
		sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Expected body to contain '%s', got: %%s\", body)\n", spec.BodyContains))
		sb.WriteString("\t}\n")
	}

	sb.WriteString("}\n")

	return sb.String(), nil
}

// GenerateTableDrivenTest generates a table-driven test
func (g *TestGenerator) GenerateTableDrivenTest(spec TableDrivenTestSpec) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s\n\n", spec.PackageName))
	sb.WriteString("import (\n")
	sb.WriteString("\t\"testing\"\n")
	sb.WriteString(")\n\n")

	sb.WriteString(fmt.Sprintf("func Test%s(t *testing.T) {\n", spec.FunctionName))
	sb.WriteString("\ttests := []struct {\n")
	sb.WriteString("\t\tname string\n")
	sb.WriteString("\t\tinput interface{}\n")
	sb.WriteString("\t\twant interface{}\n")
	sb.WriteString("\t\twantErr bool\n")
	sb.WriteString("\t}{\n")

	for _, tc := range spec.TestCases {
		sb.WriteString("\t\t{\n")
		sb.WriteString(fmt.Sprintf("\t\t\tname: \"%s\",\n", tc.Name))
		sb.WriteString(fmt.Sprintf("\t\t\tinput: %v,\n", tc.Input))
		sb.WriteString(fmt.Sprintf("\t\t\twant: %v,\n", tc.Expected))
		sb.WriteString(fmt.Sprintf("\t\t\twantErr: %v,\n", tc.WantErr))
		sb.WriteString("\t\t},\n")
	}

	sb.WriteString("\t}\n\n")
	sb.WriteString("\tfor _, tt := range tests {\n")
	sb.WriteString("\t\tt.Run(tt.name, func(t *testing.T) {\n")
	sb.WriteString(fmt.Sprintf("\t\t\t// got, err := %s(tt.input)\n", spec.FunctionName))
	sb.WriteString("\t\t\t// TODO: Implement test logic\n")
	sb.WriteString("\t\t})\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String(), nil
}

// GenerateIntegrationTest generates an integration test
func (g *TestGenerator) GenerateIntegrationTest(projectPath string, spec IntegrationTestSpec) error {
	var sb strings.Builder

	sb.WriteString("package test\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"net/http\"\n")
	sb.WriteString("\t\"net/http/httptest\"\n")
	sb.WriteString("\t\"testing\"\n")
	if spec.SetupDB {
		sb.WriteString("\t\"database/sql\"\n")
	}
	sb.WriteString(")\n\n")

	sb.WriteString(fmt.Sprintf("// Test%s %s\n", spec.Name, spec.Description))
	sb.WriteString(fmt.Sprintf("func Test%s(t *testing.T) {\n", spec.Name))

	if spec.SetupDB {
		sb.WriteString("\t// Setup test database\n")
		sb.WriteString("\t// db := setupTestDB(t)\n")
		sb.WriteString("\t// defer db.Close()\n\n")
	}

	for _, endpoint := range spec.Endpoints {
		sb.WriteString(fmt.Sprintf("\tt.Run(\"%s\", func(t *testing.T) {\n", endpoint.Name))
		sb.WriteString(fmt.Sprintf("\t\treq := httptest.NewRequest(\"%s\", \"%s\", nil)\n",
			endpoint.Method, endpoint.Path))
		sb.WriteString("\t\tw := httptest.NewRecorder()\n\n")
		sb.WriteString("\t\t// TODO: Call handler\n\n")
		sb.WriteString("\t\tresp := w.Result()\n")
		sb.WriteString(fmt.Sprintf("\t\tif resp.StatusCode != %d {\n", endpoint.ExpectedCode))
		sb.WriteString(fmt.Sprintf("\t\t\tt.Errorf(\"Expected status %d, got %%d\", resp.StatusCode)\n",
			endpoint.ExpectedCode))
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t})\n\n")
	}

	sb.WriteString("}\n")

	// Write to file
	fileName := strings.ToLower(spec.Name) + "_test.go"
	// Convert CamelCase to snake_case for filename
	fileName = camelToSnake(spec.Name) + "_test.go"
	testDir := filepath.Join(projectPath, "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}
	testPath := filepath.Join(testDir, fileName)
	return os.WriteFile(testPath, []byte(sb.String()), 0644)
}

// camelToSnake converts CamelCase to snake_case
func camelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Check if previous char was lowercase or next char is lowercase
			if i > 0 {
				prevChar := rune(s[i-1])
				if prevChar >= 'a' && prevChar <= 'z' {
					result = append(result, '_')
				} else if i+1 < len(s) {
					nextChar := rune(s[i+1])
					if nextChar >= 'a' && nextChar <= 'z' {
						result = append(result, '_')
					}
				}
			}
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// GenerateTestFixture generates test fixture data
func (g *TestGenerator) GenerateTestFixture(projectPath string, fixture TestFixture) error {
	// Marshal data to JSON
	data, err := json.MarshalIndent(fixture.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal fixture data: %w", err)
	}

	// Write to file
	fixturePath := filepath.Join(projectPath, "test", "fixtures", fixture.Name+".json")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0755); err != nil {
		return fmt.Errorf("failed to create fixtures directory: %w", err)
	}

	return os.WriteFile(fixturePath, data, 0644)
}

// GenerateBenchmark generates a benchmark test
func (g *TestGenerator) GenerateBenchmark(spec BenchmarkSpec) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s\n\n", spec.PackageName))
	sb.WriteString("import (\n")
	sb.WriteString("\t\"testing\"\n")
	sb.WriteString(")\n\n")

	sb.WriteString(fmt.Sprintf("func Benchmark%s(b *testing.B) {\n", spec.FunctionName))
	if spec.Setup != "" {
		sb.WriteString(fmt.Sprintf("\t%s\n\n", spec.Setup))
	}
	sb.WriteString("\tb.ResetTimer()\n")
	sb.WriteString("\tfor i := 0; i < b.N; i++ {\n")
	sb.WriteString(fmt.Sprintf("\t\t// %s()\n", spec.FunctionName))
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String(), nil
}

// GenerateMockInterface generates a mock implementation of an interface
func (g *TestGenerator) GenerateMockInterface(spec MockSpec) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s\n\n", spec.PackageName))

	// Generate mock struct
	mockName := "Mock" + spec.InterfaceName
	sb.WriteString(fmt.Sprintf("// %s is a mock implementation of %s\n", mockName, spec.InterfaceName))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", mockName))
	sb.WriteString("}\n\n")

	// Generate method implementations
	for _, method := range spec.Methods {
		sb.WriteString(fmt.Sprintf("func (m *%s) %s(", mockName, method.Name))
		sb.WriteString(strings.Join(method.Parameters, ", "))
		sb.WriteString(") (")
		sb.WriteString(strings.Join(method.Returns, ", "))
		sb.WriteString(") {\n")
		sb.WriteString("\t// TODO: Implement mock behavior\n")

		// Return zero values
		if len(method.Returns) > 0 {
			sb.WriteString("\treturn ")
			for i, ret := range method.Returns {
				if i > 0 {
					sb.WriteString(", ")
				}
				if strings.Contains(ret, "error") {
					sb.WriteString("nil")
				} else if strings.HasPrefix(ret, "*") {
					sb.WriteString("nil")
				} else {
					sb.WriteString("\"\"") // Default string, adjust as needed
				}
			}
			sb.WriteString("\n")
		}

		sb.WriteString("}\n\n")
	}

	return sb.String(), nil
}

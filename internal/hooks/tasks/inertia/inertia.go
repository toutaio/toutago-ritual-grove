package inertia

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

// SetupInertiaMiddlewareTask adds Inertia middleware to main.go.
type SetupInertiaMiddlewareTask struct {
	ProjectDir string
}

func (t *SetupInertiaMiddlewareTask) Name() string {
	return "setup-inertia-middleware"
}

func (t *SetupInertiaMiddlewareTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	projectDir := t.ProjectDir
	if projectDir == "" {
		projectDir = taskCtx.WorkingDir()
	}

	mainFile := filepath.Join(projectDir, "main.go")

	// Read main.go
	content, err := os.ReadFile(mainFile)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	contentStr := string(content)

	// Check if already has inertia import
	if !strings.Contains(contentStr, "github.com/toutaio/toutago-inertia") {
		// Add import after cosan import
		contentStr = strings.Replace(contentStr,
			`"github.com/toutaio/toutago/cosan"`,
			`"github.com/toutaio/toutago/cosan"
	"github.com/toutaio/toutago-inertia"`,
			1)
	}

	// Check if already has middleware setup
	if !strings.Contains(contentStr, "inertia.NewMiddleware") {
		// Add middleware before router.Run()
		middlewareCode := `
	// Setup Inertia middleware
	router.Use(inertia.NewMiddleware(inertia.Config{
		URL:     "http://localhost:8080",
		Version: "1",
	}))

	`
		contentStr = strings.Replace(contentStr,
			"router.Run(",
			middlewareCode+"router.Run(",
			1)
	}

	return os.WriteFile(mainFile, []byte(contentStr), 0644)
}

func (t *SetupInertiaMiddlewareTask) Validate() error {
	return nil
}

// AddInertiaHandlersTask generates Inertia-compatible handlers.
type AddInertiaHandlersTask struct {
	ProjectDir string
	Resource   string
}

func (t *AddInertiaHandlersTask) Name() string {
	return "add-inertia-handlers"
}

func (t *AddInertiaHandlersTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	projectDir := t.ProjectDir
	if projectDir == "" {
		projectDir = taskCtx.WorkingDir()
	}

	resourceName := t.Resource
	if val, ok := taskCtx.Get("resource"); ok {
		if str, ok := val.(string); ok {
			resourceName = str
		}
	}

	handlersDir := filepath.Join(projectDir, "internal", "handlers")
	if err := os.MkdirAll(handlersDir, 0755); err != nil {
		return fmt.Errorf("failed to create handlers directory: %w", err)
	}

	handlerFile := filepath.Join(handlersDir, resourceName+"_handler.go")

	template := fmt.Sprintf(`package handlers

import (
	"github.com/toutaio/toutago/cosan"
	"github.com/toutaio/toutago-inertia"
)

// %[1]sIndex handles the index page.
func %[1]sIndex(ctx *cosan.Context) error {
	// TODO: Fetch %[2]s from database
	%[2]s := []map[string]interface{}{}
	
	return ctx.Inertia().Render("%[1]s/Index", inertia.Props{
		"%[2]s": %[2]s,
	})
}

// %[1]sShow handles the show page.
func %[1]sShow(ctx *cosan.Context) error {
	id := ctx.Param("id")
	
	// TODO: Fetch %[2]s from database by id
	%[2]s := map[string]interface{}{
		"id": id,
	}
	
	return ctx.Inertia().Render("%[1]s/Show", inertia.Props{
		"%[2]s": %[2]s,
	})
}

// %[1]sCreate handles creating a new %[2]s.
func %[1]sCreate(ctx *cosan.Context) error {
	// TODO: Validate and create %[2]s
	
	return ctx.Inertia().Redirect("/%[2]s")
}

// %[1]sUpdate handles updating a %[2]s.
func %[1]sUpdate(ctx *cosan.Context) error {
	id := ctx.Param("id")
	
	// TODO: Validate and update %[2]s
	_ = id
	
	return ctx.Inertia().Redirect("/%[2]s/" + id)
}

// %[1]sDelete handles deleting a %[2]s.
func %[1]sDelete(ctx *cosan.Context) error {
	id := ctx.Param("id")
	
	// TODO: Delete %[2]s
	_ = id
	
	return ctx.Inertia().Redirect("/%[2]s")
}
`,
		capitalize(resourceName), resourceName,
	)

	return os.WriteFile(handlerFile, []byte(template), 0644)
}

func (t *AddInertiaHandlersTask) Validate() error {
	if t.Resource == "" {
		return errors.New("resource is required")
	}
	return nil
}

// AddSharedDataTask adds shared data configuration.
type AddSharedDataTask struct {
	ProjectDir string
	SharedData []string
}

func (t *AddSharedDataTask) Name() string {
	return "add-shared-data"
}

func (t *AddSharedDataTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	projectDir := t.ProjectDir
	if projectDir == "" {
		projectDir = taskCtx.WorkingDir()
	}

	sharedData := t.SharedData
	if val, ok := taskCtx.Get("shared_data"); ok {
		if arr, ok := val.([]string); ok {
			sharedData = arr
		}
	}

	configDir := filepath.Join(projectDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "inertia.go")

	funcs := []string{}
	for _, data := range sharedData {
		funcs = append(funcs, generateSharedDataFunc(data))
	}

	template := fmt.Sprintf(`package config

import (
	"github.com/toutaio/toutago/cosan"
	"github.com/toutaio/toutago-inertia"
)

// SharedData returns the shared data functions.
func SharedData() map[string]inertia.SharedDataFunc {
	return map[string]inertia.SharedDataFunc{
%s	}
}

%s
`, strings.Join(funcs, "\n"), generateSharedDataHelpers(sharedData))

	return os.WriteFile(configFile, []byte(template), 0644)
}

func (t *AddSharedDataTask) Validate() error {
	return nil
}

// GenerateTypeScriptTypesTask generates TypeScript types from Go structs.
type GenerateTypeScriptTypesTask struct {
	ProjectDir string
	ModelsDir  string
	OutputDir  string
}

func (t *GenerateTypeScriptTypesTask) Name() string {
	return "generate-typescript-types"
}

func (t *GenerateTypeScriptTypesTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	projectDir := t.ProjectDir
	if projectDir == "" {
		projectDir = taskCtx.WorkingDir()
	}

	modelsDir := t.ModelsDir
	if val, ok := taskCtx.Get("models_dir"); ok {
		if str, ok := val.(string); ok {
			modelsDir = str
		}
	}

	outputDir := t.OutputDir
	if val, ok := taskCtx.Get("output_dir"); ok {
		if str, ok := val.(string); ok {
			outputDir = str
		}
	}

	outputFile := filepath.Join(outputDir, "models.d.ts")

	// Parse Go files in models directory
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, modelsDir, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse models: %w", err)
	}

	var types []string
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				typeSpec, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					return true
				}

				tsType := convertStructToTS(typeSpec.Name.Name, structType)
				types = append(types, tsType)
				return true
			})
		}
	}

	content := "// Auto-generated TypeScript types from Go structs\n\n" + strings.Join(types, "\n\n")
	return os.WriteFile(outputFile, []byte(content), 0644)
}

func (t *GenerateTypeScriptTypesTask) Validate() error {
	if t.ModelsDir == "" {
		return errors.New("models_dir is required")
	}
	if t.OutputDir == "" {
		return errors.New("output_dir is required")
	}
	return nil
}

// UpdateRoutesForInertiaTask updates route definitions for Inertia.
type UpdateRoutesForInertiaTask struct {
	ProjectDir string
	Resource   string
}

func (t *UpdateRoutesForInertiaTask) Name() string {
	return "update-routes-for-inertia"
}

func (t *UpdateRoutesForInertiaTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	projectDir := t.ProjectDir
	if projectDir == "" {
		projectDir = taskCtx.WorkingDir()
	}

	resourceName := t.Resource
	if val, ok := taskCtx.Get("resource"); ok {
		if str, ok := val.(string); ok {
			resourceName = str
		}
	}

	routesFile := filepath.Join(projectDir, "internal", "routes", "routes.go")

	content, err := os.ReadFile(routesFile)
	if err != nil {
		return fmt.Errorf("failed to read routes file: %w", err)
	}

	// Add routes before the closing brace of Setup function
	routes := fmt.Sprintf(`
	// %s routes
	router.GET("/%s", handlers.%sIndex)
	router.GET("/%s/:id", handlers.%sShow)
	router.POST("/%s", handlers.%sCreate)
	router.PUT("/%s/:id", handlers.%sUpdate)
	router.DELETE("/%s/:id", handlers.%sDelete)
`,
		capitalize(resourceName),
		resourceName, capitalize(resourceName),
		resourceName, capitalize(resourceName),
		resourceName, capitalize(resourceName),
		resourceName, capitalize(resourceName),
		resourceName, capitalize(resourceName),
	)

	modified := strings.Replace(string(content), "// Existing routes", "// Existing routes"+routes, 1)
	return os.WriteFile(routesFile, []byte(modified), 0644)
}

func (t *UpdateRoutesForInertiaTask) Validate() error {
	if t.Resource == "" {
		return errors.New("resource is required")
	}
	return nil
}

// Helper functions

func generateSharedDataFunc(name string) string {
	return fmt.Sprintf(`		"%s": Get%s,`, name, capitalize(name))
}

func generateSharedDataHelpers(sharedData []string) string {
	helpers := []string{}
	for _, data := range sharedData {
		helper := fmt.Sprintf(`// Get%s returns the %s data.
func Get%s(ctx *cosan.Context) interface{} {
	// TODO: Implement %s retrieval
	return nil
}`, capitalize(data), data, capitalize(data), data)
		helpers = append(helpers, helper)
	}
	return strings.Join(helpers, "\n\n")
}

func convertStructToTS(name string, structType *ast.StructType) string {
	var fields []string
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}

		fieldName := field.Names[0].Name
		if !ast.IsExported(fieldName) {
			continue
		}

		// Get JSON tag
		jsonName := fieldName
		if field.Tag != nil {
			tag := strings.Trim(field.Tag.Value, "`")
			if strings.Contains(tag, "json:") {
				parts := strings.Split(tag, "json:\"")
				if len(parts) > 1 {
					jsonName = strings.Split(parts[1], "\"")[0]
				}
			}
		}

		tsType := goTypeToTS(field.Type)
		fields = append(fields, fmt.Sprintf("  %s: %s;", jsonName, tsType))
	}

	return fmt.Sprintf("export interface %s {\n%s\n}", name, strings.Join(fields, "\n"))
}

func goTypeToTS(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "string"
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64":
			return "number"
		case "bool":
			return "boolean"
		default:
			return "any"
		}
	case *ast.ArrayType:
		return goTypeToTS(t.Elt) + "[]"
	case *ast.StarExpr:
		return goTypeToTS(t.X) + " | null"
	default:
		return "any"
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Register all Inertia tasks.
func init() {
	tasks.Register("setup-inertia-middleware", func(config map[string]interface{}) (tasks.Task, error) {
		projectDir, _ := config["project_dir"].(string)
		return &SetupInertiaMiddlewareTask{ProjectDir: projectDir}, nil
	})

	tasks.Register("add-inertia-handlers", func(config map[string]interface{}) (tasks.Task, error) {
		projectDir, _ := config["project_dir"].(string)
		resource, _ := config["resource"].(string)
		return &AddInertiaHandlersTask{ProjectDir: projectDir, Resource: resource}, nil
	})

	tasks.Register("add-shared-data", func(config map[string]interface{}) (tasks.Task, error) {
		projectDir, _ := config["project_dir"].(string)
		sharedData, _ := config["shared_data"].([]string)
		return &AddSharedDataTask{ProjectDir: projectDir, SharedData: sharedData}, nil
	})

	tasks.Register("generate-typescript-types", func(config map[string]interface{}) (tasks.Task, error) {
		projectDir, _ := config["project_dir"].(string)
		modelsDir, _ := config["models_dir"].(string)
		outputDir, _ := config["output_dir"].(string)
		return &GenerateTypeScriptTypesTask{
			ProjectDir: projectDir,
			ModelsDir:  modelsDir,
			OutputDir:  outputDir,
		}, nil
	})

	tasks.Register("update-routes-for-inertia", func(config map[string]interface{}) (tasks.Task, error) {
		projectDir, _ := config["project_dir"].(string)
		resource, _ := config["resource"].(string)
		return &UpdateRoutesForInertiaTask{ProjectDir: projectDir, Resource: resource}, nil
	})
}


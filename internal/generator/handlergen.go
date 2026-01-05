package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HandlerGenerator generates HTTP handlers
type HandlerGenerator struct{}

// NewHandlerGenerator creates a new handler generator
func NewHandlerGenerator() *HandlerGenerator {
	return &HandlerGenerator{}
}

// HandlerConfig configures handler generation
type HandlerConfig struct {
	Name        string
	Package     string
	Model       string
	Operations  []string
	CRUD        bool
	Validation  bool
	Repository  string
	CustomLogic bool
}

// GenerateHandler generates a handler file
func (g *HandlerGenerator) GenerateHandler(targetPath string, config HandlerConfig) error {
	if config.Package == "" {
		config.Package = "handlers"
	}

	if config.Repository == "" {
		config.Repository = config.Model + "Repository"
	}

	var operations []string
	if config.CRUD {
		operations = []string{"Create", "Get", "List", "Update", "Delete"}
	} else {
		operations = config.Operations
	}

	handlerName := strings.ToLower(config.Name)
	fileName := handlerName + "_handler.go"

	content := g.generateHandlerContent(config, operations)

	handlerDir := filepath.Join(targetPath, "internal", config.Package)
	if err := os.MkdirAll(handlerDir, 0750); err != nil {
		return err
	}

	handlerPath := filepath.Join(handlerDir, fileName)
	return os.WriteFile(handlerPath, []byte(content), 0600)
}

func (g *HandlerGenerator) generateHandlerContent(config HandlerConfig, operations []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`package %s

import (
	"encoding/json"
	"net/http"
	
	"github.com/gorilla/mux"
)

// %sHandler handles %s-related HTTP requests
type %sHandler struct {
	repo %s
}

// New%sHandler creates a new %s handler
func New%sHandler(repo %s) *%sHandler {
	return &%sHandler{
		repo: repo,
	}
}

`,
		config.Package,
		config.Name,
		strings.ToLower(config.Name),
		config.Name,
		config.Repository,
		config.Name,
		config.Name,
		config.Name,
		config.Repository,
		config.Name,
		config.Name,
	))

	for _, op := range operations {
		sb.WriteString(g.generateMethod(config, op))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *HandlerGenerator) generateMethod(config HandlerConfig, operation string) string {
	methodName := operation + config.Name

	// Pluralize List method names
	if operation == "List" {
		methodName = operation + config.Name + "s"
	}

	switch operation {
	case "Create":
		return g.generateCreateMethod(config, methodName)
	case "Get":
		return g.generateGetMethod(config, methodName)
	case "List":
		return g.generateListMethod(config, methodName)
	case "Update":
		return g.generateUpdateMethod(config, methodName)
	case "Delete":
		return g.generateDeleteMethod(config, methodName)
	default:
		return g.generateCustomMethod(config, methodName, operation)
	}
}

func (g *HandlerGenerator) generateCreateMethod(config HandlerConfig, methodName string) string {
	validation := ""
	if config.Validation {
		validation = `
	// Validate input
	if err := item.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
`
	}

	return fmt.Sprintf(`// %s creates a new %s
func (h *%sHandler) %s(w http.ResponseWriter, r *http.Request) {
	var item %s
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
%s
	// Create in repository
	created, err := h.repo.Create(r.Context(), &item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}
`,
		methodName,
		strings.ToLower(config.Name),
		config.Name,
		methodName,
		config.Model,
		validation,
	)
}

func (g *HandlerGenerator) generateGetMethod(config HandlerConfig, methodName string) string {
	return fmt.Sprintf(`// %s retrieves a %s by ID
func (h *%sHandler) %s(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	item, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}
`,
		methodName,
		strings.ToLower(config.Name),
		config.Name,
		methodName,
	)
}

func (g *HandlerGenerator) generateListMethod(config HandlerConfig, methodName string) string {
	return fmt.Sprintf(`// %s retrieves all %ss
func (h *%sHandler) %s(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
`,
		methodName,
		strings.ToLower(config.Name),
		config.Name,
		methodName,
	)
}

func (g *HandlerGenerator) generateUpdateMethod(config HandlerConfig, methodName string) string {
	validation := ""
	if config.Validation {
		validation = `
	// Validate input
	if err := item.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
`
	}

	return fmt.Sprintf(`// %s updates a %s
func (h *%sHandler) %s(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	var item %s
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
%s
	updated, err := h.repo.Update(r.Context(), id, &item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
`,
		methodName,
		strings.ToLower(config.Name),
		config.Name,
		methodName,
		config.Model,
		validation,
	)
}

func (g *HandlerGenerator) generateDeleteMethod(config HandlerConfig, methodName string) string {
	return fmt.Sprintf(`// %s deletes a %s
func (h *%sHandler) %s(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if err := h.repo.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
`,
		methodName,
		strings.ToLower(config.Name),
		config.Name,
		methodName,
	)
}

func (g *HandlerGenerator) generateCustomMethod(config HandlerConfig, methodName, operation string) string {
	todoComment := ""
	if config.CustomLogic {
		todoComment = "\n\t// TODO: Implement custom logic for " + operation
	}

	return fmt.Sprintf(`// %s handles %s operation
func (h *%sHandler) %s(w http.ResponseWriter, r *http.Request) {%s
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
`,
		methodName,
		operation,
		config.Name,
		methodName,
		todoComment,
	)
}

// GenerateHandlerTests generates test file for handler
func (g *HandlerGenerator) GenerateHandlerTests(targetPath string, config HandlerConfig) error {
	if config.Package == "" {
		config.Package = "handlers"
	}

	handlerName := strings.ToLower(config.Name)
	fileName := handlerName + "_handler_test.go"

	content := fmt.Sprintf(`package %s

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	
	"github.com/gorilla/mux"
)

func Test%sHandler_Create(t *testing.T) {
	// Setup
	handler := New%sHandler(nil)
	
	// Test case
	payload := map[string]interface{}{
		"name": "test",
	}
	
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/%ss", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	
	handler.Create%s(w, req)
	
	// Assertions would go here
}

func Test%sHandler_Get(t *testing.T) {
	handler := New%sHandler(nil)
	
	req := httptest.NewRequest(http.MethodGet, "/%ss/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()
	
	handler.Get%s(w, req)
	
	// Assertions would go here
}

func Test%sHandler_List(t *testing.T) {
	handler := New%sHandler(nil)
	
	req := httptest.NewRequest(http.MethodGet, "/%ss", nil)
	w := httptest.NewRecorder()
	
	handler.List%ss(w, req)
	
	// Assertions would go here
}
`,
		config.Package,
		config.Name,
		config.Name,
		strings.ToLower(config.Name),
		config.Name,
		config.Name,
		config.Name,
		strings.ToLower(config.Name),
		config.Name,
		config.Name,
		config.Name,
		strings.ToLower(config.Name),
		config.Name,
	)

	handlerDir := filepath.Join(targetPath, "internal", config.Package)
	if err := os.MkdirAll(handlerDir, 0750); err != nil {
		return err
	}

	testPath := filepath.Join(handlerDir, fileName)
	return os.WriteFile(testPath, []byte(content), 0600)
}

// GenerateMultiple generates multiple handlers
func (g *HandlerGenerator) GenerateMultiple(targetPath string, configs []HandlerConfig) error {
	for _, config := range configs {
		if err := g.GenerateHandler(targetPath, config); err != nil {
			return fmt.Errorf("failed to generate %s handler: %w", config.Name, err)
		}
	}
	return nil
}

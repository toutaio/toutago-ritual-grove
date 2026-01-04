package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModelGenerator generates data models
type ModelGenerator struct{}

// NewModelGenerator creates a new model generator
func NewModelGenerator() *ModelGenerator {
	return &ModelGenerator{}
}

// Field represents a model field
type Field struct {
	Name string
	Type string
	Tags string
}

// Relationship represents a model relationship
type Relationship struct {
	Name  string
	Type  string // BelongsTo, HasMany, ManyToMany
	Model string
}

// ModelConfig configures model generation
type ModelConfig struct {
	Name               string
	Package            string
	Fields             []Field
	Relationships      []Relationship
	Timestamps         bool
	SoftDelete         bool
	Validation         bool
	JSONMethods        bool
	GenerateRepository bool
}

// GenerateModel generates a model file
func (g *ModelGenerator) GenerateModel(targetPath string, config ModelConfig) error {
	if config.Package == "" {
		config.Package = "models"
	}
	
	modelName := strings.ToLower(config.Name)
	fileName := modelName + ".go"
	
	content := g.generateModelContent(config)
	
	modelDir := filepath.Join(targetPath, "internal", config.Package)
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return err
	}
	
	modelPath := filepath.Join(modelDir, fileName)
	return os.WriteFile(modelPath, []byte(content), 0644)
}

func (g *ModelGenerator) generateModelContent(config ModelConfig) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf(`package %s

import (
	"time"
)

`, config.Package))
	
	// Generate struct
	sb.WriteString(fmt.Sprintf("// %s represents a %s entity\n", config.Name, strings.ToLower(config.Name)))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", config.Name))
	
	// Add ID field
	sb.WriteString("\tID uint `json:\"id\" db:\"id\"`\n")
	
	// Add regular fields
	for _, field := range config.Fields {
		sb.WriteString(fmt.Sprintf("\t%s %s", field.Name, field.Type))
		if field.Tags != "" {
			sb.WriteString(fmt.Sprintf(" `%s`", field.Tags))
		}
		sb.WriteString("\n")
	}
	
	// Add relationship foreign keys
	for _, rel := range config.Relationships {
		if rel.Type == "BelongsTo" {
			fkName := rel.Model + "ID"
			sb.WriteString(fmt.Sprintf("\t%s uint `json:\"%s\" db:\"%s\"`\n",
				fkName,
				toSnakeCase(fkName),
				toSnakeCase(fkName)))
		}
	}
	
	// Add timestamps
	if config.Timestamps {
		sb.WriteString("\tCreatedAt time.Time `json:\"created_at\" db:\"created_at\"`\n")
		sb.WriteString("\tUpdatedAt time.Time `json:\"updated_at\" db:\"updated_at\"`\n")
	}
	
	// Add soft delete
	if config.SoftDelete {
		sb.WriteString("\tDeletedAt *time.Time `json:\"deleted_at,omitempty\" db:\"deleted_at\"`\n")
	}
	
	sb.WriteString("}\n\n")
	
	// Generate validation method
	if config.Validation {
		sb.WriteString(g.generateValidationMethod(config))
	}
	
	// Generate JSON methods
	if config.JSONMethods {
		sb.WriteString(g.generateJSONMethods(config))
	}
	
	return sb.String()
}

func (g *ModelGenerator) generateValidationMethod(config ModelConfig) string {
	return fmt.Sprintf(`// Validate validates the %s model
func (m *%s) Validate() error {
	// TODO: Add custom validation logic
	return nil
}

`, config.Name, config.Name)
}

func (g *ModelGenerator) generateJSONMethods(config ModelConfig) string {
	return fmt.Sprintf(`// MarshalJSON customizes JSON marshaling
func (m *%s) MarshalJSON() ([]byte, error) {
	// TODO: Implement custom JSON marshaling
	return nil, nil
}

// UnmarshalJSON customizes JSON unmarshaling
func (m *%s) UnmarshalJSON(data []byte) error {
	// TODO: Implement custom JSON unmarshaling
	return nil
}

`, config.Name, config.Name)
}

// GenerateRepository generates a repository interface
func (g *ModelGenerator) GenerateRepository(targetPath string, config ModelConfig) error {
	repoName := config.Name + "Repository"
	fileName := strings.ToLower(config.Name) + "_repository.go"
	
	content := fmt.Sprintf(`package repository

import (
	"context"
	
	"your-module/internal/models"
)

// %s defines the interface for %s data access
type %s interface {
	Create(ctx context.Context, item *models.%s) (*models.%s, error)
	GetByID(ctx context.Context, id string) (*models.%s, error)
	List(ctx context.Context) ([]*models.%s, error)
	Update(ctx context.Context, id string, item *models.%s) (*models.%s, error)
	Delete(ctx context.Context, id string) error
}

// %sImpl implements %s
type %sImpl struct {
	// db or storage dependency here
}

// New%s creates a new %s repository
func New%s() *%sImpl {
	return &%sImpl{}
}

// Create creates a new %s
func (r *%sImpl) Create(ctx context.Context, item *models.%s) (*models.%s, error) {
	// TODO: Implement create
	return item, nil
}

// GetByID retrieves a %s by ID
func (r *%sImpl) GetByID(ctx context.Context, id string) (*models.%s, error) {
	// TODO: Implement get by ID
	return nil, nil
}

// List retrieves all %ss
func (r *%sImpl) List(ctx context.Context) ([]*models.%s, error) {
	// TODO: Implement list
	return nil, nil
}

// Update updates a %s
func (r *%sImpl) Update(ctx context.Context, id string, item *models.%s) (*models.%s, error) {
	// TODO: Implement update
	return item, nil
}

// Delete deletes a %s
func (r *%sImpl) Delete(ctx context.Context, id string) error {
	// TODO: Implement delete
	return nil
}
`,
		repoName, config.Name, repoName, config.Name, config.Name, config.Name,
		config.Name, config.Name, config.Name,
		config.Name, repoName, config.Name,
		config.Name, config.Name, config.Name, config.Name, config.Name,
		config.Name, config.Name, config.Name, config.Name,
		config.Name, config.Name, config.Name,
		strings.ToLower(config.Name), config.Name, config.Name,
		config.Name, config.Name, config.Name, config.Name,
		config.Name, config.Name,
	)
	
	repoDir := filepath.Join(targetPath, "internal", "repository")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return err
	}
	
	repoPath := filepath.Join(repoDir, fileName)
	return os.WriteFile(repoPath, []byte(content), 0644)
}

// GenerateMultiple generates multiple models
func (g *ModelGenerator) GenerateMultiple(targetPath string, configs []ModelConfig) error {
	for _, config := range configs {
		if err := g.GenerateModel(targetPath, config); err != nil {
			return fmt.Errorf("failed to generate %s model: %w", config.Name, err)
		}
	}
	return nil
}

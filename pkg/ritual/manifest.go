package ritual

// Manifest represents the complete ritual.yaml definition
type Manifest struct {
	Ritual        RitualMeta    `yaml:"ritual"`
	Compatibility Compatibility `yaml:"compatibility,omitempty"`
	Dependencies  Dependencies  `yaml:"dependencies,omitempty"`
	Questions     []Question    `yaml:"questions,omitempty"`
	Files         FilesSection  `yaml:"files,omitempty"`
	Migrations    []Migration   `yaml:"migrations,omitempty"`
	Hooks         ManifestHooks `yaml:"hooks,omitempty"`
	MultiTenancy  *MultiTenancy `yaml:"multi_tenancy,omitempty"`
	Telemetry     *Telemetry    `yaml:"telemetry,omitempty"`
	Parent        *ParentRitual `yaml:"parent,omitempty"`
}

// RitualMeta contains ritual metadata
type RitualMeta struct {
	Name           string   `yaml:"name"`
	Version        string   `yaml:"version"`
	Description    string   `yaml:"description"`
	Author         string   `yaml:"author,omitempty"`
	License        string   `yaml:"license,omitempty"`
	Homepage       string   `yaml:"homepage,omitempty"`
	Repository     string   `yaml:"repository,omitempty"`
	Tags           []string `yaml:"tags,omitempty"`
	TemplateEngine string   `yaml:"template_engine,omitempty"` // fith, go-template, custom
}

// Compatibility defines version requirements
type Compatibility struct {
	MinToutaVersion string `yaml:"min_touta_version,omitempty"`
	MaxToutaVersion string `yaml:"max_touta_version,omitempty"`
	MinGoVersion    string `yaml:"min_go_version,omitempty"`
	MaxGoVersion    string `yaml:"max_go_version,omitempty"`
}

// Dependencies defines required packages and rituals
type Dependencies struct {
	Packages []string             `yaml:"packages,omitempty"`
	Rituals  []string             `yaml:"rituals,omitempty"`
	Database *DatabaseRequirement `yaml:"database,omitempty"`
}

// DatabaseRequirement specifies database needs
type DatabaseRequirement struct {
	Required   bool     `yaml:"required"`
	Types      []string `yaml:"types,omitempty"` // postgres, mysql, sqlite
	MinVersion string   `yaml:"min_version,omitempty"`
}

// QuestionType represents the type of question
type QuestionType string

const (
	QuestionTypeText        QuestionType = "text"
	QuestionTypePassword    QuestionType = "password"
	QuestionTypeChoice      QuestionType = "choice"
	QuestionTypeMultiChoice QuestionType = "multi_choice"
	QuestionTypeBoolean     QuestionType = "boolean"
	QuestionTypeNumber      QuestionType = "number"
	QuestionTypePath        QuestionType = "path"
	QuestionTypeURL         QuestionType = "url"
	QuestionTypeEmail       QuestionType = "email"
)

// Question represents an interactive prompt
type Question struct {
	Name      string             `yaml:"name"`
	Prompt    string             `yaml:"prompt"`
	Type      QuestionType       `yaml:"type"`
	Required  bool               `yaml:"required,omitempty"`
	Default   interface{}        `yaml:"default,omitempty"`
	Choices   []string           `yaml:"choices,omitempty"`
	Validate  *ValidationRule    `yaml:"validate,omitempty"`
	Condition *QuestionCondition `yaml:"condition,omitempty"`
	Helper    *QuestionHelper    `yaml:"helper,omitempty"`
	Group     string             `yaml:"group,omitempty"`
	Step      int                `yaml:"step,omitempty"`
}

// ValidationRule defines validation constraints
type ValidationRule struct {
	Pattern string `yaml:"pattern,omitempty"` // regex pattern
	Min     *int   `yaml:"min,omitempty"`
	Max     *int   `yaml:"max,omitempty"`
	MinLen  *int   `yaml:"min_len,omitempty"`
	MaxLen  *int   `yaml:"max_len,omitempty"`
	Custom  string `yaml:"custom,omitempty"` // custom validator name
}

// QuestionCondition defines conditional display
type QuestionCondition struct {
	Field      string              `yaml:"field,omitempty"`
	Equals     interface{}         `yaml:"equals,omitempty"`
	NotEquals  interface{}         `yaml:"not_equals,omitempty"`
	Expression string              `yaml:"expression,omitempty"`
	And        []QuestionCondition `yaml:"and,omitempty"`
	Or         []QuestionCondition `yaml:"or,omitempty"`
	Not        *QuestionCondition  `yaml:"not,omitempty"`
}

// QuestionHelper provides helper tools
type QuestionHelper struct {
	Type   string                 `yaml:"type"` // db_test, url_check, path_check
	Config map[string]interface{} `yaml:"config,omitempty"`
}

// FilesSection defines template and static files
type FilesSection struct {
	Templates   []FileMapping `yaml:"templates,omitempty"`
	Static      []FileMapping `yaml:"static,omitempty"`
	Directories []string      `yaml:"directories,omitempty"` // directories to create
	Protected   []string      `yaml:"protected,omitempty"`   // files to never overwrite
}

// FileMapping maps source to destination
type FileMapping struct {
	Source      string `yaml:"src"`
	Destination string `yaml:"dest"`
	Optional    bool   `yaml:"optional,omitempty"`
	Condition   string `yaml:"condition,omitempty"`
}

// Migration represents a version migration
type Migration struct {
	FromVersion string           `yaml:"from_version"`
	ToVersion   string           `yaml:"to_version"`
	Description string           `yaml:"description,omitempty"`
	Up          MigrationHandler `yaml:"up"`
	Down        MigrationHandler `yaml:"down"`
	Idempotent  bool             `yaml:"idempotent,omitempty"`
}

// MigrationHandler defines migration logic
type MigrationHandler struct {
	SQL    []string `yaml:"sql,omitempty"`
	Script string   `yaml:"script,omitempty"`
	GoCode string   `yaml:"go_code,omitempty"`
}

// ManifestHooks defines lifecycle hooks
type ManifestHooks struct {
	PreInstall  []string `yaml:"pre_install,omitempty"`
	PostInstall []string `yaml:"post_install,omitempty"`
	PreUpdate   []string `yaml:"pre_update,omitempty"`
	PostUpdate  []string `yaml:"post_update,omitempty"`
	PreDeploy   []string `yaml:"pre_deploy,omitempty"`
	PostDeploy  []string `yaml:"post_deploy,omitempty"`
}

// MultiTenancy defines multi-tenant settings
type MultiTenancy struct {
	Enabled       bool   `yaml:"enabled"`
	DatabaseMode  string `yaml:"database_mode"` // shared, separate
	TenantIDField string `yaml:"tenant_id_field,omitempty"`
}

// Telemetry defines monitoring settings
type Telemetry struct {
	Enabled bool              `yaml:"enabled,omitempty"`
	Metrics []string          `yaml:"metrics,omitempty"`
	Config  map[string]string `yaml:"config,omitempty"`
}

// ParentRitual defines ritual inheritance
type ParentRitual struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Source  string `yaml:"source,omitempty"` // git URL or tarball
}

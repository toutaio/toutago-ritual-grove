package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// Validator validates ritual manifests
type Validator struct{
	ritualPath string
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// SetRitualPath sets the base path for the ritual (for file reference validation)
func (v *Validator) SetRitualPath(path string) {
	v.ritualPath = path
}

// Validate validates a ritual manifest
func (v *Validator) Validate(manifest *ritual.Manifest) error {
	if err := v.validateMetadata(manifest); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	if err := v.validateCompatibility(manifest); err != nil {
		return fmt.Errorf("compatibility validation failed: %w", err)
	}

	if err := v.validateQuestions(manifest); err != nil {
		return fmt.Errorf("questions validation failed: %w", err)
	}

	if err := v.validateFiles(manifest); err != nil {
		return fmt.Errorf("files validation failed: %w", err)
	}

	if err := v.validateMigrations(manifest); err != nil {
		return fmt.Errorf("migrations validation failed: %w", err)
	}

	return nil
}

func (v *Validator) validateMetadata(manifest *ritual.Manifest) error {
	if manifest.Ritual.Name == "" {
		return fmt.Errorf("ritual name is required")
	}

	// Validate name format (lowercase, alphanumeric, hyphens)
	if !regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(manifest.Ritual.Name) {
		return fmt.Errorf("ritual name must start with lowercase letter and contain only lowercase letters, numbers, and hyphens")
	}

	if manifest.Ritual.Version == "" {
		return fmt.Errorf("ritual version is required")
	}

	// Validate semantic version format
	if !v.isValidSemver(manifest.Ritual.Version) {
		return fmt.Errorf("ritual version must be valid semantic version (e.g., 1.0.0)")
	}

	// Validate template engine
	if manifest.Ritual.TemplateEngine != "" {
		validEngines := []string{"fith", "go-template"}
		if !contains(validEngines, manifest.Ritual.TemplateEngine) {
			return fmt.Errorf("invalid template engine: %s (must be one of: %s)",
				manifest.Ritual.TemplateEngine, strings.Join(validEngines, ", "))
		}
	}

	return nil
}

func (v *Validator) validateCompatibility(manifest *ritual.Manifest) error {
	if manifest.Compatibility.MinToutaVersion != "" {
		if !v.isValidSemver(manifest.Compatibility.MinToutaVersion) {
			return fmt.Errorf("min_touta_version must be valid semantic version")
		}
	}

	if manifest.Compatibility.MaxToutaVersion != "" {
		if !v.isValidSemver(manifest.Compatibility.MaxToutaVersion) {
			return fmt.Errorf("max_touta_version must be valid semantic version")
		}
	}

	if manifest.Compatibility.MinGoVersion != "" {
		if !v.isValidGoVersion(manifest.Compatibility.MinGoVersion) {
			return fmt.Errorf("min_go_version must be valid Go version")
		}
	}

	return nil
}

func (v *Validator) validateQuestions(manifest *ritual.Manifest) error {
	questionNames := make(map[string]bool)

	for i, q := range manifest.Questions {
		if q.Name == "" {
			return fmt.Errorf("question %d: name is required", i)
		}

		// Check for duplicate names
		if questionNames[q.Name] {
			return fmt.Errorf("question %s: duplicate question name", q.Name)
		}
		questionNames[q.Name] = true

		if q.Prompt == "" {
			return fmt.Errorf("question %s: prompt is required", q.Name)
		}

		if q.Type == "" {
			return fmt.Errorf("question %s: type is required", q.Name)
		}

		// Validate type-specific requirements
		switch q.Type {
		case ritual.QuestionTypeChoice, ritual.QuestionTypeMultiChoice:
			if len(q.Choices) == 0 {
				return fmt.Errorf("question %s: choices required for choice type", q.Name)
			}
		}

		// Validate conditions reference existing questions
		if q.Condition != nil {
			if q.Condition.Field == "" {
				return fmt.Errorf("question %s: condition field is required", q.Name)
			}
			// Note: We can't validate if field exists yet since questions are processed in order
			// This would need a second pass
		}

		// Validate validation rules
		if q.Validate != nil {
			if q.Validate.Pattern != "" {
				if _, err := regexp.Compile(q.Validate.Pattern); err != nil {
					return fmt.Errorf("question %s: invalid regex pattern: %w", q.Name, err)
				}
			}
		}
	}

	return nil
}

func (v *Validator) validateFiles(manifest *ritual.Manifest) error {
	// Validate template mappings
	for i, tmpl := range manifest.Files.Templates {
		if tmpl.Source == "" {
			return fmt.Errorf("template %d: source is required", i)
		}
		if tmpl.Destination == "" {
			return fmt.Errorf("template %d: destination is required", i)
		}
	}

	// Validate static file mappings
	for i, static := range manifest.Files.Static {
		if static.Source == "" {
			return fmt.Errorf("static file %d: source is required", i)
		}
		if static.Destination == "" {
			return fmt.Errorf("static file %d: destination is required", i)
		}
	}

	return nil
}

func (v *Validator) validateMigrations(manifest *ritual.Manifest) error {
	for i, m := range manifest.Migrations {
		if m.FromVersion == "" {
			return fmt.Errorf("migration %d: from_version is required", i)
		}
		if m.ToVersion == "" {
			return fmt.Errorf("migration %d: to_version is required", i)
		}

		if !v.isValidSemver(m.FromVersion) {
			return fmt.Errorf("migration %d: from_version must be valid semantic version", i)
		}
		if !v.isValidSemver(m.ToVersion) {
			return fmt.Errorf("migration %d: to_version must be valid semantic version", i)
		}

		// Check that at least one up handler is defined
		if len(m.Up.SQL) == 0 && m.Up.Script == "" && m.Up.GoCode == "" {
			return fmt.Errorf("migration %s->%s: at least one up handler (sql, script, or go_code) is required",
				m.FromVersion, m.ToVersion)
		}

		// Warn if no down handler (but don't error)
		if len(m.Down.SQL) == 0 && m.Down.Script == "" && m.Down.GoCode == "" {
			// Down handler is optional but recommended
		}
	}

	return nil
}

// isValidSemver checks if a version string is valid semantic version
func (v *Validator) isValidSemver(version string) bool {
	pattern := `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
	return regexp.MustCompile(pattern).MatchString(version)
}

// isValidGoVersion checks if a version string is valid Go version
func (v *Validator) isValidGoVersion(version string) bool {
	pattern := `^(0|[1-9]\d*)\.(0|[1-9]\d*)(?:\.(0|[1-9]\d*))?$`
	return regexp.MustCompile(pattern).MatchString(version)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidateFileReferences checks that all referenced template/static files exist
func (v *Validator) ValidateFileReferences(manifest *ritual.Manifest) error {
if v.ritualPath == "" {
return nil // Skip if no path set
}

for _, tmpl := range manifest.Files.Templates {
// Check if source file exists
filePath := filepath.Join(v.ritualPath, tmpl.Source)
if _, err := os.Stat(filePath); os.IsNotExist(err) {
return fmt.Errorf("template file not found: %s", tmpl.Source)
}
}

for _, static := range manifest.Files.Static {
filePath := filepath.Join(v.ritualPath, static.Source)
if _, err := os.Stat(filePath); os.IsNotExist(err) {
return fmt.Errorf("static file not found: %s", static.Source)
}
}

return nil
}

// ValidateVersionConstraints checks that version constraints are logically valid
func (v *Validator) ValidateVersionConstraints(manifest *ritual.Manifest) error {
if manifest.Compatibility.MinToutaVersion != "" && manifest.Compatibility.MaxToutaVersion != "" {
// Simple string comparison (more sophisticated semver comparison could be added)
min := manifest.Compatibility.MinToutaVersion
max := manifest.Compatibility.MaxToutaVersion

if compareVersions(min, max) > 0 {
return fmt.Errorf("min_touta_version (%s) is greater than max_touta_version (%s)", min, max)
}
}

return nil
}

// ValidateQuestionConditions validates question conditional logic
func (v *Validator) ValidateQuestionConditions(manifest *ritual.Manifest) error {
// Build map of question names
questionNames := make(map[string]bool)
for _, q := range manifest.Questions {
questionNames[q.Name] = true
}

// Check each question's conditions
for _, q := range manifest.Questions {
if q.Condition != nil {
if err := v.validateCondition(q.Condition, questionNames, q.Name); err != nil {
return fmt.Errorf("invalid condition for question %s: %w", q.Name, err)
}
}
}

// Check for circular dependencies
if err := v.detectCircularConditions(manifest.Questions); err != nil {
return err
}

return nil
}

func (v *Validator) validateCondition(cond *ritual.QuestionCondition, validNames map[string]bool, currentQuestion string) error {
if cond.Field != "" {
if !validNames[cond.Field] {
return fmt.Errorf("condition references non-existent field: %s", cond.Field)
}
if cond.Field == currentQuestion {
return fmt.Errorf("question cannot depend on itself")
}
}

// Recursively validate And/Or/Not conditions
for _, subCond := range cond.And {
if err := v.validateCondition(&subCond, validNames, currentQuestion); err != nil {
return err
}
}
for _, subCond := range cond.Or {
if err := v.validateCondition(&subCond, validNames, currentQuestion); err != nil {
return err
}
}
if cond.Not != nil {
if err := v.validateCondition(cond.Not, validNames, currentQuestion); err != nil {
return err
}
}

return nil
}

func (v *Validator) detectCircularConditions(questions []ritual.Question) error {
// Build dependency graph
deps := make(map[string][]string)
for _, q := range questions {
if q.Condition != nil {
deps[q.Name] = v.extractDependencies(q.Condition)
}
}

// Check for cycles using DFS
visited := make(map[string]bool)
recStack := make(map[string]bool)

for question := range deps {
if v.hasCycle(question, deps, visited, recStack) {
return fmt.Errorf("circular dependency detected in question conditions involving: %s", question)
}
}

return nil
}

func (v *Validator) extractDependencies(cond *ritual.QuestionCondition) []string {
var deps []string
if cond.Field != "" {
deps = append(deps, cond.Field)
}
for _, subCond := range cond.And {
deps = append(deps, v.extractDependencies(&subCond)...)
}
for _, subCond := range cond.Or {
deps = append(deps, v.extractDependencies(&subCond)...)
}
if cond.Not != nil {
deps = append(deps, v.extractDependencies(cond.Not)...)
}
return deps
}

func (v *Validator) hasCycle(node string, graph map[string][]string, visited, recStack map[string]bool) bool {
visited[node] = true
recStack[node] = true

for _, neighbor := range graph[node] {
if !visited[neighbor] {
if v.hasCycle(neighbor, graph, visited, recStack) {
return true
}
} else if recStack[neighbor] {
return true
}
}

recStack[node] = false
return false
}

// CheckCommonMistakes returns warnings for common ritual authoring mistakes
func (v *Validator) CheckCommonMistakes(manifest *ritual.Manifest) []string {
var warnings []string

// Check if config/env files are protected
protectedMap := make(map[string]bool)
for _, p := range manifest.Files.Protected {
protectedMap[p] = true
}

for _, tmpl := range manifest.Files.Templates {
dest := tmpl.Destination

// Warn about unprotected config files
if (strings.Contains(dest, "config") && (strings.Contains(dest, ".yaml") || strings.Contains(dest, ".yml") || strings.Contains(dest, ".json"))) {
if !protectedMap[dest] && !matchesAnyPattern(dest, manifest.Files.Protected) {
warnings = append(warnings, fmt.Sprintf("Consider protecting config file: %s", dest))
}
}

// Warn about unprotected .env files
if strings.Contains(dest, ".env") {
if !protectedMap[dest] && !matchesAnyPattern(dest, manifest.Files.Protected) {
warnings = append(warnings, fmt.Sprintf("Consider protecting environment file: %s", dest))
}
}
}

// Warn if no tests are generated
hasTests := false
for _, tmpl := range manifest.Files.Templates {
if strings.Contains(tmpl.Destination, "_test.go") {
hasTests = true
break
}
}
if !hasTests {
warnings = append(warnings, "No test files found - consider adding generated tests")
}

return warnings
}

// CheckMigrationReversibility returns warnings for migrations without down handlers
func (v *Validator) CheckMigrationReversibility(manifest *ritual.Manifest) []string {
var warnings []string

for _, m := range manifest.Migrations {
hasDown := len(m.Down.SQL) > 0 || m.Down.Script != "" || m.Down.GoCode != ""
if !hasDown {
warnings = append(warnings, fmt.Sprintf("Migration %s->%s lacks down handler for rollback",
m.FromVersion, m.ToVersion))
}
}

return warnings
}

func matchesAnyPattern(path string, patterns []string) bool {
for _, pattern := range patterns {
if matched, _ := filepath.Match(pattern, path); matched {
return true
}
if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
return true
}
}
return false
}

func compareVersions(v1, v2 string) int {
parts1 := strings.Split(v1, ".")
parts2 := strings.Split(v2, ".")

for i := 0; i < len(parts1) && i < len(parts2); i++ {
if parts1[i] < parts2[i] {
return -1
}
if parts1[i] > parts2[i] {
return 1
}
}

if len(parts1) < len(parts2) {
return -1
}
if len(parts1) > len(parts2) {
return 1
}
return 0
}

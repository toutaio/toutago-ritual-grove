package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/toutaio/toutago-ritual-grove/internal/commands"
	"github.com/toutaio/toutago-ritual-grove/internal/generator"
	"github.com/toutaio/toutago-ritual-grove/internal/questionnaire"
	"github.com/toutaio/toutago-ritual-grove/internal/registry"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// RitualCommand creates the main ritual command with subcommands
func RitualCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ritual",
		Short: "Manage project rituals (scaffolding templates)",
		Long: `Rituals are reusable project templates for quickly scaffolding
new Toutā applications with best practices and common patterns.

Use rituals to create:
  - Basic websites
  - Blogs
  - APIs
  - Custom project types`,
	}

	// Add subcommands
	cmd.AddCommand(initCommand())
	cmd.AddCommand(listCommand())
	cmd.AddCommand(infoCommand())
	cmd.AddCommand(validateCommand())
	cmd.AddCommand(createCommand())
	cmd.AddCommand(planCommand())
	cmd.AddCommand(searchCommand())
	cmd.AddCommand(updateCommand())
	cmd.AddCommand(migrateCommand())

	return cmd
}

// initCommand initializes a project from a ritual
func initCommand() *cobra.Command {
	var outputPath string
	var skipQuestions bool
	var initGit bool

	cmd := &cobra.Command{
		Use:   "init <ritual-name>",
		Short: "Initialize a new project from a ritual",
		Long: `Initialize a new project from a ritual template.

The ritual will ask questions about your project and generate
the appropriate files and structure based on your answers.

Example:
  touta ritual init basic-site
  touta ritual init blog --output ./my-blog
  touta ritual init blog --git --output ./my-blog`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ritualName := args[0]
			if outputPath == "" {
				outputPath = "."
			}
			return initRitual(ritualName, outputPath, skipQuestions, initGit)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory (default: current directory)")
	cmd.Flags().BoolVar(&skipQuestions, "yes", false, "Skip questions and use defaults")
	cmd.Flags().BoolVar(&initGit, "git", false, "Initialize git repository after creation")

	return cmd
}

// listCommand lists available rituals
func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available rituals",
		Long: `List all available rituals from the local registry.

Rituals are stored in:
  - Built-in: <ritual-grove>/rituals/
  - Local: ~/.touta/rituals/
  - Project: ./rituals/`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listRituals()
		},
	}

	return cmd
}

// infoCommand shows information about a ritual
func infoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <ritual-name>",
		Short: "Show detailed information about a ritual",
		Long: `Display detailed information about a specific ritual including:
  - Name and version
  - Description
  - Author
  - Questions that will be asked
  - Files that will be generated
  - Dependencies required`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ritualName := args[0]
			return showRitualInfo(ritualName)
		},
	}

	return cmd
}

// validateCommand validates a ritual.yaml file
func validateCommand() *cobra.Command {
	var ritualPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a ritual.yaml file",
		Long: `Validate a ritual.yaml file for correctness.

This checks:
  - YAML syntax
  - Required fields
  - Version format
  - Template references
  - Migration structure`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if ritualPath == "" {
				ritualPath = "."
			}
			return validateRitual(ritualPath)
		},
	}

	cmd.Flags().StringVarP(&ritualPath, "path", "p", "", "Path to ritual directory (default: current directory)")

	return cmd
}

// createCommand creates a new ritual template
func createCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new ritual template",
		Long: `Create a new ritual template with the basic structure.

This will create:
  - ritual.yaml with basic metadata
  - templates/ directory
  - static/ directory
  - migrations/ directory (optional)
  - README.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ritualName := args[0]
			return createRitual(ritualName)
		},
	}

	return cmd
}

// planCommand shows deployment plan for updates
func planCommand() *cobra.Command {
	return commands.NewPlanCommand()
}

// initRitual initializes a project from a ritual
func initRitual(ritualName, outputPath string, skipQuestions bool, initGit bool) error {
	// Create registry
	reg := registry.NewRegistry()

	// Scan for rituals
	if err := reg.Scan(); err != nil {
		return fmt.Errorf("failed to scan for rituals: %w", err)
	}

	// Find ritual in registry
	ritualMeta, err := reg.Get(ritualName)
	if err != nil {
		return fmt.Errorf("ritual %q not found: %w\n\nTry 'touta ritual list' to see available rituals", ritualName, err)
	}

	fmt.Printf("🌱 Initializing project from ritual: %s\n\n", ritualName)

	// Load ritual manifest
	manifest, err := reg.Load(ritualName)
	if err != nil {
		return fmt.Errorf("failed to load ritual manifest: %w", err)
	}

	// Validate manifest
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("invalid ritual: %w", err)
	}

	// Run questionnaire
	variables := make(map[string]interface{})
	if !skipQuestions && len(manifest.Questions) > 0 {
		adapter := questionnaire.NewCLIAdapter(manifest.Questions, nil)
		answers, err := adapter.Run()
		if err != nil {
			return fmt.Errorf("questionnaire failed: %w", err)
		}
		variables = answers
	} else {
		// Use defaults
		for _, question := range manifest.Questions {
			if question.Default != nil {
				variables[question.Name] = question.Default
			}
		}
	}

	// Add project metadata variables
	projectName := filepath.Base(outputPath)
	if projectName == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		projectName = filepath.Base(cwd)
	}

	// Generate module path (github.com/user/project or example.com/project)
	modulePath := fmt.Sprintf("example.com/%s", projectName)
	if userVar, ok := variables["github_user"]; ok {
		modulePath = fmt.Sprintf("github.com/%s/%s", userVar, projectName)
	}

	variables["project_name"] = projectName
	variables["module_path"] = modulePath
	variables["ritual_name"] = ritualName
	variables["ritual_version"] = manifest.Ritual.Version

	// Generate files
	gen := generator.NewFileGenerator("go")
	vars := generator.NewVariables()
	for k, v := range variables {
		vars.Set(k, v)
	}
	gen.SetVariables(vars)

	fmt.Printf("📝 Generating project files...\n")
	if err := gen.GenerateFiles(manifest, ritualMeta.Path, outputPath); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	// Initialize git repository if requested
	if initGit {
		if err := initGitRepository(outputPath); err != nil {
			fmt.Printf("⚠️  Warning: failed to initialize git repository: %v\n", err)
		} else {
			fmt.Printf("✓ Initialized git repository\n")
		}
	}

	fmt.Printf("\n✅ Project initialized successfully!\n\n")
	fmt.Printf("Next steps:\n")
	fmt.Printf("  cd %s\n", outputPath)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  touta serve\n\n")

	return nil
}

// initGitRepository initializes a git repository (duplicated from internal/cli/create.go for now)
func initGitRepository(targetPath string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = targetPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %s: %w", string(output), err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = targetPath
	if err := cmd.Run(); err != nil {
		return nil
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit from ritual")
	cmd.Dir = targetPath
	_ = cmd.Run()

	return nil
}

// listRituals lists available rituals
func listRituals() error {
	reg := registry.NewRegistry()

	// Scan for rituals
	if err := reg.Scan(); err != nil {
		return fmt.Errorf("failed to scan for rituals: %w", err)
	}

	rituals := reg.List()

	if len(rituals) == 0 {
		fmt.Println("No rituals found.")
		fmt.Println("\nTo create a ritual, use: touta ritual create <name>")
		return nil
	}

	fmt.Println("Available rituals:")
	fmt.Println()
	for _, r := range rituals {
		fmt.Printf("  📦 %s (%s)\n", r.Name, r.Version)
		if r.Description != "" {
			fmt.Printf("     %s\n", r.Description)
		}
		fmt.Println()
	}

	return nil
}

// showRitualInfo shows detailed information about a ritual
func showRitualInfo(ritualName string) error {
	reg := registry.NewRegistry()

	// Scan for rituals
	if err := reg.Scan(); err != nil {
		return fmt.Errorf("failed to scan for rituals: %w", err)
	}

	ritualMeta, err := reg.Get(ritualName)
	if err != nil {
		return fmt.Errorf("ritual %q not found: %w", ritualName, err)
	}

	manifest, err := reg.Load(ritualName)
	if err != nil {
		return fmt.Errorf("failed to load ritual manifest: %w", err)
	}

	fmt.Printf("📦 %s\n\n", manifest.Ritual.Name)
	fmt.Printf("Version:     %s\n", manifest.Ritual.Version)
	fmt.Printf("Description: %s\n", manifest.Ritual.Description)
	if manifest.Ritual.Author != "" {
		fmt.Printf("Author:      %s\n", manifest.Ritual.Author)
	}

	fmt.Println("\nCompatibility:")
	if manifest.Compatibility.MinToutaVersion != "" {
		fmt.Printf("  Min Toutā version: %s\n", manifest.Compatibility.MinToutaVersion)
	}
	if manifest.Compatibility.MinGoVersion != "" {
		fmt.Printf("  Go version:        %s\n", manifest.Compatibility.MinGoVersion)
	}

	if len(manifest.Dependencies.Packages) > 0 {
		fmt.Println("\nGo Dependencies:")
		for _, pkg := range manifest.Dependencies.Packages {
			fmt.Printf("  - %s\n", pkg)
		}
	}

	if len(manifest.Questions) > 0 {
		fmt.Printf("\nQuestions (%d):\n", len(manifest.Questions))
		for _, q := range manifest.Questions {
			required := ""
			if q.Required {
				required = " (required)"
			}
			fmt.Printf("  - %s: %s%s\n", q.Name, q.Prompt, required)
		}
	}

	templateCount := len(manifest.Files.Templates)
	staticCount := len(manifest.Files.Static)
	fmt.Printf("\nFiles: %d templates, %d static files\n", templateCount, staticCount)
	fmt.Printf("Path: %s\n", ritualMeta.Path)

	return nil
}

// validateRitual validates a ritual.yaml file
func validateRitual(ritualPath string) error {
	manifestPath := filepath.Join(ritualPath, "ritual.yaml")

	// Check if file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("ritual.yaml not found in %s", ritualPath)
	}

	// Load manifest
	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	if err != nil {
		return fmt.Errorf("failed to load ritual.yaml: %w", err)
	}

	// Validate
	if err := manifest.Validate(); err != nil {
		fmt.Printf("❌ Validation failed:\n\n")
		return err
	}

	fmt.Printf("✅ Ritual is valid!\n\n")
	fmt.Printf("Name:    %s\n", manifest.Ritual.Name)
	fmt.Printf("Version: %s\n", manifest.Ritual.Version)

	return nil
}

// createRitual creates a new ritual template
func createRitual(ritualName string) error {
	// Create ritual directory
	if err := os.MkdirAll(ritualName, 0750); err != nil {
		return fmt.Errorf("failed to create ritual directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{
		filepath.Join(ritualName, "templates"),
		filepath.Join(ritualName, "static"),
		filepath.Join(ritualName, "migrations"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create ritual.yaml
	ritualYAML := fmt.Sprintf(`ritual:
  name: %s
  version: 1.0.0
  description: A custom ritual template
  author: ""

compatibility:
  min_touta_version: "0.1.0"
  min_go_version: "1.22"

questions:
  - name: project_name
    type: text
    prompt: "What is your project name?"
    required: true
    default: "my-project"

files:
  templates: []
  static: []
  protected: []

hooks:
  pre_install: []
  post_install: []
`, ritualName)

	ritualYAMLPath := filepath.Join(ritualName, "ritual.yaml")
	if err := os.WriteFile(ritualYAMLPath, []byte(ritualYAML), 0600); err != nil {
		return fmt.Errorf("failed to create ritual.yaml: %w", err)
	}

	// Create README.md
	readme := fmt.Sprintf(`# %s Ritual

## Description

A custom ritual template for Toutā projects.

## Usage

`+"```bash"+`
touta ritual init %s
`+"```"+`

## Questions

- **project_name**: The name of your project

## Generated Files

TODO: Document what files this ritual generates

## Requirements

- Toutā 0.1.0+
- Go 1.22+
`, ritualName, ritualName)

	readmePath := filepath.Join(ritualName, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0600); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	fmt.Printf("✅ Created ritual template: %s\n\n", ritualName)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Edit %s/ritual.yaml\n", ritualName)
	fmt.Printf("  2. Add templates to %s/templates/\n", ritualName)
	fmt.Printf("  3. Add static files to %s/static/\n", ritualName)
	fmt.Printf("  4. Test with: touta ritual validate --path %s\n\n", ritualName)

	return nil
}

// searchCommand searches for rituals
func searchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for rituals by name, tag, or description",
		Long: `Search for available rituals in the registry.

Example:
  touta ritual search blog
  touta ritual search api`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			return searchRituals(query)
		},
	}

	return cmd
}

// updateCommand updates a project to a new ritual version
func updateCommand() *cobra.Command {
	var toVersion string
	var dryRun bool
	var force bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update project to new ritual version",
		Long: `Update the current project to a new version of its ritual.

This command will:
  - Check the current ritual version
  - Run migrations if needed
  - Create backups before updating
  - Rollback on error (unless --force)

Example:
  touta ritual update --to 1.2.0
  touta ritual update --to 1.2.0 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateProject(".", toVersion, dryRun, force)
		},
	}

	cmd.Flags().StringVar(&toVersion, "to", "", "Target version to update to (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without making changes")
	cmd.Flags().BoolVar(&force, "force", false, "Force update even if migrations fail")
	if err := cmd.MarkFlagRequired("to"); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}

	return cmd
}

// migrateCommand runs pending migrations
func migrateCommand() *cobra.Command {
	var up bool
	var down bool
	var toVersion string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run pending migrations",
		Long: `Run pending migrations for the current project.

Example:
  touta ritual migrate --up
  touta ritual migrate --down --to 1.0.0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !up && !down {
				return fmt.Errorf("must specify --up or --down")
			}
			if up && down {
				return fmt.Errorf("cannot specify both --up and --down")
			}
			return runMigrations(".", up, toVersion)
		},
	}

	cmd.Flags().BoolVar(&up, "up", false, "Run forward migrations")
	cmd.Flags().BoolVar(&down, "down", false, "Run rollback migrations")
	cmd.Flags().StringVar(&toVersion, "to", "", "Target version to migrate to")

	return cmd
}

// searchRituals searches for rituals matching a query
func searchRituals(query string) error {
	reg := registry.NewRegistry()

	if err := reg.Scan(); err != nil {
		return fmt.Errorf("failed to scan for rituals: %w", err)
	}

	results := reg.Search(query)
	if len(results) == 0 {
		fmt.Printf("No rituals found matching '%s'\n", query)
		return nil
	}

	fmt.Printf("Found %d ritual(s) matching '%s':\n\n", len(results), query)
	for _, r := range results {
		fmt.Printf("  📦 %s (%s)\n", r.Name, r.Version)
		if r.Description != "" {
			fmt.Printf("     %s\n", r.Description)
		}
		if len(r.Tags) > 0 {
			fmt.Printf("     Tags: %v\n", r.Tags)
		}
		fmt.Println()
	}

	return nil
}

// updateProject updates a project to a new ritual version
func updateProject(projectPath, toVersion string, dryRun, force bool) error {
	handler := commands.NewUpdateHandler()

	opts := commands.UpdateOptions{
		ToVersion: toVersion,
		DryRun:    dryRun,
		Force:     force,
	}

	return handler.Execute(projectPath, opts)
}

// runMigrations runs pending migrations
func runMigrations(projectPath string, up bool, toVersion string) error {
	return fmt.Errorf("migration command not yet implemented - use 'update' command instead")
}

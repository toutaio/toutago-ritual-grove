package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/toutaio/toutago-ritual-grove/internal/cli"
	"github.com/toutaio/toutago-ritual-grove/internal/registry"
	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

// Flags represents command line flags
type Flags struct {
	JSON bool
	Path string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "version":
		runVersionCommand()
	case "list":
		flags := parseFlags(os.Args[1:])
		var paths []string
		if flags.Path != "" {
			paths = []string{flags.Path}
		}

		var err error
		if flags.JSON {
			err = runListCommandJSON(paths)
		} else {
			err = runListCommand(paths)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ritual create <ritual-name> [project-path] [--yes] [--dry-run]")
			os.Exit(1)
		}

		ritualName := os.Args[2]
		projectPath := "."
		if len(os.Args) > 3 && !strings.HasPrefix(os.Args[3], "--") {
			projectPath = os.Args[3]
		}

		dryRun := false
		useDefaults := false
		for _, arg := range os.Args[3:] {
			if arg == "--dry-run" {
				dryRun = true
			}
			if arg == "--yes" {
				useDefaults = true
			}
		}

		if err := runCreateCommand(ritualName, projectPath, dryRun, useDefaults); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "mixin":
		fmt.Println("Managing mixins...")
		fmt.Println("(Not implemented yet)")
	case "deploy":
		fmt.Println("Deploying project...")
		fmt.Println("(Not implemented yet)")
	case "clean":
		if err := runCleanCommand(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Ritual Grove - Application Recipe System for Toutā")
	fmt.Printf("Version: %s\n\n", ritual.Version)
	fmt.Println("Usage: ritual <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list [--json] [--path <dir>]   List available rituals")
	fmt.Println("  create <name>                   Create project from ritual")
	fmt.Println("  clean [--all|--embedded]        Clear ritual cache")
	fmt.Println("  mixin add <name>                Add mixin to project")
	fmt.Println("  mixin list                      List available mixins")
	fmt.Println("  deploy                          Deploy project")
	fmt.Println("  update                          Update project to newer ritual version")
	fmt.Println("  version                         Show version information")
	fmt.Println("  help                            Show this help message")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --json                          Output in JSON format")
	fmt.Println("  --path <dir>                    Custom ritual search path")
	fmt.Println()
	fmt.Println("For more information, see:")
	fmt.Println("  https://github.com/toutaio/toutago-ritual-grove")
}

// parseFlags parses command line flags
func parseFlags(args []string) Flags {
	flags := Flags{}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			flags.JSON = true
		case "--path":
			if i+1 < len(args) {
				flags.Path = args[i+1]
				i++
			}
		}
	}

	return flags
}

// runVersionCommand displays version information
func runVersionCommand() {
	fmt.Printf("Ritual Grove v%s\n", ritual.Version)
	fmt.Println("Application Recipe System for Toutā Framework")
}

// runListCommand lists available rituals
func runListCommand(customPaths []string) error {
	reg := registry.NewRegistry()

	// Add custom paths if provided
	for _, path := range customPaths {
		reg.AddSearchPath(path)
	}

	// Scan for rituals
	if err := reg.Scan(); err != nil {
		return fmt.Errorf("failed to scan rituals: %w", err)
	}

	// Get all rituals
	rituals := reg.List()

	if len(rituals) == 0 {
		fmt.Println("No rituals found")
		return nil
	}

	fmt.Printf("Available Rituals (%d found):\n\n", len(rituals))

	for _, meta := range rituals {
		fmt.Printf("  %s (v%s)\n", meta.Name, meta.Version)
		if meta.Description != "" {
			fmt.Printf("    %s\n", meta.Description)
		}
		if len(meta.Tags) > 0 {
			fmt.Printf("    Tags: %v\n", meta.Tags)
		}
		if meta.Author != "" {
			fmt.Printf("    Author: %s\n", meta.Author)
		}
		fmt.Printf("    Source: %s\n", meta.Source)
		fmt.Println()
	}

	return nil
}

// runListCommandJSON lists rituals in JSON format
func runListCommandJSON(customPaths []string) error {
	reg := registry.NewRegistry()

	for _, path := range customPaths {
		reg.AddSearchPath(path)
	}

	if err := reg.Scan(); err != nil {
		return fmt.Errorf("failed to scan rituals: %w", err)
	}

	rituals := reg.List()

	// Convert to JSON-friendly format
	type RitualJSON struct {
		Name        string   `json:"name"`
		Version     string   `json:"version"`
		Description string   `json:"description"`
		Author      string   `json:"author,omitempty"`
		Tags        []string `json:"tags,omitempty"`
		Source      string   `json:"source"`
	}

	jsonRituals := make([]RitualJSON, len(rituals))
	for i, meta := range rituals {
		jsonRituals[i] = RitualJSON{
			Name:        meta.Name,
			Version:     meta.Version,
			Description: meta.Description,
			Author:      meta.Author,
			Tags:        meta.Tags,
			Source:      string(meta.Source),
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonRituals)
}

// runCreateCommand creates a project from a ritual
func runCreateCommand(ritualName, projectPath string, dryRun, useDefaults bool) error {
	// Find ritual
	reg := registry.NewRegistry()
	if err := reg.Scan(); err != nil {
		return fmt.Errorf("failed to scan rituals: %w", err)
	}

	meta, err := reg.Get(ritualName)
	if err != nil {
		return fmt.Errorf("ritual not found: %s", ritualName)
	}

	fmt.Printf("Using ritual: %s v%s\n", meta.Name, meta.Version)
	if meta.Description != "" {
		fmt.Printf("  %s\n", meta.Description)
	}
	fmt.Println()

	// Load ritual
	var answers map[string]interface{}
	if useDefaults {
		// Use default answers
		loader := ritual.NewLoader(meta.Path)
		manifest, err := loader.Load(meta.Path)
		if err != nil {
			return fmt.Errorf("failed to load ritual: %w", err)
		}

		answers = make(map[string]interface{})
		for _, q := range manifest.Questions {
			if q.Default != nil {
				answers[q.Name] = q.Default
			}
		}
	}

	// Execute workflow
	return cli.Execute(meta.Path, projectPath, answers, dryRun)
}

// runCleanCommand clears the ritual cache
func runCleanCommand(args []string) error {
reg := registry.NewRegistry()

clearAll := false
clearEmbedded := false

for _, arg := range args {
switch arg {
case "--all":
clearAll = true
case "--embedded":
clearEmbedded = true
}
}

// Default: clear embedded cache only
if !clearAll && !clearEmbedded {
clearEmbedded = true
}

if clearAll {
fmt.Printf("Clearing all cached rituals at: %s\n", reg.GetCachePath())

// Show cache size before clearing
if size, err := reg.GetCacheSize(); err == nil {
fmt.Printf("Cache size: %.2f MB\n", float64(size)/(1024*1024))
}

if err := reg.ClearCache(); err != nil {
return fmt.Errorf("failed to clear cache: %w", err)
}

fmt.Println("✓ All cache cleared successfully")
} else if clearEmbedded {
fmt.Println("Clearing embedded ritual cache...")

if err := reg.ClearEmbeddedCache(); err != nil {
return fmt.Errorf("failed to clear embedded cache: %w", err)
}

fmt.Println("✓ Embedded ritual cache cleared successfully")
fmt.Println("  Next scan will re-extract embedded rituals from the binary")
}

return nil
}

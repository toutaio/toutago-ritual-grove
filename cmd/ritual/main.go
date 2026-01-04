// Package main provides the CLI entry point for Ritual Grove
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/toutaio/toutago-ritual-grove/internal/generator"
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
			fmt.Println("Usage: ritual create <project-name>")
			os.Exit(1)
		}
		fmt.Println("Creating project from ritual...")
		fmt.Println("(Interactive questionnaire not fully implemented yet)")
	case "mixin":
		fmt.Println("Managing mixins...")
		fmt.Println("(Not implemented yet)")
	case "deploy":
		fmt.Println("Deploying project...")
		fmt.Println("(Not implemented yet)")
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
func runCreateCommand(ritualPath, targetPath string, answers map[string]interface{}) error {
	// Load ritual
	loader := ritual.NewLoader(ritualPath)
	manifest, err := loader.Load(ritualPath)
	if err != nil {
		return fmt.Errorf("failed to load ritual: %w", err)
	}
	
	// Create scaffolder
	scaffolder := generator.NewProjectScaffolder()
	
	// Convert answers to Variables
	vars := generator.NewVariables()
	for key, value := range answers {
		vars.Set(key, value)
	}
	
	// Generate project
	if err := scaffolder.GenerateFromRitual(targetPath, ritualPath, manifest, vars); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}
	
	fmt.Printf("Project created successfully at: %s\n", targetPath)
	return nil
}

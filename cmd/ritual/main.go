// Package main provides the CLI entry point for Ritual Grove
package main

import (
	"fmt"
	"os"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "version":
		fmt.Printf("Ritual Grove v%s\n", ritual.Version)
	case "list":
		fmt.Println("Listing available rituals...")
		fmt.Println("(Not implemented yet)")
	case "create":
		fmt.Println("Creating project from ritual...")
		fmt.Println("(Not implemented yet)")
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
	fmt.Println("  list                  List available rituals")
	fmt.Println("  create <name>         Create project from ritual")
	fmt.Println("  mixin add <name>      Add mixin to project")
	fmt.Println("  mixin list            List available mixins")
	fmt.Println("  deploy                Deploy project")
	fmt.Println("  update                Update project to newer ritual version")
	fmt.Println("  version               Show version information")
	fmt.Println("  help                  Show this help message")
	fmt.Println()
	fmt.Println("For more information, see:")
	fmt.Println("  https://github.com/toutaio/toutago-ritual-grove")
}

package inertia_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks/inertia"
)

func TestSetupInertiaMiddleware(t *testing.T) {
	t.Run("adds inertia middleware to main.go", func(t *testing.T) {
		tmpDir := t.TempDir()
		mainFile := filepath.Join(tmpDir, "main.go")

		// Create a basic main.go
		content := `package main

import (
	"github.com/toutaio/toutago/cosan"
)

func main() {
	router := cosan.NewRouter()
	router.Run(":8080")
}
`
		err := os.WriteFile(mainFile, []byte(content), 0644)
		require.NoError(t, err)

		// Execute task
		task := &inertia.SetupInertiaMiddlewareTask{ProjectDir: tmpDir}
		taskCtx := tasks.NewTaskContext()
		taskCtx.SetWorkingDir(tmpDir)
		
		err = task.Execute(context.Background(), taskCtx)
		require.NoError(t, err)

		// Verify middleware was added
		modified, err := os.ReadFile(mainFile)
		require.NoError(t, err)
		assert.Contains(t, string(modified), "github.com/toutaio/toutago-inertia")
		assert.Contains(t, string(modified), "inertia.NewMiddleware")
		assert.Contains(t, string(modified), "router.Use(")
	})

	t.Run("fails if main.go not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		task := &inertia.SetupInertiaMiddlewareTask{ProjectDir: tmpDir}
		taskCtx := tasks.NewTaskContext()
		
		err := task.Execute(context.Background(), taskCtx)
		assert.Error(t, err)
	})
}

func TestAddInertiaHandlers(t *testing.T) {
	t.Run("generates inertia handler file", func(t *testing.T) {
		tmpDir := t.TempDir()
		handlersDir := filepath.Join(tmpDir, "internal", "handlers")

		task := &inertia.AddInertiaHandlersTask{
			ProjectDir: tmpDir,
			Resource:   "posts",
		}
		taskCtx := tasks.NewTaskContext()
		taskCtx.SetWorkingDir(tmpDir)
		taskCtx.Set("resource", "posts")
		
		err := task.Execute(context.Background(), taskCtx)
		require.NoError(t, err)

		// Verify handler file created
		handlerFile := filepath.Join(handlersDir, "posts_handler.go")
		assert.FileExists(t, handlerFile)

		content, err := os.ReadFile(handlerFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "package handlers")
		assert.Contains(t, string(content), "ctx.Inertia().Render")
		assert.Contains(t, string(content), "Index")
		assert.Contains(t, string(content), "Show")
		assert.Contains(t, string(content), "Create")
		assert.Contains(t, string(content), "Update")
		assert.Contains(t, string(content), "Delete")
	})
}

func TestAddSharedData(t *testing.T) {
	t.Run("adds shared data configuration", func(t *testing.T) {
		tmpDir := t.TempDir()

		task := &inertia.AddSharedDataTask{
			ProjectDir: tmpDir,
			SharedData: []string{"user", "flash"},
		}
		taskCtx := tasks.NewTaskContext()
		taskCtx.SetWorkingDir(tmpDir)
		taskCtx.Set("shared_data", []string{"user", "flash"})
		
		err := task.Execute(context.Background(), taskCtx)
		require.NoError(t, err)

		// Verify config file created
		configFile := filepath.Join(tmpDir, "config", "inertia.go")
		assert.FileExists(t, configFile)

		content, err := os.ReadFile(configFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "package config")
		assert.Contains(t, string(content), "SharedData")
		assert.Contains(t, string(content), "func GetUser")
		assert.Contains(t, string(content), "func GetFlash")
	})
}

func TestGenerateTypeScriptTypes(t *testing.T) {
	t.Run("generates TypeScript types from Go structs", func(t *testing.T) {
		tmpDir := t.TempDir()
		modelsDir := filepath.Join(tmpDir, "internal", "models")
		err := os.MkdirAll(modelsDir, 0755)
		require.NoError(t, err)

		// Create a sample model
		modelContent := `package models

type Post struct {
	ID        int64  ` + "`json:\"id\"`" + `
	Title     string ` + "`json:\"title\"`" + `
	Content   string ` + "`json:\"content\"`" + `
	Published bool   ` + "`json:\"published\"`" + `
}
`
		err = os.WriteFile(filepath.Join(modelsDir, "post.go"), []byte(modelContent), 0644)
		require.NoError(t, err)

		typesDir := filepath.Join(tmpDir, "frontend", "types")
		err = os.MkdirAll(typesDir, 0755)
		require.NoError(t, err)

		task := &inertia.GenerateTypeScriptTypesTask{
			ProjectDir: tmpDir,
			ModelsDir:  modelsDir,
			OutputDir:  typesDir,
		}
		taskCtx := tasks.NewTaskContext()
		taskCtx.SetWorkingDir(tmpDir)
		taskCtx.Set("models_dir", modelsDir)
		taskCtx.Set("output_dir", typesDir)
		
		err = task.Execute(context.Background(), taskCtx)
		require.NoError(t, err)

		// Verify TypeScript types generated
		typesFile := filepath.Join(typesDir, "models.d.ts")
		assert.FileExists(t, typesFile)

		content, err := os.ReadFile(typesFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "export interface Post")
		assert.Contains(t, string(content), "id: number")
		assert.Contains(t, string(content), "title: string")
		assert.Contains(t, string(content), "content: string")
		assert.Contains(t, string(content), "published: boolean")
	})
}

func TestUpdateRoutesForInertia(t *testing.T) {
	t.Run("updates route definitions for Inertia", func(t *testing.T) {
		tmpDir := t.TempDir()
		routesFile := filepath.Join(tmpDir, "internal", "routes", "routes.go")
		err := os.MkdirAll(filepath.Dir(routesFile), 0755)
		require.NoError(t, err)

		// Create basic routes file
		content := `package routes

import "github.com/toutaio/toutago/cosan"

func Setup(router *cosan.Router) {
	// Existing routes
}
`
		err = os.WriteFile(routesFile, []byte(content), 0644)
		require.NoError(t, err)

		task := &inertia.UpdateRoutesForInertiaTask{
			ProjectDir: tmpDir,
			Resource:   "posts",
		}
		taskCtx := tasks.NewTaskContext()
		taskCtx.SetWorkingDir(tmpDir)
		taskCtx.Set("resource", "posts")
		
		err = task.Execute(context.Background(), taskCtx)
		require.NoError(t, err)

		// Verify routes updated
		modified, err := os.ReadFile(routesFile)
		require.NoError(t, err)
		assert.Contains(t, string(modified), "/posts")
		assert.Contains(t, string(modified), "GET")
		assert.Contains(t, string(modified), "POST")
		assert.Contains(t, string(modified), "PUT")
		assert.Contains(t, string(modified), "DELETE")
	})
}

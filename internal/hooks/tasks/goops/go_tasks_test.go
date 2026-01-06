package goops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

func TestGoRunTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple Go program.
	goMod := `module example.com/test

go 1.21
`
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello from go run")
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0600); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0600); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	task := &GoRunTask{File: "main.go", Dir: tmpDir}

	if err := task.Validate(); err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	ctx := tasks.NewTaskContext()
	ctx.SetWorkingDir(tmpDir)

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestGoRunTask_WithArgs(t *testing.T) {
	tmpDir := t.TempDir()

	goMod := `module example.com/test

go 1.21
`
	mainGo := `package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Println("Got args:", os.Args[1:])
	}
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0600); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0600); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	task := &GoRunTask{
		File: "main.go",
		Args: []string{"arg1", "arg2"},
		Dir:  tmpDir,
	}

	ctx := tasks.NewTaskContext()
	ctx.SetWorkingDir(tmpDir)

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestGoRunTask_WithEnv(t *testing.T) {
	tmpDir := t.TempDir()

	goMod := `module example.com/test

go 1.21
`
	mainGo := `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("TEST_VAR:", os.Getenv("TEST_VAR"))
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0600); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0600); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	task := &GoRunTask{
		File: "main.go",
		Dir:  tmpDir,
		Env:  map[string]string{"TEST_VAR": "test_value"},
	}

	ctx := tasks.NewTaskContext()
	ctx.SetWorkingDir(tmpDir)

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestGoRunTask_Validation(t *testing.T) {
	tests := []struct {
		name      string
		task      *GoRunTask
		wantError bool
	}{
		{
			name:      "valid task",
			task:      &GoRunTask{File: "main.go"},
			wantError: false,
		},
		{
			name:      "missing file",
			task:      &GoRunTask{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestGoModDownloadTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple go.mod.
	goMod := `module example.com/test

go 1.21

require github.com/stretchr/testify v1.8.4
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	task := &GoModDownloadTask{Dir: tmpDir}

	ctx := tasks.NewTaskContext()
	ctx.SetWorkingDir(tmpDir)

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Logf("Note: go mod download may fail in test environment: %v", err)
		// Don't fail the test as this requires network access.
	}
}

func TestGoBuildTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple Go program.
	goMod := `module example.com/test

go 1.21
`
	mainGo := `package main

func main() {
	println("Hello")
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "testbin")
	task := &GoBuildTask{
		Dir:    tmpDir,
		Output: outputPath,
	}

	ctx := tasks.NewTaskContext()
	ctx.SetWorkingDir(tmpDir)

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify binary exists.
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("Binary not created: %v", err)
	}
}

func TestGoTestTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple test.
	goMod := `module example.com/test

go 1.21
`
	testGo := `package main

import "testing"

func TestExample(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math is broken")
	}
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "example_test.go"), []byte(testGo), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	task := &GoTestTask{Dir: tmpDir}

	ctx := tasks.NewTaskContext()
	ctx.SetWorkingDir(tmpDir)

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestGoFmtTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod first.
	goMod := `module example.com/test

go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create an unformatted Go file.
	badFormat := `package main

func main(  ){
println(  "hello"  )
}
`
	filePath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(filePath, []byte(badFormat), 0644); err != nil {
		t.Fatalf("Failed to create Go file: %v", err)
	}

	task := &GoFmtTask{Dir: tmpDir}

	ctx := tasks.NewTaskContext()
	ctx.SetWorkingDir(tmpDir)

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify file was formatted.
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read formatted file: %v", err)
	}

	// After formatting, should have proper spacing.
	formatted := string(content)
	if formatted == badFormat {
		t.Error("File was not formatted")
	}
}

func TestExecGoTask(t *testing.T) {
	task := &ExecGoTask{
		Command: []string{"version"},
	}

	ctx := tasks.NewTaskContext()

	err := task.Execute(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestExecGoTaskValidation(t *testing.T) {
	// Test empty command.
	task := &ExecGoTask{}
	if err := task.Validate(); err == nil {
		t.Error("Expected validation error for empty command")
	}
}

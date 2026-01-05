package deployment

import (
	"testing"
)

func TestDiffGenerator_DetectChanges(t *testing.T) {
	generator := NewDiffGenerator()

	currentFiles := map[string]string{
		"main.go":       "package main\n// v1.0",
		"handler.go":    "package handler\n// v1.0",
		"config.yaml":   "port: 8080",
		"deprecated.go": "package old",
	}

	newFiles := map[string]string{
		"main.go":    "package main\n// v1.1",    // modified
		"handler.go": "package handler\n// v1.0", // unchanged
		"config.yaml": "port: 8080\nenv: prod",  // modified
		"new_file.go": "package new",             // added
		// deprecated.go removed
	}

	diff := generator.GenerateDiff(currentFiles, newFiles)

	if len(diff.Added) != 1 {
		t.Errorf("Expected 1 added file, got %d", len(diff.Added))
	}
	if diff.Added[0] != "new_file.go" {
		t.Errorf("Expected new_file.go to be added, got %s", diff.Added[0])
	}

	if len(diff.Modified) != 2 {
		t.Errorf("Expected 2 modified files, got %d", len(diff.Modified))
	}

	if len(diff.Deleted) != 1 {
		t.Errorf("Expected 1 deleted file, got %d", len(diff.Deleted))
	}
	if diff.Deleted[0] != "deprecated.go" {
		t.Errorf("Expected deprecated.go to be deleted, got %s", diff.Deleted[0])
	}

	if len(diff.Unchanged) != 1 {
		t.Errorf("Expected 1 unchanged file, got %d", len(diff.Unchanged))
	}
}

func TestDiffGenerator_WithProtectedFiles(t *testing.T) {
	generator := NewDiffGenerator()

	currentFiles := map[string]string{
		"main.go":       "package main\n// user modified",
		"generated.go":  "package gen\n// generated",
	}

	newFiles := map[string]string{
		"main.go":      "package main\n// ritual version",
		"generated.go": "package gen\n// updated gen",
	}

	protectedFiles := []string{"main.go"}

	diff := generator.GenerateDiffWithProtected(currentFiles, newFiles, protectedFiles)

	// main.go should be in conflicts because it's protected and modified
	if len(diff.Conflicts) != 1 {
		t.Errorf("Expected 1 conflict, got %d", len(diff.Conflicts))
	}
	if diff.Conflicts[0] != "main.go" {
		t.Errorf("Expected main.go in conflicts, got %s", diff.Conflicts[0])
	}

	// generated.go should be in modified (not protected)
	hasGeneratedInModified := false
	for _, f := range diff.Modified {
		if f == "generated.go" {
			hasGeneratedInModified = true
			break
		}
	}
	if !hasGeneratedInModified {
		t.Error("Expected generated.go in modified files")
	}
}

func TestDiffGenerator_Summary(t *testing.T) {
	diff := &FileDiff{
		Added:     []string{"a.go", "b.go"},
		Modified:  []string{"c.go"},
		Deleted:   []string{"d.go"},
		Unchanged: []string{"e.go", "f.go"},
		Conflicts: []string{"g.go"},
	}

	summary := diff.Summary()

	if summary.TotalChanges != 5 {
		t.Errorf("TotalChanges = %d, want 5", summary.TotalChanges)
	}
	if summary.AddedCount != 2 {
		t.Errorf("AddedCount = %d, want 2", summary.AddedCount)
	}
	if summary.ModifiedCount != 1 {
		t.Errorf("ModifiedCount = %d, want 1", summary.ModifiedCount)
	}
	if summary.DeletedCount != 1 {
		t.Errorf("DeletedCount = %d, want 1", summary.DeletedCount)
	}
	if summary.ConflictCount != 1 {
		t.Errorf("ConflictCount = %d, want 1", summary.ConflictCount)
	}
}

func TestDiffGenerator_EmptyDiff(t *testing.T) {
	generator := NewDiffGenerator()

	files := map[string]string{
		"main.go": "package main",
	}

	diff := generator.GenerateDiff(files, files)

	if len(diff.Added) != 0 {
		t.Errorf("Expected no added files, got %d", len(diff.Added))
	}
	if len(diff.Modified) != 0 {
		t.Errorf("Expected no modified files, got %d", len(diff.Modified))
	}
	if len(diff.Deleted) != 0 {
		t.Errorf("Expected no deleted files, got %d", len(diff.Deleted))
	}
	if len(diff.Unchanged) != 1 {
		t.Errorf("Expected 1 unchanged file, got %d", len(diff.Unchanged))
	}
}

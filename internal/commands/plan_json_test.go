package commands

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/toutaio/toutago-ritual-grove/internal/deployment"
	"github.com/toutaio/toutago-ritual-grove/internal/storage"
)

// TestPlanJSON_BasicOutput tests JSON output structure
func TestPlanJSON_BasicOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create state
	state := &storage.State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
		InstalledAt:   time.Now(),
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create a mock plan
	plan := &deployment.DeploymentPlan{
		CurrentVersion: "1.0.0",
		TargetVersion:  "1.1.0",
		FilesAdded:     []string{"new.go"},
		FilesModified:  []string{"main.go"},
		FilesDeleted:   []string{"old.go"},
		MigrationsToRun: []string{"1.0.0 → 1.1.0: Add table"},
		Conflicts: []deployment.Conflict{
			{File: "config.yaml", Reason: "modified locally and in update"},
		},
		EstimatedDuration: 120 * time.Second,
	}

	// Capture JSON output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputPlanJSON(plan)
	if err != nil {
		t.Fatalf("Failed to output JSON: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf [4096]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, output)
	}

	// Verify structure
	if result["current_version"] != "1.0.0" {
		t.Error("Missing or incorrect current_version")
	}

	if result["target_version"] != "1.1.0" {
		t.Error("Missing or incorrect target_version")
	}

	// Verify files object
	files, ok := result["files"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing files object")
	}

	toAdd, ok := files["to_add"].([]interface{})
	if !ok || len(toAdd) != 1 {
		t.Error("Incorrect to_add files")
	}

	// Verify migrations
	migrations, ok := result["migrations"].([]interface{})
	if !ok || len(migrations) != 1 {
		t.Error("Incorrect migrations")
	}

	// Verify conflicts
	conflicts, ok := result["conflicts"].([]interface{})
	if !ok || len(conflicts) != 1 {
		t.Error("Incorrect conflicts")
	}

	// Verify estimated duration
	duration, ok := result["estimated_duration_seconds"].(float64)
	if !ok || duration != 120 {
		t.Errorf("Incorrect estimated_duration, got %v", result["estimated_duration_seconds"])
	}

	// Verify requires manual intervention
	requiresManual, ok := result["requires_manual_intervention"].(bool)
	if !ok || !requiresManual {
		t.Error("Should require manual intervention due to conflicts")
	}
}

// TestPlanJSON_NoConflicts tests JSON output without conflicts
func TestPlanJSON_NoConflicts(t *testing.T) {
	plan := &deployment.DeploymentPlan{
		CurrentVersion:    "1.0.0",
		TargetVersion:     "1.1.0",
		FilesAdded:        []string{},
		FilesModified:     []string{},
		FilesDeleted:      []string{},
		MigrationsToRun:   []string{},
		Conflicts:         []deployment.Conflict{},
		EstimatedDuration: 30 * time.Second,
	}

	// Capture JSON output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputPlanJSON(plan)
	if err != nil {
		t.Fatalf("Failed to output JSON: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf [4096]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Should not require manual intervention
	requiresManual, ok := result["requires_manual_intervention"].(bool)
	if !ok || requiresManual {
		t.Error("Should not require manual intervention without conflicts")
	}

	// Verify empty arrays
	conflicts, ok := result["conflicts"].([]interface{})
	if !ok || len(conflicts) != 0 {
		t.Error("Should have empty conflicts array")
	}
}

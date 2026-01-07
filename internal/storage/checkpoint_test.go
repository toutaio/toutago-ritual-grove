package storage

import (
	"testing"
	"time"
)

// TestCheckpoint_Create tests creating a state checkpoint
func TestCheckpoint_Create(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial state
	state := &State{
		RitualName:    "test-ritual",
		RitualVersion: "1.0.0",
		InstalledAt:   time.Now(),
	}

	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save initial state: %v", err)
	}

	// Create checkpoint manager
	cm := NewCheckpointManager(tmpDir)

	// Create checkpoint
	checkpointID, err := cm.CreateCheckpoint("before-update", state)
	if err != nil {
		t.Fatalf("Failed to create checkpoint: %v", err)
	}

	if checkpointID == "" {
		t.Error("Checkpoint ID should not be empty")
	}

	// Verify checkpoint exists
	checkpoints, err := cm.ListCheckpoints()
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 1 {
		t.Errorf("Expected 1 checkpoint, got %d", len(checkpoints))
	}

	if checkpoints[0].ID != checkpointID {
		t.Errorf("Expected checkpoint ID %s, got %s", checkpointID, checkpoints[0].ID)
	}
}

// TestCheckpoint_Restore tests restoring from a checkpoint
func TestCheckpoint_Restore(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial state
	state := &State{
		RitualName:    "test-ritual",
		RitualVersion: "1.0.0",
	}
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create checkpoint
	cm := NewCheckpointManager(tmpDir)
	checkpointID, err := cm.CreateCheckpoint("v1.0.0", state)
	if err != nil {
		t.Fatalf("Failed to create checkpoint: %v", err)
	}

	// Modify state
	state.RitualVersion = "2.0.0"
	if err := state.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save modified state: %v", err)
	}

	// Restore from checkpoint
	if err := cm.RestoreCheckpoint(checkpointID); err != nil {
		t.Fatalf("Failed to restore checkpoint: %v", err)
	}

	// Verify restoration
	restored, err := LoadState(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load restored state: %v", err)
	}

	if restored.RitualVersion != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", restored.RitualVersion)
	}
}

// TestCheckpoint_List tests listing checkpoints
func TestCheckpoint_List(t *testing.T) {
	tmpDir := t.TempDir()

	state := &State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
	}

	cm := NewCheckpointManager(tmpDir)

	// Create multiple checkpoints
	_, err := cm.CreateCheckpoint("checkpoint-1", state)
	if err != nil {
		t.Fatalf("Failed to create checkpoint 1: %v", err)
	}

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	state.RitualVersion = "1.1.0"
	_, err = cm.CreateCheckpoint("checkpoint-2", state)
	if err != nil {
		t.Fatalf("Failed to create checkpoint 2: %v", err)
	}

	// List checkpoints
	checkpoints, err := cm.ListCheckpoints()
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints, got %d", len(checkpoints))
	}

	// Should be sorted by timestamp (newest first)
	if checkpoints[0].Label != "checkpoint-2" {
		t.Errorf("Expected newest checkpoint first, got %s", checkpoints[0].Label)
	}
}

// TestCheckpoint_Delete tests deleting a checkpoint
func TestCheckpoint_Delete(t *testing.T) {
	tmpDir := t.TempDir()

	state := &State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
	}

	cm := NewCheckpointManager(tmpDir)
	checkpointID, err := cm.CreateCheckpoint("to-delete", state)
	if err != nil {
		t.Fatalf("Failed to create checkpoint: %v", err)
	}

	// Delete checkpoint
	if err := cm.DeleteCheckpoint(checkpointID); err != nil {
		t.Fatalf("Failed to delete checkpoint: %v", err)
	}

	// Verify deletion
	checkpoints, err := cm.ListCheckpoints()
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 0 {
		t.Errorf("Expected 0 checkpoints after deletion, got %d", len(checkpoints))
	}
}

// TestCheckpoint_CleanOld tests cleaning old checkpoints
func TestCheckpoint_CleanOld(t *testing.T) {
	tmpDir := t.TempDir()

	state := &State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
	}

	cm := NewCheckpointManager(tmpDir)

	// Create 5 checkpoints
	for i := 0; i < 5; i++ {
		_, err := cm.CreateCheckpoint("checkpoint", state)
		if err != nil {
			t.Fatalf("Failed to create checkpoint %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Keep only 3 newest
	if err := cm.CleanOldCheckpoints(3); err != nil {
		t.Fatalf("Failed to clean old checkpoints: %v", err)
	}

	// Verify only 3 remain
	checkpoints, err := cm.ListCheckpoints()
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 3 {
		t.Errorf("Expected 3 checkpoints, got %d", len(checkpoints))
	}
}

// TestCheckpoint_GetByLabel tests finding checkpoint by label
func TestCheckpoint_GetByLabel(t *testing.T) {
	tmpDir := t.TempDir()

	state := &State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
	}

	cm := NewCheckpointManager(tmpDir)
	_, err := cm.CreateCheckpoint("my-label", state)
	if err != nil {
		t.Fatalf("Failed to create checkpoint: %v", err)
	}

	// Find by label
	checkpoint, err := cm.GetCheckpointByLabel("my-label")
	if err != nil {
		t.Fatalf("Failed to get checkpoint by label: %v", err)
	}

	if checkpoint.Label != "my-label" {
		t.Errorf("Expected label 'my-label', got '%s'", checkpoint.Label)
	}
}

// TestCheckpoint_AutoClean tests automatic cleanup on create
func TestCheckpoint_AutoClean(t *testing.T) {
	tmpDir := t.TempDir()

	state := &State{
		RitualName:    "test",
		RitualVersion: "1.0.0",
	}

	cm := NewCheckpointManager(tmpDir)
	cm.SetMaxCheckpoints(3) // Limit to 3 checkpoints

	// Create 5 checkpoints
	for i := 0; i < 5; i++ {
		_, err := cm.CreateCheckpoint("checkpoint", state)
		if err != nil {
			t.Fatalf("Failed to create checkpoint %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Should automatically keep only 3
	checkpoints, err := cm.ListCheckpoints()
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 3 {
		t.Errorf("Expected auto-cleanup to keep 3 checkpoints, got %d", len(checkpoints))
	}
}

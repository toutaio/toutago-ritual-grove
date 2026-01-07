package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Checkpoint represents a saved state at a point in time
type Checkpoint struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	Timestamp time.Time `json:"timestamp"`
	State     *State    `json:"state"`
}

// CheckpointManager manages state checkpoints for point-in-time restoration
type CheckpointManager struct {
	projectPath    string
	maxCheckpoints int
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(projectPath string) *CheckpointManager {
	return &CheckpointManager{
		projectPath:    projectPath,
		maxCheckpoints: 10, // Default to keeping 10 checkpoints
	}
}

// SetMaxCheckpoints sets the maximum number of checkpoints to retain
func (cm *CheckpointManager) SetMaxCheckpoints(max int) {
	cm.maxCheckpoints = max
}

// CreateCheckpoint creates a new checkpoint with the given label
func (cm *CheckpointManager) CreateCheckpoint(label string, state *State) (string, error) {
	// Use nanoseconds for uniqueness
	checkpointID := fmt.Sprintf("%d-%s", time.Now().UnixNano(), label)

	checkpoint := Checkpoint{
		ID:        checkpointID,
		Label:     label,
		Timestamp: time.Now(),
		State:     state,
	}

	// Save checkpoint
	checkpointDir := filepath.Join(cm.projectPath, ".ritual", "checkpoints")
	if err := os.MkdirAll(checkpointDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create checkpoints directory: %w", err)
	}

	checkpointFile := filepath.Join(checkpointDir, checkpointID+".json")
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	if err := os.WriteFile(checkpointFile, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write checkpoint: %w", err)
	}

	// Auto-cleanup old checkpoints
	if err := cm.CleanOldCheckpoints(cm.maxCheckpoints); err != nil {
		// Log but don't fail on cleanup error
		fmt.Printf("Warning: failed to clean old checkpoints: %v\n", err)
	}

	return checkpointID, nil
}

// RestoreCheckpoint restores the project state from a checkpoint
func (cm *CheckpointManager) RestoreCheckpoint(checkpointID string) error {
	checkpointFile := filepath.Join(cm.projectPath, ".ritual", "checkpoints", checkpointID+".json")

	data, err := os.ReadFile(checkpointFile)
	if err != nil {
		return fmt.Errorf("failed to read checkpoint: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	// Restore state
	if err := checkpoint.State.Save(cm.projectPath); err != nil {
		return fmt.Errorf("failed to restore state: %w", err)
	}

	return nil
}

// ListCheckpoints returns all checkpoints sorted by timestamp (newest first)
func (cm *CheckpointManager) ListCheckpoints() ([]Checkpoint, error) {
	checkpointDir := filepath.Join(cm.projectPath, ".ritual", "checkpoints")

	// Check if directory exists
	if _, err := os.Stat(checkpointDir); os.IsNotExist(err) {
		return []Checkpoint{}, nil
	}

	entries, err := os.ReadDir(checkpointDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoints directory: %w", err)
	}

	var checkpoints []Checkpoint
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		checkpointFile := filepath.Join(checkpointDir, entry.Name())
		data, err := os.ReadFile(checkpointFile)
		if err != nil {
			continue // Skip unreadable files
		}

		var checkpoint Checkpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			continue // Skip invalid checkpoints
		}

		checkpoints = append(checkpoints, checkpoint)
	}

	// Sort by timestamp, newest first
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].Timestamp.After(checkpoints[j].Timestamp)
	})

	return checkpoints, nil
}

// DeleteCheckpoint deletes a specific checkpoint
func (cm *CheckpointManager) DeleteCheckpoint(checkpointID string) error {
	checkpointFile := filepath.Join(cm.projectPath, ".ritual", "checkpoints", checkpointID+".json")

	if err := os.Remove(checkpointFile); err != nil {
		return fmt.Errorf("failed to delete checkpoint: %w", err)
	}

	return nil
}

// CleanOldCheckpoints keeps only the N most recent checkpoints
func (cm *CheckpointManager) CleanOldCheckpoints(keepCount int) error {
	checkpoints, err := cm.ListCheckpoints()
	if err != nil {
		return err
	}

	// Already sorted newest first, so delete from keepCount onward
	for i := keepCount; i < len(checkpoints); i++ {
		if err := cm.DeleteCheckpoint(checkpoints[i].ID); err != nil {
			return err
		}
	}

	return nil
}

// GetCheckpointByLabel finds a checkpoint by its label (returns most recent if multiple)
func (cm *CheckpointManager) GetCheckpointByLabel(label string) (*Checkpoint, error) {
	checkpoints, err := cm.ListCheckpoints()
	if err != nil {
		return nil, err
	}

	for _, checkpoint := range checkpoints {
		if checkpoint.Label == label {
			return &checkpoint, nil
		}
	}

	return nil, fmt.Errorf("checkpoint with label '%s' not found", label)
}

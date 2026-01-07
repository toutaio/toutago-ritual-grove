package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const maxHistoryEntries = 100

// DeploymentRecord represents a single deployment attempt.
type DeploymentRecord struct {
	Timestamp   time.Time `yaml:"timestamp"`
	FromVersion string    `yaml:"from_version"`
	ToVersion   string    `yaml:"to_version"`
	Status      string    `yaml:"status"` // "success", "failure", "rollback"
	Message     string    `yaml:"message,omitempty"`
	Errors      []string  `yaml:"errors,omitempty"`
	Warnings    []string  `yaml:"warnings,omitempty"`
	Duration    string    `yaml:"duration,omitempty"`
}

// DeploymentHistory tracks all deployment attempts for a project.
type DeploymentHistory struct {
	Deployments []DeploymentRecord `yaml:"deployments"`
}

// AddDeployment adds a new deployment record to the history.
// Automatically limits history to the last maxHistoryEntries records.
func (h *DeploymentHistory) AddDeployment(record DeploymentRecord) {
	h.Deployments = append(h.Deployments, record)

	// Keep only the last maxHistoryEntries
	if len(h.Deployments) > maxHistoryEntries {
		h.Deployments = h.Deployments[len(h.Deployments)-maxHistoryEntries:]
	}
}

// GetLatestSuccessful returns the most recent successful deployment, or nil if none found.
func (h *DeploymentHistory) GetLatestSuccessful() *DeploymentRecord {
	for i := len(h.Deployments) - 1; i >= 0; i-- {
		if h.Deployments[i].Status == "success" {
			return &h.Deployments[i]
		}
	}
	return nil
}

// GetFailures returns all failed deployment records.
func (h *DeploymentHistory) GetFailures() []DeploymentRecord {
	var failures []DeploymentRecord
	for _, d := range h.Deployments {
		if d.Status == "failure" {
			failures = append(failures, d)
		}
	}
	return failures
}

// GetRollbacks returns all rollback deployment records.
func (h *DeploymentHistory) GetRollbacks() []DeploymentRecord {
	var rollbacks []DeploymentRecord
	for _, d := range h.Deployments {
		if d.Status == "rollback" {
			rollbacks = append(rollbacks, d)
		}
	}
	return rollbacks
}

// Save saves the deployment history to .ritual/history.yaml.
func (h *DeploymentHistory) Save(projectPath string) error {
	stateDir := filepath.Join(projectPath, ".ritual")
	if err := os.MkdirAll(stateDir, 0750); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	historyPath := filepath.Join(stateDir, "history.yaml")
	data, err := yaml.Marshal(h)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(historyPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// LoadDeploymentHistory loads the deployment history from .ritual/history.yaml.
func LoadDeploymentHistory(projectPath string) (*DeploymentHistory, error) {
	historyPath := filepath.Join(projectPath, ".ritual", "history.yaml")
	// #nosec G304 - historyPath is constructed from validated components
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var history DeploymentHistory
	if err := yaml.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return &history, nil
}

// LoadOrCreateHistory loads existing history or creates a new one.
func LoadOrCreateHistory(projectPath string) *DeploymentHistory {
	history, err := LoadDeploymentHistory(projectPath)
	if err != nil {
		// If history doesn't exist, create new
		return &DeploymentHistory{
			Deployments: []DeploymentRecord{},
		}
	}
	return history
}

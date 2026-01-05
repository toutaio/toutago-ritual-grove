package deployment

import (
	"testing"
	"time"

	"github.com/toutaio/toutago-ritual-grove/pkg/ritual"
)

func TestDeploymentPlanner_AnalyzeChanges(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		targetVersion  string
		currentRitual  *ritual.Manifest
		targetRitual   *ritual.Manifest
		wantStepCount  int
		wantConflicts  bool
	}{
		{
			name:           "simple version bump",
			currentVersion: "1.0.0",
			targetVersion:  "1.1.0",
			currentRitual: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
				Files: ritual.FilesSection{
					Templates: []ritual.FileMapping{
						{Source: "main.go.tmpl", Destination: "main.go"},
					},
				},
			},
			targetRitual: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test",
					Version: "1.1.0",
				},
				Files: ritual.FilesSection{
					Templates: []ritual.FileMapping{
						{Source: "main.go.tmpl", Destination: "main.go"},
						{Source: "config.go.tmpl", Destination: "config.go"},
					},
				},
			},
			wantStepCount: 2, // Add new file + update version
			wantConflicts: false,
		},
		{
			name:           "breaking change version",
			currentVersion: "1.0.0",
			targetVersion:  "2.0.0",
			currentRitual: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
			},
			targetRitual: &ritual.Manifest{
				Ritual: ritual.RitualMeta{
					Name:    "test",
					Version: "2.0.0",
				},
			},
			wantStepCount: 1,
			wantConflicts: true, // Major version bump
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planner := NewPlanner()
			plan, err := planner.AnalyzeChanges(tt.currentRitual, tt.targetRitual)
			if err != nil {
				t.Fatalf("AnalyzeChanges() error = %v", err)
			}

			if len(plan.Steps) < tt.wantStepCount {
				t.Errorf("AnalyzeChanges() got %d steps, want at least %d", len(plan.Steps), tt.wantStepCount)
			}

			if (len(plan.Conflicts) > 0) != tt.wantConflicts {
				t.Errorf("AnalyzeChanges() conflicts = %v, want %v", len(plan.Conflicts) > 0, tt.wantConflicts)
			}
		})
	}
}

func TestDeploymentPlan_EstimateDuration(t *testing.T) {
	plan := &DeploymentPlan{
		Steps: []DeploymentStep{
			{Type: StepBackup, Description: "Backup files"},
			{Type: StepMigration, Description: "Run migration"},
			{Type: StepUpdateFiles, Description: "Update 5 files"},
		},
	}

	duration := plan.EstimateDuration()
	if duration <= 0 {
		t.Errorf("EstimateDuration() = %v, want > 0", duration)
	}

	// Backup and migration should take longer
	minExpected := 5 * time.Second
	if duration < minExpected {
		t.Errorf("EstimateDuration() = %v, want >= %v", duration, minExpected)
	}
}

func TestDeploymentPlan_DetectConflicts(t *testing.T) {
	tests := []struct {
		name          string
		modifiedFiles []string
		targetFiles   []string
		wantConflicts int
	}{
		{
			name:          "no conflicts",
			modifiedFiles: []string{},
			targetFiles:   []string{"main.go", "config.go"},
			wantConflicts: 0,
		},
		{
			name:          "modified file conflict",
			modifiedFiles: []string{"main.go"},
			targetFiles:   []string{"main.go", "config.go"},
			wantConflicts: 1,
		},
		{
			name:          "multiple conflicts",
			modifiedFiles: []string{"main.go", "config.go"},
			targetFiles:   []string{"main.go", "config.go", "utils.go"},
			wantConflicts: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &DeploymentPlan{}
			conflicts := plan.detectConflicts(tt.modifiedFiles, tt.targetFiles)
			if len(conflicts) != tt.wantConflicts {
				t.Errorf("detectConflicts() = %d conflicts, want %d", len(conflicts), tt.wantConflicts)
			}
		})
	}
}

func TestDeploymentStep_String(t *testing.T) {
	tests := []struct {
		name string
		step DeploymentStep
		want string
	}{
		{
			name: "backup step",
			step: DeploymentStep{
				Type:        StepBackup,
				Description: "Backup project files",
			},
			want: "backup: Backup project files",
		},
		{
			name: "migration step",
			step: DeploymentStep{
				Type:        StepMigration,
				Description: "Run migration v1.1.0",
			},
			want: "migration: Run migration v1.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.step.String(); got != tt.want {
				t.Errorf("DeploymentStep.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlanner_GenerateReport(t *testing.T) {
	plan := &DeploymentPlan{
		CurrentVersion: "1.0.0",
		TargetVersion:  "1.1.0",
		Steps: []DeploymentStep{
			{Type: StepBackup, Description: "Backup files"},
			{Type: StepUpdateFiles, Description: "Update 3 files"},
		},
		Conflicts: []Conflict{
			{File: "main.go", Reason: "File was manually modified"},
		},
		EstimatedDuration: 10 * time.Second,
	}

	planner := NewPlanner()
	report := planner.GenerateReport(plan)

	if report == "" {
		t.Error("GenerateReport() returned empty report")
	}

	// Report should contain key information
	if !contains(report, "1.0.0") || !contains(report, "1.1.0") {
		t.Error("Report missing version information")
	}
	if !contains(report, "Backup") {
		t.Error("Report missing step information")
	}
	if !contains(report, "main.go") {
		t.Error("Report missing conflict information")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

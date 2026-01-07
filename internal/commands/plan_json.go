package commands

import (
	"encoding/json"
	"fmt"

	"github.com/toutaio/toutago-ritual-grove/internal/deployment"
)

func outputPlanJSON(plan *deployment.DeploymentPlan) error {
	// Convert conflicts to simple strings
	conflicts := make([]string, len(plan.Conflicts))
	for i, c := range plan.Conflicts {
		conflicts[i] = c.File
	}

	output := map[string]interface{}{
		"current_version": plan.CurrentVersion,
		"target_version":  plan.TargetVersion,
		"files": map[string]interface{}{
			"to_add":    plan.FilesAdded,
			"to_modify": plan.FilesModified,
			"to_delete": plan.FilesDeleted,
		},
		"migrations":                   plan.MigrationsToRun,
		"conflicts":                    conflicts,
		"estimated_duration_seconds":   int(plan.EstimatedDuration.Seconds()),
		"requires_manual_intervention": len(plan.Conflicts) > 0,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

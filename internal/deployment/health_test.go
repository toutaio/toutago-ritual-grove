package deployment

import (
	"context"
	"testing"
	"time"
)

func TestHealthChecker_DatabaseCheck(t *testing.T) {
	checker := NewHealthChecker()

	// Test with invalid config (should fail)
	dbCheck := DatabaseHealthCheck{
		DSN:     "invalid://connection",
		Driver:  "mysql",
		Timeout: time.Second,
	}

	result := checker.CheckDatabase(context.Background(), dbCheck)

	if result.Healthy {
		t.Error("Expected database check to fail with invalid DSN")
	}
	if result.Error == "" {
		t.Error("Expected error message")
	}
}

func TestHealthChecker_EndpointCheck(t *testing.T) {
	checker := NewHealthChecker()

	// Test with unreachable endpoint
	endpointCheck := EndpointHealthCheck{
		URL:            "http://localhost:99999/health",
		Timeout:        time.Second,
		ExpectedStatus: 200,
	}

	result := checker.CheckEndpoint(context.Background(), endpointCheck)

	if result.Healthy {
		t.Error("Expected endpoint check to fail")
	}
}

func TestHealthChecker_ConfigValidation(t *testing.T) {
	checker := NewHealthChecker()

	// Valid config
	config := map[string]interface{}{
		"port":     8080,
		"database": "testdb",
		"required": "value",
	}

	validation := ConfigValidation{
		RequiredKeys: []string{"port", "database", "required"},
		ValidValues: map[string][]interface{}{
			"port": {8080, 8081, 8082},
		},
	}

	result := checker.ValidateConfig(config, validation)

	if !result.Healthy {
		t.Errorf("Expected config validation to pass, got: %s", result.Error)
	}
}

func TestHealthChecker_ConfigValidation_Missing(t *testing.T) {
	checker := NewHealthChecker()

	// Missing required key
	config := map[string]interface{}{
		"port": 8080,
	}

	validation := ConfigValidation{
		RequiredKeys: []string{"port", "database"},
	}

	result := checker.ValidateConfig(config, validation)

	if result.Healthy {
		t.Error("Expected config validation to fail for missing key")
	}
	if result.Error == "" {
		t.Error("Expected error message about missing key")
	}
}

func TestHealthChecker_ConfigValidation_InvalidValue(t *testing.T) {
	checker := NewHealthChecker()

	config := map[string]interface{}{
		"port": 9999, // Invalid value
	}

	validation := ConfigValidation{
		RequiredKeys: []string{"port"},
		ValidValues: map[string][]interface{}{
			"port": {8080, 8081, 8082},
		},
	}

	result := checker.ValidateConfig(config, validation)

	if result.Healthy {
		t.Error("Expected config validation to fail for invalid value")
	}
}

func TestHealthChecker_MultiCheck(t *testing.T) {
	checker := NewHealthChecker()

	checks := []HealthCheck{
		{
			Name: "config_check",
			Type: "config",
			Config: map[string]interface{}{
				"port": 8080,
			},
		},
	}

	results := checker.RunChecks(context.Background(), checks)

	if len(results) != len(checks) {
		t.Errorf("Expected %d results, got %d", len(checks), len(results))
	}
}

func TestHealthChecker_OverallHealth(t *testing.T) {
	results := []HealthCheckResult{
		{Name: "check1", Healthy: true},
		{Name: "check2", Healthy: true},
		{Name: "check3", Healthy: true},
	}

	checker := NewHealthChecker()
	overall := checker.GetOverallHealth(results)

	if !overall.Healthy {
		t.Error("Expected overall health to be healthy when all checks pass")
	}
	if overall.TotalChecks != 3 {
		t.Errorf("TotalChecks = %d, want 3", overall.TotalChecks)
	}
	if overall.PassedChecks != 3 {
		t.Errorf("PassedChecks = %d, want 3", overall.PassedChecks)
	}
}

func TestHealthChecker_OverallHealth_WithFailures(t *testing.T) {
	results := []HealthCheckResult{
		{Name: "check1", Healthy: true},
		{Name: "check2", Healthy: false, Error: "failed"},
		{Name: "check3", Healthy: true},
	}

	checker := NewHealthChecker()
	overall := checker.GetOverallHealth(results)

	if overall.Healthy {
		t.Error("Expected overall health to be unhealthy when some checks fail")
	}
	if overall.PassedChecks != 2 {
		t.Errorf("PassedChecks = %d, want 2", overall.PassedChecks)
	}
	if overall.FailedChecks != 1 {
		t.Errorf("FailedChecks = %d, want 1", overall.FailedChecks)
	}
}

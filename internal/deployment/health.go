package deployment

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

// HealthChecker performs health checks on various system components
type HealthChecker struct{}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}

// HealthCheck defines a health check configuration
type HealthCheck struct {
	Name   string
	Type   string // "database", "endpoint", "config"
	Config map[string]interface{}
}

// HealthCheckResult contains the result of a health check
type HealthCheckResult struct {
	Name      string
	Healthy   bool
	Error     string
	Duration  time.Duration
	Timestamp time.Time
}

// OverallHealth summarizes multiple health check results
type OverallHealth struct {
	Healthy      bool
	TotalChecks  int
	PassedChecks int
	FailedChecks int
	Results      []HealthCheckResult
}

// DatabaseHealthCheck contains database health check configuration
type DatabaseHealthCheck struct {
	DSN     string
	Driver  string
	Timeout time.Duration
}

// EndpointHealthCheck contains HTTP endpoint health check configuration
type EndpointHealthCheck struct {
	URL            string
	Timeout        time.Duration
	ExpectedStatus int
	Method         string
}

// ConfigValidation contains configuration validation rules
type ConfigValidation struct {
	RequiredKeys []string
	ValidValues  map[string][]interface{}
}

// CheckDatabase verifies database connectivity
func (h *HealthChecker) CheckDatabase(ctx context.Context, check DatabaseHealthCheck) HealthCheckResult {
	result := HealthCheckResult{
		Name:      "database",
		Timestamp: time.Now(),
	}
	
	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()
	
	// Create context with timeout
	if check.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, check.Timeout)
		defer cancel()
	}
	
	// Open database connection
	db, err := sql.Open(check.Driver, check.DSN)
	if err != nil {
		result.Error = fmt.Sprintf("failed to open database: %v", err)
		return result
	}
	defer db.Close()
	
	// Ping database
	if err := db.PingContext(ctx); err != nil {
		result.Error = fmt.Sprintf("failed to ping database: %v", err)
		return result
	}
	
	result.Healthy = true
	return result
}

// CheckEndpoint verifies HTTP endpoint availability
func (h *HealthChecker) CheckEndpoint(ctx context.Context, check EndpointHealthCheck) HealthCheckResult {
	result := HealthCheckResult{
		Name:      "endpoint",
		Timestamp: time.Now(),
	}
	
	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()
	
	method := check.Method
	if method == "" {
		method = "GET"
	}
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: check.Timeout,
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, method, check.URL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result
	}
	
	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("failed to reach endpoint: %v", err)
		return result
	}
	defer resp.Body.Close()
	
	// Check status code
	expectedStatus := check.ExpectedStatus
	if expectedStatus == 0 {
		expectedStatus = 200
	}
	
	if resp.StatusCode != expectedStatus {
		result.Error = fmt.Sprintf("unexpected status code: got %d, want %d", resp.StatusCode, expectedStatus)
		return result
	}
	
	result.Healthy = true
	return result
}

// ValidateConfig validates configuration values
func (h *HealthChecker) ValidateConfig(config map[string]interface{}, validation ConfigValidation) HealthCheckResult {
	result := HealthCheckResult{
		Name:      "config",
		Timestamp: time.Now(),
		Healthy:   true,
	}
	
	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()
	
	// Check required keys
	for _, key := range validation.RequiredKeys {
		if _, exists := config[key]; !exists {
			result.Healthy = false
			result.Error = fmt.Sprintf("missing required config key: %s", key)
			return result
		}
	}
	
	// Validate values
	for key, validValues := range validation.ValidValues {
		value, exists := config[key]
		if !exists {
			continue
		}
		
		// Check if value is in valid values list
		valid := false
		for _, validValue := range validValues {
			if value == validValue {
				valid = true
				break
			}
		}
		
		if !valid {
			result.Healthy = false
			result.Error = fmt.Sprintf("invalid value for %s: got %v, want one of %v", key, value, validValues)
			return result
		}
	}
	
	return result
}

// RunChecks executes multiple health checks
func (h *HealthChecker) RunChecks(ctx context.Context, checks []HealthCheck) []HealthCheckResult {
	results := make([]HealthCheckResult, 0, len(checks))
	
	for _, check := range checks {
		var result HealthCheckResult
		
		switch check.Type {
		case "database":
			// Extract database check config
			dbCheck := DatabaseHealthCheck{
				DSN:    getStringFromConfig(check.Config, "dsn"),
				Driver: getStringFromConfig(check.Config, "driver"),
			}
			if timeout := getIntFromConfig(check.Config, "timeout"); timeout > 0 {
				dbCheck.Timeout = time.Duration(timeout) * time.Second
			}
			result = h.CheckDatabase(ctx, dbCheck)
			
		case "endpoint":
			// Extract endpoint check config
			endpointCheck := EndpointHealthCheck{
				URL:            getStringFromConfig(check.Config, "url"),
				ExpectedStatus: getIntFromConfig(check.Config, "expected_status"),
				Method:         getStringFromConfig(check.Config, "method"),
			}
			if timeout := getIntFromConfig(check.Config, "timeout"); timeout > 0 {
				endpointCheck.Timeout = time.Duration(timeout) * time.Second
			}
			result = h.CheckEndpoint(ctx, endpointCheck)
			
		case "config":
			validation := ConfigValidation{
				RequiredKeys: getStringSliceFromConfig(check.Config, "required_keys"),
			}
			result = h.ValidateConfig(check.Config, validation)
			
		default:
			result = HealthCheckResult{
				Name:      check.Name,
				Healthy:   false,
				Error:     fmt.Sprintf("unknown check type: %s", check.Type),
				Timestamp: time.Now(),
			}
		}
		
		result.Name = check.Name
		results = append(results, result)
	}
	
	return results
}

// GetOverallHealth summarizes health check results
func (h *HealthChecker) GetOverallHealth(results []HealthCheckResult) OverallHealth {
	overall := OverallHealth{
		Healthy:      true,
		TotalChecks:  len(results),
		PassedChecks: 0,
		FailedChecks: 0,
		Results:      results,
	}
	
	for _, result := range results {
		if result.Healthy {
			overall.PassedChecks++
		} else {
			overall.FailedChecks++
			overall.Healthy = false
		}
	}
	
	return overall
}

// Helper functions to extract values from config map

func getStringFromConfig(config map[string]interface{}, key string) string {
	if v, ok := config[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getIntFromConfig(config map[string]interface{}, key string) int {
	if v, ok := config[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return 0
}

func getStringSliceFromConfig(config map[string]interface{}, key string) []string {
	if v, ok := config[key]; ok {
		if slice, ok := v.([]string); ok {
			return slice
		}
		if slice, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

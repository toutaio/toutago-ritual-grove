package generator

import (
	"strings"
)

// DockerImage returns the appropriate Docker image for a database type
func DockerImage(databaseType string) string {
	switch databaseType {
	case "postgres":
		return "postgres:16-alpine"
	case "mysql":
		return "mysql:8-alpine"
	default:
		return ""
	}
}

// DockerPort returns the default port for a database type
func DockerPort(databaseType string) int {
	switch databaseType {
	case "postgres":
		return 5432
	case "mysql":
		return 3306
	default:
		return 0
	}
}

// HealthCheck returns the health check command for a database type
func HealthCheck(databaseType string) string {
	switch databaseType {
	case "postgres":
		return "pg_isready -U ${DB_USER}"
	case "mysql":
		return "mysqladmin ping -h localhost -u root -p${DB_ROOT_PASSWORD}"
	default:
		return ""
	}
}

// HasFrontend returns true if the frontend type requires a separate build service
func HasFrontend(frontendType string) bool {
	switch frontendType {
	case "inertia-vue":
		return true
	case "htmx", "traditional", "":
		return false
	default:
		return false
	}
}

// DBUser generates a default database username from project name
func DBUser(projectName string) string {
	// Replace dashes with underscores and append _user
	clean := strings.ReplaceAll(projectName, "-", "_")
	return clean + "_user"
}

// DBName generates a default database name from project name
func DBName(projectName string) string {
	// Replace dashes with underscores and append _db
	clean := strings.ReplaceAll(projectName, "-", "_")
	return clean + "_db"
}

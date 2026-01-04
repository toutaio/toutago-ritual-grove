package questionnaire

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// PathCheckType defines the type of path validation to perform
type PathCheckType int

const (
	PathCheckAny PathCheckType = iota
	PathCheckFile
	PathCheckDirectory
	PathCheckWritable
)

// PathValidationResult contains information about a validated path
type PathValidationResult struct {
	Exists      bool
	IsFile      bool
	IsDirectory bool
	IsWritable  bool
	AbsPath     string
}

// DatabaseHelper provides utilities for testing database connections
type DatabaseHelper struct{}

// TestConnection tests if a database connection can be established
func (h *DatabaseHelper) TestConnection(ctx context.Context, dsn, dbType string) error {
	if dsn == "" {
		return errors.New("DSN cannot be empty")
	}

	// Normalize database type
	dbType = strings.ToLower(dbType)

	// Validate database type
	switch dbType {
	case "postgres", "postgresql":
		dbType = "postgres"
	case "mysql":
		dbType = "mysql"
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Try to open connection
	db, err := sql.Open(dbType, dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Log but don't fail on close error
		}
	}()

	// Test connection with context
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// URLHelper provides utilities for checking URL availability
type URLHelper struct{}

// CheckAvailability checks if a URL is reachable
func (h *URLHelper) CheckAvailability(ctx context.Context, url string) error {
	if url == "" {
		return errors.New("URL cannot be empty")
	}

	// Parse URL to validate format
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("URL must start with http:// or https://")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Perform request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach URL: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log but don't fail on close error
		}
	}()

	// Check if response is successful (2xx or 3xx)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("URL returned status code %d", resp.StatusCode)
	}

	return nil
}

// PathHelper provides utilities for validating file paths
type PathHelper struct{}

// ValidatePath validates a file system path based on the check type
func (h *PathHelper) ValidatePath(path string, checkType PathCheckType) (*PathValidationResult, error) {
	if path == "" {
		return nil, errors.New("path cannot be empty")
	}

	result := &PathValidationResult{}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	result.AbsPath = absPath

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Exists = false

			// For writable check, verify parent directory is writable
			if checkType == PathCheckWritable {
				parentDir := filepath.Dir(absPath)
				parentInfo, parentErr := os.Stat(parentDir)
				if parentErr != nil {
					return nil, fmt.Errorf("parent directory does not exist: %w", parentErr)
				}

				// Check if parent is writable
				if parentInfo.Mode().Perm()&0200 != 0 {
					result.IsWritable = true
					return result, nil
				}
				return nil, errors.New("parent directory is not writable")
			}

			// For other checks on non-existent paths
			if checkType == PathCheckFile || checkType == PathCheckDirectory {
				return nil, fmt.Errorf("path does not exist: %s", absPath)
			}

			return result, nil
		}
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	result.Exists = true
	result.IsFile = info.Mode().IsRegular()
	result.IsDirectory = info.IsDir()

	// Perform type-specific validation
	switch checkType {
	case PathCheckFile:
		if !result.IsFile {
			return nil, errors.New("path exists but is not a file")
		}
	case PathCheckDirectory:
		if !result.IsDirectory {
			return nil, errors.New("path exists but is not a directory")
		}
	case PathCheckWritable:
		// Check if path is writable
		if info.Mode().Perm()&0200 != 0 {
			result.IsWritable = true
		} else {
			return nil, errors.New("path is not writable")
		}
	}

	return result, nil
}

// GitHelper provides utilities for validating Git repositories
type GitHelper struct{}

// ValidateRepository validates if a path is a valid Git repository
func (h *GitHelper) ValidateRepository(path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("path does not exist")
		}
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return errors.New("path is not a directory")
	}

	// Check if .git directory exists
	gitDir := filepath.Join(path, ".git")
	gitInfo, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("not a git repository (no .git directory found)")
		}
		return fmt.Errorf("failed to check .git directory: %w", err)
	}

	if !gitInfo.IsDir() {
		return errors.New(".git exists but is not a directory")
	}

	// Try to run git status to verify it's a valid repository
	cmd := exec.Command("git", "-C", path, "status")
	if err := cmd.Run(); err != nil {
		return errors.New("directory contains .git but is not a valid git repository")
	}

	return nil
}

// PortHelper provides utilities for checking port availability
type PortHelper struct{}

// CheckAvailability checks if a port is available
func (h *PortHelper) CheckAvailability(ctx context.Context, port int) (bool, error) {
	if port <= 0 || port > 65535 {
		return false, fmt.Errorf("invalid port number: %d (must be between 1 and 65535)", port)
	}

	// Try to listen on the port
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// Port is likely in use or permission denied
		if strings.Contains(err.Error(), "address already in use") {
			return false, nil // Port is not available
		}
		return false, fmt.Errorf("failed to check port: %w", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			// Log but don't fail on close error
		}
	}()

	return true, nil // Port is available
}

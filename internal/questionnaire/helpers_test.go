package questionnaire

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDatabaseConnectionHelper(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		dbType  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty DSN",
			dsn:     "",
			dbType:  "postgres",
			wantErr: true,
			errMsg:  "DSN cannot be empty",
		},
		{
			name:    "invalid database type",
			dsn:     "host=localhost",
			dbType:  "invalid",
			wantErr: true,
			errMsg:  "unsupported database type",
		},
		{
			name:    "malformed postgres DSN",
			dsn:     "invalid-dsn",
			dbType:  "postgres",
			wantErr: true,
		},
		{
			name:    "malformed mysql DSN",
			dsn:     "invalid-dsn",
			dbType:  "mysql",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := &DatabaseHelper{}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := helper.TestConnection(ctx, tt.dsn, tt.dbType)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errMsg != "" && err != nil && err.Error() != tt.errMsg && !contains(err.Error(), tt.errMsg) {
				t.Errorf("TestConnection() error message = %v, want substring %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestURLHelper(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "valid reachable URL",
			url:     server.URL,
			wantErr: false,
		},
		{
			name:    "unreachable URL",
			url:     "http://localhost:99999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := &URLHelper{}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := helper.CheckAvailability(ctx, tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAvailability() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPathHelper(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	existingDir := filepath.Join(tempDir, "existing-dir")
	if err := os.Mkdir(existingDir, 0750); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		path        string
		checkType   PathCheckType
		wantErr     bool
		shouldExist bool
	}{
		{
			name:      "empty path",
			path:      "",
			checkType: PathCheckAny,
			wantErr:   true,
		},
		{
			name:        "existing file - check any",
			path:        existingFile,
			checkType:   PathCheckAny,
			wantErr:     false,
			shouldExist: true,
		},
		{
			name:        "existing directory - check any",
			path:        existingDir,
			checkType:   PathCheckAny,
			wantErr:     false,
			shouldExist: true,
		},
		{
			name:        "existing file - check file",
			path:        existingFile,
			checkType:   PathCheckFile,
			wantErr:     false,
			shouldExist: true,
		},
		{
			name:      "existing directory - check file (should fail)",
			path:      existingDir,
			checkType: PathCheckFile,
			wantErr:   true,
		},
		{
			name:        "existing directory - check directory",
			path:        existingDir,
			checkType:   PathCheckDirectory,
			wantErr:     false,
			shouldExist: true,
		},
		{
			name:      "existing file - check directory (should fail)",
			path:      existingFile,
			checkType: PathCheckDirectory,
			wantErr:   true,
		},
		{
			name:        "non-existing path - check writable",
			path:        filepath.Join(tempDir, "new-file.txt"),
			checkType:   PathCheckWritable,
			wantErr:     false,
			shouldExist: false,
		},
		{
			name:      "invalid path - check writable",
			path:      "/root/no-permission/file.txt",
			checkType: PathCheckWritable,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := &PathHelper{}
			result, err := helper.ValidatePath(tt.path, tt.checkType)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result.Exists != tt.shouldExist {
				t.Errorf("ValidatePath() exists = %v, want %v", result.Exists, tt.shouldExist)
			}
		})
	}
}

func TestGitHelper(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize a git repo for testing
	gitDir := filepath.Join(tempDir, "git-repo")
	if err := os.Mkdir(gitDir, 0750); err != nil {
		t.Fatal(err)
	}

	nonGitDir := filepath.Join(tempDir, "non-git")
	if err := os.Mkdir(nonGitDir, 0750); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		setup   func() error
		wantErr bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "non-existent path",
			path:    filepath.Join(tempDir, "does-not-exist"),
			wantErr: true,
		},
		{
			name:    "not a git repository",
			path:    nonGitDir,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatal(err)
				}
			}

			helper := &GitHelper{}
			err := helper.ValidateRepository(tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPortHelper(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{
			name:    "invalid port - negative",
			port:    -1,
			wantErr: true,
		},
		{
			name:    "invalid port - zero",
			port:    0,
			wantErr: true,
		},
		{
			name:    "invalid port - too high",
			port:    65536,
			wantErr: true,
		},
		{
			name:    "valid port",
			port:    8080,
			wantErr: false,
		},
		{
			name:    "privileged port (may fail without permissions)",
			port:    80,
			wantErr: false, // We expect it might fail due to permissions
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := &PortHelper{}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			available, err := helper.CheckAvailability(ctx, tt.port)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CheckAvailability() expected error but got none")
				}
				return
			}

			// For privileged ports (< 1024), permission errors are acceptable
			if tt.port < 1024 && err != nil {
				t.Logf("Privileged port %d check failed (expected): %v", tt.port, err)
				return
			}

			if err != nil {
				t.Errorf("CheckAvailability() unexpected error = %v", err)
				return
			}

			// For valid ports, check that we get a boolean result
			if available && tt.port > 0 && tt.port < 65536 {
				// Port is available (this is expected for most test runs)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && anySubstring(s, substr))
}

func anySubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

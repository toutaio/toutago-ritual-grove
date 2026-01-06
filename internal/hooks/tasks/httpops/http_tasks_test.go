package httpops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

func TestHTTPGetTask(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantErr   bool
		setupMock func() *httptest.Server
	}{
		{
			name: "successful GET request",
			url:  "",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet {
						t.Errorf("expected GET, got %s", r.Method)
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("success"))
				}))
			},
			wantErr: false,
		},
		{
			name:    "missing URL",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupMock != nil {
				server = tt.setupMock()
				defer server.Close()
				tt.url = server.URL
			}

			task := &HTTPGetTask{URL: tt.url}
			err := task.Validate()
			if tt.name == "missing URL" {
				if err == nil {
					t.Error("expected validation error for missing URL")
				}
				return
			}

			if err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			ctx := context.Background()
			taskCtx := &tasks.TaskContext{}
			err = task.Execute(ctx, taskCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPPostTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	task := &HTTPPostTask{
		URL:  server.URL,
		Body: `{"key":"value"}`,
	}

	if err := task.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	ctx := context.Background()
	taskCtx := &tasks.TaskContext{}
	err := task.Execute(ctx, taskCtx)

	if err != nil {
		t.Errorf("Execute() failed: %v", err)
	}
}

func TestHTTPDownloadTask(t *testing.T) {
	content := "test file content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "download.txt")

	task := &HTTPDownloadTask{
		URL:    server.URL,
		Output: outputPath,
	}

	if err := task.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	ctx := context.Background()
	taskCtx := &tasks.TaskContext{}
	err := task.Execute(ctx, taskCtx)

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify file was created with correct content.
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}

	if string(data) != content {
		t.Errorf("downloaded content = %q, want %q", string(data), content)
	}
}

func TestHTTPHealthCheckTask(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		retries    int
		delay      string
		wantErr    bool
	}{
		{
			name:       "successful health check",
			statusCode: http.StatusOK,
			retries:    1,
			delay:      "100ms",
			wantErr:    false,
		},
		{
			name:       "health check with retries",
			statusCode: http.StatusServiceUnavailable,
			retries:    2,
			delay:      "100ms",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			task := &HTTPHealthCheckTask{
				URL:     server.URL,
				Retries: tt.retries,
				Delay:   tt.delay,
			}

			if err := task.Validate(); err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			ctx := context.Background()
			taskCtx := &tasks.TaskContext{}
			err := task.Execute(ctx, taskCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPTaskValidation(t *testing.T) {
	tests := []struct {
		name    string
		task    tasks.Task
		wantErr bool
	}{
		{
			name:    "HTTPGetTask missing URL",
			task:    &HTTPGetTask{},
			wantErr: true,
		},
		{
			name:    "HTTPPostTask missing URL",
			task:    &HTTPPostTask{},
			wantErr: true,
		},
		{
			name:    "HTTPDownloadTask missing URL",
			task:    &HTTPDownloadTask{Output: "/tmp/file"},
			wantErr: true,
		},
		{
			name:    "HTTPDownloadTask missing output",
			task:    &HTTPDownloadTask{URL: "http://example.com"},
			wantErr: true,
		},
		{
			name:    "HTTPHealthCheckTask missing URL",
			task:    &HTTPHealthCheckTask{},
			wantErr: true,
		},
		{
			name:    "HTTPHealthCheckTask invalid delay",
			task:    &HTTPHealthCheckTask{URL: "http://example.com", Delay: "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

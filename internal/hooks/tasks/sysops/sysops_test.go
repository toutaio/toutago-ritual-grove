package sysops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

func TestWaitForServiceTask_Name(t *testing.T) {
	task := &WaitForServiceTask{}
	if got := task.Name(); got != "wait-for-service" {
		t.Errorf("WaitForServiceTask.Name() = %v, want %v", got, "wait-for-service")
	}
}

func TestWaitForServiceTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    *WaitForServiceTask
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid HTTP URL",
			task: &WaitForServiceTask{
				URL:     "http://localhost:8080/health",
				Timeout: 30,
			},
			wantErr: false,
		},
		{
			name: "valid TCP host:port",
			task: &WaitForServiceTask{
				Host:    "localhost",
				Port:    5432,
				Timeout: 10,
			},
			wantErr: false,
		},
		{
			name:    "missing URL and host",
			task:    &WaitForServiceTask{},
			wantErr: true,
			errMsg:  "either url or host:port is required",
		},
		{
			name: "both URL and host provided",
			task: &WaitForServiceTask{
				URL:  "http://localhost:8080",
				Host: "localhost",
				Port: 8080,
			},
			wantErr: true,
			errMsg:  "cannot specify both url and host:port",
		},
		{
			name: "host without port",
			task: &WaitForServiceTask{
				Host: "localhost",
			},
			wantErr: true,
			errMsg:  "port is required when host is specified",
		},
		{
			name: "missing timeout defaults",
			task: &WaitForServiceTask{
				URL: "http://localhost:8080",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("WaitForServiceTask.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("WaitForServiceTask.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("WaitForServiceTask.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestWaitForServiceTask_ExecuteHTTP(t *testing.T) {
	// Create test HTTP server
	ready := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ready {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer server.Close()

	taskCtx := tasks.NewTaskContext()
	ctx := context.Background()

	t.Run("service becomes ready", func(t *testing.T) {
		task := &WaitForServiceTask{
			URL:      server.URL,
			Timeout:  5,
			Interval: 1,
		}

		// Make server ready after delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			ready = true
		}()

		err := task.Execute(ctx, taskCtx)
		if err != nil {
			t.Errorf("WaitForServiceTask.Execute() unexpected error = %v", err)
		}
	})

	t.Run("timeout waiting for service", func(t *testing.T) {
		ready = false
		task := &WaitForServiceTask{
			URL:      server.URL,
			Timeout:  1,
			Interval: 1,
		}

		err := task.Execute(ctx, taskCtx)
		if err == nil {
			t.Error("WaitForServiceTask.Execute() expected timeout error")
		}
	})
}

func TestNotifyTask_Name(t *testing.T) {
	task := &NotifyTask{}
	if got := task.Name(); got != "notify" {
		t.Errorf("NotifyTask.Name() = %v, want %v", got, "notify")
	}
}

func TestNotifyTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    *NotifyTask
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid log notification",
			task: &NotifyTask{
				Type:    "log",
				Message: "Deployment complete",
			},
			wantErr: false,
		},
		{
			name: "valid webhook notification",
			task: &NotifyTask{
				Type:    "webhook",
				Message: "Build successful",
				URL:     "https://hooks.slack.com/test",
			},
			wantErr: false,
		},
		{
			name:    "missing message",
			task:    &NotifyTask{Type: "log"},
			wantErr: true,
			errMsg:  "message is required",
		},
		{
			name: "invalid type",
			task: &NotifyTask{
				Type:    "invalid",
				Message: "test",
			},
			wantErr: true,
			errMsg:  "type must be 'log' or 'webhook'",
		},
		{
			name: "webhook without URL",
			task: &NotifyTask{
				Type:    "webhook",
				Message: "test",
			},
			wantErr: true,
			errMsg:  "url is required for webhook notifications",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("NotifyTask.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("NotifyTask.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("NotifyTask.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestNotifyTask_ExecuteLog(t *testing.T) {
	task := &NotifyTask{
		Type:    "log",
		Message: "Test notification",
		Level:   "info",
	}

	taskCtx := tasks.NewTaskContext()
	ctx := context.Background()

	err := task.Execute(ctx, taskCtx)
	if err != nil {
		t.Errorf("NotifyTask.Execute() unexpected error = %v", err)
	}
}

func TestNotifyTask_ExecuteWebhook(t *testing.T) {
	received := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	task := &NotifyTask{
		Type:    "webhook",
		Message: "Test webhook",
		URL:     server.URL,
	}

	taskCtx := tasks.NewTaskContext()
	ctx := context.Background()

	err := task.Execute(ctx, taskCtx)
	if err != nil {
		t.Errorf("NotifyTask.Execute() unexpected error = %v", err)
	}

	if !received {
		t.Error("Webhook was not called")
	}
}

func TestTaskRegistration(t *testing.T) {
	// Test that tasks are registered correctly
	t.Run("wait-for-service registration", func(t *testing.T) {
		task, err := tasks.Create("wait-for-service", map[string]interface{}{
			"url":     "http://localhost:8080",
			"timeout": 30,
		})
		if err != nil {
			t.Errorf("Failed to create wait-for-service task: %v", err)
		}
		if task == nil {
			t.Error("Expected task to be created")
		}
		if task.Name() != "wait-for-service" {
			t.Errorf("Expected task name 'wait-for-service', got %v", task.Name())
		}
	})

	t.Run("notify registration", func(t *testing.T) {
		task, err := tasks.Create("notify", map[string]interface{}{
			"type":    "log",
			"message": "test",
		})
		if err != nil {
			t.Errorf("Failed to create notify task: %v", err)
		}
		if task == nil {
			t.Error("Expected task to be created")
		}
		if task.Name() != "notify" {
			t.Errorf("Expected task name 'notify', got %v", task.Name())
		}
	})
}

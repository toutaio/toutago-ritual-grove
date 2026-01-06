// Package sysops provides system operation tasks for hooks.
package sysops

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

// WaitForServiceTask waits for a service to become available.
type WaitForServiceTask struct {
	URL      string // HTTP/HTTPS URL to check
	Host     string // TCP host to check
	Port     int    // TCP port to check
	Timeout  int    // Timeout in seconds (default: 60)
	Interval int    // Interval between checks in seconds (default: 2)
}

func (t *WaitForServiceTask) Name() string {
	return "wait-for-service"
}

func (t *WaitForServiceTask) Validate() error {
	if t.URL == "" && t.Host == "" {
		return errors.New("either url or host:port is required")
	}
	if t.URL != "" && t.Host != "" {
		return errors.New("cannot specify both url and host:port")
	}
	if t.Host != "" && t.Port == 0 {
		return errors.New("port is required when host is specified")
	}
	return nil
}

func (t *WaitForServiceTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	timeout := t.Timeout
	if timeout == 0 {
		timeout = 60
	}
	interval := t.Interval
	if interval == 0 {
		interval = 2
	}

	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		var available bool
		var err error

		if t.URL != "" {
			available, err = t.checkHTTP()
		} else {
			available, err = t.checkTCP()
		}

		if available {
			return nil
		}

		if time.Now().After(deadline) {
			if err != nil {
				return fmt.Errorf("timeout waiting for service: %w", err)
			}
			return fmt.Errorf("timeout waiting for service after %d seconds", timeout)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			continue
		}
	}
}

func (t *WaitForServiceTask) checkHTTP() (bool, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Get(t.URL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300, nil
}

func (t *WaitForServiceTask) checkTCP() (bool, error) {
	address := fmt.Sprintf("%s:%d", t.Host, t.Port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return false, err
	}
	conn.Close()
	return true, nil
}

// NotifyTask sends a notification.
type NotifyTask struct {
	Type    string            // Notification type: "log" or "webhook"
	Message string            // Notification message
	Level   string            // Log level for log type (info, warn, error)
	URL     string            // Webhook URL for webhook type
	Headers map[string]string // Optional headers for webhook
}

func (t *NotifyTask) Name() string {
	return "notify"
}

func (t *NotifyTask) Validate() error {
	if t.Message == "" {
		return errors.New("message is required")
	}
	if t.Type != "log" && t.Type != "webhook" {
		return errors.New("type must be 'log' or 'webhook'")
	}
	if t.Type == "webhook" && t.URL == "" {
		return errors.New("url is required for webhook notifications")
	}
	return nil
}

func (t *NotifyTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	switch t.Type {
	case "log":
		return t.executeLog()
	case "webhook":
		return t.executeWebhook(ctx)
	default:
		return fmt.Errorf("unsupported notification type: %s", t.Type)
	}
}

func (t *NotifyTask) executeLog() error {
	level := t.Level
	if level == "" {
		level = "info"
	}

	prefix := fmt.Sprintf("[%s] ", level)
	log.Printf("%s%s\n", prefix, t.Message)
	return nil
}

func (t *NotifyTask) executeWebhook(ctx context.Context) error {
	payload := map[string]string{
		"message": t.Message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	return nil
}

// Register system operation tasks.
func init() {
	tasks.Register("wait-for-service", func(config map[string]interface{}) (tasks.Task, error) {
		url, _ := config["url"].(string)
		host, _ := config["host"].(string)
		port, _ := config["port"].(int)
		timeout, _ := config["timeout"].(int)
		interval, _ := config["interval"].(int)

		// Handle float64 from JSON unmarshaling
		if port == 0 {
			if portFloat, ok := config["port"].(float64); ok {
				port = int(portFloat)
			}
		}
		if timeout == 0 {
			if timeoutFloat, ok := config["timeout"].(float64); ok {
				timeout = int(timeoutFloat)
			}
		}
		if interval == 0 {
			if intervalFloat, ok := config["interval"].(float64); ok {
				interval = int(intervalFloat)
			}
		}

		task := &WaitForServiceTask{
			URL:      url,
			Host:     host,
			Port:     port,
			Timeout:  timeout,
			Interval: interval,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})

	tasks.Register("notify", func(config map[string]interface{}) (tasks.Task, error) {
		notifType, _ := config["type"].(string)
		message, _ := config["message"].(string)
		level, _ := config["level"].(string)
		url, _ := config["url"].(string)

		headers := make(map[string]string)
		if h, ok := config["headers"].(map[string]interface{}); ok {
			for k, v := range h {
				if str, ok := v.(string); ok {
					headers[k] = str
				}
			}
		}

		task := &NotifyTask{
			Type:    notifType,
			Message: message,
			Level:   level,
			URL:     url,
			Headers: headers,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})
}

package httpops

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/toutaio/toutago-ritual-grove/internal/hooks/tasks"
)

// HTTPGetTask sends an HTTP GET request.
type HTTPGetTask struct {
	URL     string            // URL to GET
	Headers map[string]string // Optional headers
}

func (t *HTTPGetTask) Name() string {
	return "http-get"
}

func (t *HTTPGetTask) Validate() error {
	if t.URL == "" {
		return errors.New("url is required")
	}
	return nil
}

func (t *HTTPGetTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range t.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// HTTPPostTask sends an HTTP POST request.
type HTTPPostTask struct {
	URL     string            // URL to POST
	Body    string            // Request body
	Headers map[string]string // Optional headers
}

func (t *HTTPPostTask) Name() string {
	return "http-post"
}

func (t *HTTPPostTask) Validate() error {
	if t.URL == "" {
		return errors.New("url is required")
	}
	return nil
}

func (t *HTTPPostTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.URL, strings.NewReader(t.Body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set default content type.
	req.Header.Set("Content-Type", "application/json")

	for key, value := range t.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// HTTPDownloadTask downloads a file from a URL.
type HTTPDownloadTask struct {
	URL    string // URL to download from
	Output string // Output file path
}

func (t *HTTPDownloadTask) Name() string {
	return "http-download"
}

func (t *HTTPDownloadTask) Validate() error {
	if t.URL == "" {
		return errors.New("url is required")
	}
	if t.Output == "" {
		return errors.New("output is required")
	}
	return nil
}

func (t *HTTPDownloadTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	outputPath := t.Output
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(taskCtx.WorkingDir(), outputPath)
	}

	// Ensure output directory exists.
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// HTTPHealthCheckTask performs health checks with retries.
type HTTPHealthCheckTask struct {
	URL     string // URL to check
	Retries int    // Number of retries (default: 3)
	Delay   string // Delay between retries (default: "1s")
}

func (t *HTTPHealthCheckTask) Name() string {
	return "http-health-check"
}

func (t *HTTPHealthCheckTask) Validate() error {
	if t.URL == "" {
		return errors.New("url is required")
	}

	if t.Delay != "" {
		if _, err := time.ParseDuration(t.Delay); err != nil {
			return fmt.Errorf("invalid delay format: %w", err)
		}
	}

	return nil
}

func (t *HTTPHealthCheckTask) Execute(ctx context.Context, taskCtx *tasks.TaskContext) error {
	retries := t.Retries
	if retries == 0 {
		retries = 3
	}

	delay := 1 * time.Second
	if t.Delay != "" {
		var err error
		delay, err = time.ParseDuration(t.Delay)
		if err != nil {
			return fmt.Errorf("invalid delay: %w", err)
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for i := 0; i < retries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
	}

	return fmt.Errorf("health check failed after %d attempts", retries)
}

// Register HTTP operation tasks.
func init() {
	tasks.Register("http-get", func(config map[string]interface{}) (tasks.Task, error) {
		url, _ := config["url"].(string)
		headers, _ := config["headers"].(map[string]string)

		task := &HTTPGetTask{
			URL:     url,
			Headers: headers,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})

	tasks.Register("http-post", func(config map[string]interface{}) (tasks.Task, error) {
		url, _ := config["url"].(string)
		body, _ := config["body"].(string)
		headers, _ := config["headers"].(map[string]string)

		task := &HTTPPostTask{
			URL:     url,
			Body:    body,
			Headers: headers,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})

	tasks.Register("http-download", func(config map[string]interface{}) (tasks.Task, error) {
		url, _ := config["url"].(string)
		output, _ := config["output"].(string)

		task := &HTTPDownloadTask{
			URL:    url,
			Output: output,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})

	tasks.Register("http-health-check", func(config map[string]interface{}) (tasks.Task, error) {
		url, _ := config["url"].(string)
		retries, _ := config["retries"].(int)
		delay, _ := config["delay"].(string)

		task := &HTTPHealthCheckTask{
			URL:     url,
			Retries: retries,
			Delay:   delay,
		}

		if err := task.Validate(); err != nil {
			return nil, err
		}

		return task, nil
	})
}

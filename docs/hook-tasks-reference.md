# Hook Tasks Reference

This document provides a comprehensive reference for all available hook tasks in the Ritual Grove system.

## Table of Contents

- [File Operation Tasks](#file-operation-tasks)
- [Go Operation Tasks](#go-operation-tasks)
- [Database Operation Tasks](#database-operation-tasks)
- [HTTP Operation Tasks](#http-operation-tasks)
- [Validation Tasks](#validation-tasks)
- [Environment Tasks](#environment-tasks)
- [System Operation Tasks](#system-operation-tasks)
- [Inertia.js Tasks](#inertiajs-tasks)

## File Operation Tasks

### mkdir

Creates a directory with specified permissions.

**Configuration:**
```yaml
hooks:
  pre-install:
    - task: mkdir
      path: "storage/logs"
      perm: 0755
```

**Parameters:**
- `path` (required): Directory path to create
- `perm` (required): File permissions (octal, e.g., 0755)

### copy

Copies files or directories.

**Configuration:**
```yaml
hooks:
  post-install:
    - task: copy
      src: "templates/config.yaml"
      dest: "config/app.yaml"
```

**Parameters:**
- `src` (required): Source file or directory
- `dest` (required): Destination path

### move

Moves files or directories.

**Configuration:**
```yaml
hooks:
  pre-update:
    - task: move
      src: "old-config.yaml"
      dest: "backup/old-config.yaml"
```

**Parameters:**
- `src` (required): Source file or directory
- `dest` (required): Destination path

### remove

Removes files or directories.

**Configuration:**
```yaml
hooks:
  post-update:
    - task: remove
      path: "tmp/cache"
```

**Parameters:**
- `path` (required): File or directory to remove

### chmod

Changes file or directory permissions (cross-platform).

**Configuration:**
```yaml
hooks:
  post-install:
    - task: chmod
      path: "scripts/deploy.sh"
      perm: 0755
```

**Parameters:**
- `path` (required): File or directory path
- `perm` (required): New permissions (octal, e.g., 0644)

### template-render

Renders a Go template to a file.

**Configuration:**
```yaml
hooks:
  post-install:
    - task: template-render
      template: "templates/readme.tmpl"
      dest: "README.md"
      data:
        project_name: "My Project"
        version: "1.0.0"
```

**Parameters:**
- `template` (required): Path to template file
- `dest` (required): Output file path
- `data` (optional): Template data as key-value pairs

### validate-files

Validates that required files exist.

**Configuration:**
```yaml
hooks:
  pre-install:
    - task: validate-files
      files:
        - "go.mod"
        - "main.go"
        - "config/app.yaml"
```

**Parameters:**
- `files` (required): List of file paths to check

## Go Operation Tasks

### go-mod-tidy

Runs `go mod tidy` to clean up dependencies.

**Configuration:**
```yaml
hooks:
  post-install:
    - task: go-mod-tidy
```

**Parameters:** None

### go-mod-download

Downloads Go module dependencies.

**Configuration:**
```yaml
hooks:
  pre-install:
    - task: go-mod-download
```

**Parameters:** None

### go-build

Builds a Go binary.

**Configuration:**
```yaml
hooks:
  post-install:
    - task: go-build
      output: "bin/myapp"
      package: "./cmd/myapp"
      ldflags: "-s -w"
```

**Parameters:**
- `output` (optional): Output binary path
- `package` (optional): Package to build (default: ".")
- `ldflags` (optional): Linker flags

### go-test

Runs Go tests.

**Configuration:**
```yaml
hooks:
  pre-deploy:
    - task: go-test
      args: ["-v", "-cover"]
      packages: ["./..."]
```

**Parameters:**
- `args` (optional): Additional test arguments
- `packages` (optional): Packages to test (default: ["./..."])

### go-fmt

Formats Go code.

**Configuration:**
```yaml
hooks:
  pre-commit:
    - task: go-fmt
```

**Parameters:** None

### go-run

Runs a Go program.

**Configuration:**
```yaml
hooks:
  post-install:
    - task: go-run
      package: "./cmd/setup"
      args: ["--init"]
      env:
        ENV: "development"
```

**Parameters:**
- `package` (required): Package to run
- `args` (optional): Command-line arguments
- `env` (optional): Environment variables

### exec-go

Executes an arbitrary Go command.

**Configuration:**
```yaml
hooks:
  pre-build:
    - task: exec-go
      command: ["generate", "./..."]
```

**Parameters:**
- `command` (required): Go subcommand and arguments

## Database Operation Tasks

### db-migrate

Runs database migrations (implementation in `db_migrate.go`).

**Configuration:**
```yaml
hooks:
  post-install:
    - task: db-migrate
      direction: "up"
      steps: 0
```

**Parameters:**
- `direction` (optional): "up" or "down" (default: "up")
- `steps` (optional): Number of migrations to run (0 = all)

**Note:** Other database tasks (db-backup, db-restore, db-seed, db-exec) are placeholder implementations that require database connection integration.

## HTTP Operation Tasks

### http-get

Sends an HTTP GET request.

**Configuration:**
```yaml
hooks:
  pre-deploy:
    - task: http-get
      url: "https://api.example.com/status"
      headers:
        Authorization: "Bearer ${API_TOKEN}"
```

**Parameters:**
- `url` (required): Request URL
- `headers` (optional): Request headers as key-value pairs

### http-post

Sends an HTTP POST request.

**Configuration:**
```yaml
hooks:
  post-deploy:
    - task: http-post
      url: "https://hooks.slack.com/services/XXX"
      body: '{"text": "Deployment complete"}'
      headers:
        Content-Type: "application/json"
```

**Parameters:**
- `url` (required): Request URL
- `body` (optional): Request body
- `headers` (optional): Request headers

### http-download

Downloads a file from a URL.

**Configuration:**
```yaml
hooks:
  pre-install:
    - task: http-download
      url: "https://example.com/asset.zip"
      dest: "downloads/asset.zip"
```

**Parameters:**
- `url` (required): Download URL
- `dest` (required): Destination file path

### http-health-check

Performs health check with retries.

**Configuration:**
```yaml
hooks:
  post-deploy:
    - task: http-health-check
      url: "http://localhost:8080/health"
      retries: 5
      delay: 2
```

**Parameters:**
- `url` (required): Health check URL
- `retries` (optional): Number of retries (default: 3)
- `delay` (optional): Delay between retries in seconds (default: 1)

## Validation Tasks

### validate-go-version

Checks if Go version meets minimum requirement.

**Configuration:**
```yaml
hooks:
  pre-install:
    - task: validate-go-version
      min_version: "1.21.0"
```

**Parameters:**
- `min_version` (required): Minimum Go version (semver format)

### validate-dependencies

Verifies that required commands are available.

**Configuration:**
```yaml
hooks:
  pre-install:
    - task: validate-dependencies
      commands: ["git", "docker", "npm"]
```

**Parameters:**
- `commands` (required): List of required command names

### validate-config

Validates that a configuration file exists and is valid.

**Configuration:**
```yaml
hooks:
  pre-deploy:
    - task: validate-config
      file: "config/production.yaml"
```

**Parameters:**
- `file` (required): Configuration file path

### port-check

Checks if a port is available.

**Configuration:**
```yaml
hooks:
  pre-install:
    - task: port-check
      port: 8080
```

**Parameters:**
- `port` (required): Port number to check

## Environment Tasks

### env-set

Sets an environment variable in a .env file.

**Configuration:**
```yaml
hooks:
  post-install:
    - task: env-set
      file: ".env"
      key: "DATABASE_URL"
      value: "postgres://localhost:5432/mydb"
```

**Parameters:**
- `file` (required): .env file path
- `key` (required): Environment variable name
- `value` (required): Environment variable value

### env-check

Validates that required environment variables are set.

**Configuration:**
```yaml
hooks:
  pre-deploy:
    - task: env-check
      vars: ["DATABASE_URL", "API_KEY", "SECRET_KEY"]
```

**Parameters:**
- `vars` (required): List of required environment variable names

## System Operation Tasks

### wait-for-service

Waits for a service to become available (HTTP or TCP).

**Configuration:**
```yaml
hooks:
  post-deploy:
    # HTTP service
    - task: wait-for-service
      url: "http://localhost:8080/health"
      timeout: 60
      interval: 2
    # TCP port
    - task: wait-for-service
      host: "localhost"
      port: 5432
      timeout: 30
```

**Parameters:**
- `url` (optional): HTTP/HTTPS URL to check
- `host` (optional): TCP host to check
- `port` (optional): TCP port to check
- `timeout` (optional): Timeout in seconds (default: 60)
- `interval` (optional): Check interval in seconds (default: 2)

**Note:** Must specify either `url` or `host:port`, not both.

### notify

Sends a notification (log or webhook).

**Configuration:**
```yaml
hooks:
  post-deploy:
    # Log notification
    - task: notify
      type: "log"
      message: "Deployment completed successfully"
      level: "info"
    # Webhook notification
    - task: notify
      type: "webhook"
      message: "Deployment complete"
      url: "https://hooks.slack.com/services/XXX"
      headers:
        Content-Type: "application/json"
```

**Parameters:**
- `type` (required): "log" or "webhook"
- `message` (required): Notification message
- `level` (optional): Log level for log type ("info", "warn", "error")
- `url` (required for webhook): Webhook URL
- `headers` (optional for webhook): HTTP headers

## Inertia.js Tasks

These tasks are specific to Inertia.js integration. See the Inertia integration documentation for details.

- `setup-inertia-middleware`
- `add-inertia-handlers`
- `add-shared-data`
- `generate-typescript-types`
- `update-routes-for-inertia`

## Using Tasks in Rituals

Tasks are defined in the `hooks` section of `ritual.yaml`:

```yaml
name: my-ritual
version: 1.0.0

hooks:
  pre-install:
    - task: validate-go-version
      min_version: "1.21.0"
    - task: validate-dependencies
      commands: ["git", "docker"]
  
  post-install:
    - task: go-mod-tidy
    - task: env-set
      file: ".env"
      key: "APP_ENV"
      value: "development"
  
  post-deploy:
    - task: wait-for-service
      url: "http://localhost:8080/health"
      timeout: 60
    - task: notify
      type: "log"
      message: "Application deployed successfully"
      level: "info"
```

## Task Execution Context

All tasks receive a `TaskContext` with:
- **Working Directory**: The project directory where tasks execute
- **Variables**: Access to ritual variables and answers
- **Environment**: Environment variables from the ritual

Tasks are executed in the order they appear in the hook definition.

## Error Handling

If a task fails:
1. Execution stops immediately
2. Subsequent tasks in the hook are skipped
3. The error is reported to the user
4. For update/deploy operations, rollback may be triggered

Use the `--dry-run` flag to preview task execution without making changes.

## Creating Custom Tasks

To create custom tasks:

1. Implement the `Task` interface:
   ```go
   type Task interface {
       Name() string
       Validate() error
       Execute(ctx context.Context, taskCtx *TaskContext) error
   }
   ```

2. Register the task in `init()`:
   ```go
   func init() {
       tasks.Register("my-task", func(config map[string]interface{}) (tasks.Task, error) {
           // Parse config and return task instance
       })
   }
   ```

3. Import the package to register the task

See existing task implementations in `internal/hooks/tasks/` for examples.

# Deployment Management Guide

This guide covers Ritual Grove's deployment management features, including history tracking and protected file management.

## Deployment History

Ritual Grove automatically tracks all deployment attempts (installations, updates, rollbacks) in `.ritual/history.yaml`.

### What Gets Tracked

Each deployment record includes:
- **Timestamp**: When the deployment occurred
- **From/To Versions**: Version transition (e.g., "1.0.0" → "1.1.0")
- **Status**: `success`, `failure`, or `rollback`
- **Message**: Human-readable deployment summary
- **Errors**: List of errors if deployment failed
- **Warnings**: Non-fatal warnings
- **Duration**: How long the deployment took

### Example History File

```yaml
deployments:
  - timestamp: 2026-01-07T10:30:00Z
    from_version: ""
    to_version: "1.0.0"
    status: success
    message: "Initial installation"
    duration: "2.5s"
    
  - timestamp: 2026-01-07T14:20:00Z
    from_version: "1.0.0"
    to_version: "1.1.0"
    status: success
    message: "Updated to v1.1.0"
    duration: "1.8s"
    
  - timestamp: 2026-01-07T15:00:00Z
    from_version: "1.1.0"
    to_version: "1.2.0"
    status: failure
    message: "Migration failed"
    errors:
      - "database connection timeout"
      - "failed to apply migration v1.2.0"
```

### Using the History API

```go
import "github.com/toutaio/toutago-ritual-grove/internal/storage"

// Load or create history
history := storage.LoadOrCreateHistory("/path/to/project")

// Add a successful deployment
history.AddDeployment(storage.DeploymentRecord{
    Timestamp:   time.Now(),
    FromVersion: "1.0.0",
    ToVersion:   "1.1.0",
    Status:      "success",
    Message:     "Successfully updated",
    Duration:    "2.1s",
})

// Add a failed deployment
history.AddDeployment(storage.DeploymentRecord{
    Timestamp:   time.Now(),
    FromVersion: "1.1.0",
    ToVersion:   "1.2.0",
    Status:      "failure",
    Message:     "Migration failed",
    Errors:      []string{"database error", "rollback completed"},
})

// Save history
if err := history.Save("/path/to/project"); err != nil {
    log.Fatalf("Failed to save history: %v", err)
}

// Query history
latest := history.GetLatestSuccessful()
if latest != nil {
    fmt.Printf("Last successful deployment: %s\n", latest.ToVersion)
}

failures := history.GetFailures()
fmt.Printf("Total failures: %d\n", len(failures))

rollbacks := history.GetRollbacks()
fmt.Printf("Total rollbacks: %d\n", len(rollbacks))
```

### History Size Limits

History is automatically limited to the last **100 deployments** to prevent unbounded growth. Older entries are automatically removed when new ones are added.

## Protected Files

Protected files are never overwritten during ritual updates, ensuring user customizations and secrets remain safe.

### Defining Protected Files

#### Method 1: Ritual Definition

In your `ritual.yaml`:

```yaml
files:
  protected:
    - config/secrets.env
    - config/database.yaml
    - "*.local.yaml"
    - "config/*.custom.yaml"
```

#### Method 2: User-Defined List

Create `.ritual/protected.txt` in your project:

```text
# Protected files - do not overwrite during updates
# Supports glob patterns like *.env or config/*.yaml

config/secrets.env
config/database.yaml
*.local.yaml
config/*.custom.yaml
my-custom-handler.go
```

### Pattern Matching

Protected files support glob patterns:

- `*.env` - Matches all `.env` files
- `config/*.yaml` - Matches all YAML files in config/
- `secrets.*` - Matches files starting with "secrets."
- `config/*/database.yaml` - Matches database.yaml in any subdirectory of config/

### Using the Protected File API

```go
import "github.com/toutaio/toutago-ritual-grove/internal/storage"

// Load state
state, _ := storage.LoadState("/path/to/project")

// Create protected file manager
pm := storage.NewProtectedFileManager(state)

// Check if file is protected
if pm.IsProtected("config/secrets.env") {
    fmt.Println("File is protected - will not overwrite")
}

// Add a protected file
pm.AddProtectedFile("my-config.yaml")

// Remove from protected list
pm.RemoveProtectedFile("old-config.yaml")

// Load user-defined protected files
userFiles, err := pm.LoadUserProtectedFiles("/path/to/project")
if err != nil {
    log.Printf("Failed to load user protected files: %v", err)
}

// Get all protected files
allProtected := pm.GetAllProtectedFiles()
for _, file := range allProtected {
    fmt.Printf("Protected: %s\n", file)
}

// Save protected list to .ritual/protected.txt
if err := pm.SaveProtectedList("/path/to/project"); err != nil {
    log.Fatalf("Failed to save protected list: %v", err)
}
```

### Update Behavior with Protected Files

When updating a ritual:

1. **Protected files are never overwritten**
   - If a file is protected and exists, it's skipped
   - The update shows what *would* change but doesn't apply it

2. **Diff is shown for protected files**
   - You can see what changed in the ritual
   - Manual merge is your responsibility

3. **New protected files are created**
   - If a protected file doesn't exist yet, it's created normally
   - Protection only applies to existing files

### Best Practices

1. **Protect configuration files**
   ```text
   config/*.yaml
   config/*.env
   .env
   .env.local
   ```

2. **Protect customized code**
   ```text
   internal/custom/*
   handlers/custom_*.go
   ```

3. **Protect secrets**
   ```text
   *.key
   *.pem
   secrets.*
   *secret*
   ```

4. **Document protected files**
   - Add comments in `.ritual/protected.txt`
   - Document why files are protected

5. **Review diffs carefully**
   - When updating, review what would change
   - Manually merge important updates

## Combining History and Protected Files

```go
// Complete deployment workflow
func UpdateRitual(projectPath, newVersion string) error {
    // Load state and history
    state, _ := storage.LoadState(projectPath)
    history := storage.LoadOrCreateHistory(projectPath)
    pm := storage.NewProtectedFileManager(state)
    
    // Load user-defined protected files
    _, _ = pm.LoadUserProtectedFiles(projectPath)
    
    startTime := time.Now()
    record := storage.DeploymentRecord{
        Timestamp:   startTime,
        FromVersion: state.RitualVersion,
        ToVersion:   newVersion,
    }
    
    // Perform update (simplified)
    err := performUpdate(projectPath, newVersion, pm)
    
    record.Duration = time.Since(startTime).String()
    
    if err != nil {
        record.Status = "failure"
        record.Message = "Update failed"
        record.Errors = []string{err.Error()}
    } else {
        record.Status = "success"
        record.Message = fmt.Sprintf("Successfully updated to %s", newVersion)
        state.RitualVersion = newVersion
        state.UpdatedAt = time.Now()
    }
    
    // Save everything
    history.AddDeployment(record)
    _ = history.Save(projectPath)
    _ = state.Save(projectPath)
    
    return err
}
```

## CLI Integration (Future)

Planned CLI commands:

```bash
# View deployment history
touta ritual history

# View latest deployment
touta ritual history --latest

# View failures only
touta ritual history --failures

# Manage protected files
touta ritual protect add config/secrets.env
touta ritual protect remove old-file.txt
touta ritual protect list

# Dry-run update showing protected file diffs
touta ritual update --dry-run --show-protected
```

## File Locations

- **History**: `.ritual/history.yaml`
- **Protected List**: `.ritual/protected.txt`
- **State**: `.ritual/state.yaml`

All files are created automatically and should not be committed to version control (add to `.gitignore`).

## Troubleshooting

### History file corrupted

```bash
# Backup and recreate
mv .ritual/history.yaml .ritual/history.yaml.bak
# History will be recreated on next operation
```

### Protected file overwritten

```bash
# Check if it was actually protected
cat .ritual/protected.txt

# Restore from backup if available
# (Ritual Grove creates automatic backups before updates)
```

### Pattern not matching

```go
// Test pattern matching
pm := storage.NewProtectedFileManager(state)
pm.AddProtectedFile("*.env")

// Both should return true
pm.IsProtected("secrets.env")
pm.IsProtected("config/database.env")
```

## Related Documentation

- [Update System](update-system.md)
- [State Management](state-management.md)
- [Rollback Guide](rollback-guide.md)

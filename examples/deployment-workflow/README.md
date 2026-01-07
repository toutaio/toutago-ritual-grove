# Deployment Workflow Example

This example demonstrates how to use ritual-grove for production deployments, including updates, rollbacks, and migrations.

## Scenario

You have a blog application in production running ritual version 1.0.0. You want to:
1. Update to version 1.1.0 (adds comment feature)
2. Handle any issues by rolling back if needed
3. Understand the complete deployment workflow

## Prerequisites

- A project created with the blog ritual v1.0.0
- Ritual-grove CLI installed
- Basic understanding of rituals

## Workflow Overview

```
┌─────────────────┐
│  Production     │
│  Blog v1.0.0    │
└────────┬────────┘
         │
         ▼
    [Plan Update]
         │
         ▼
   [Review Changes]
         │
         ├─── Conflicts? ───► [Resolve Manually]
         │                            │
         ▼                            ▼
  [Create Backup] ◄──────────────────┘
         │
         ▼
   [Run Update]
         │
         ├─── Success? ───► [Test & Monitor]
         │                            │
         ▼                            ▼
   [Rollback]                    [Complete]
         │
         ▼
  [Restore Backup]
```

## Step-by-Step Guide

### 1. Check Current State

```bash
cd /path/to/your/blog

# Verify current ritual version
cat .ritual/state.yaml
```

Expected output:
```yaml
ritual_name: blog
ritual_version: 1.0.0
installed_at: 2026-01-05T10:00:00Z
```

### 2. Plan the Update

Before making any changes, see what will happen:

```bash
ritual plan --to-version 1.1.0
```

This shows:
- Files that will be added, modified, or deleted
- Migrations that will run
- Potential conflicts

### 3. Create Manual Backup

While `ritual update` creates automatic backups, you can create a manual one:

```bash
ritual backup create --description "Before adding comments feature"
```

### 4. Run the Update

```bash
ritual update --to-version 1.1.0
```

The update process:
1. Creates automatic backup
2. Applies file changes
3. Runs migrations
4. Executes update hooks
5. Updates ritual state

### 5. Test the Updated Application

```bash
go test ./...
go run main.go
```

### 6. If Issues Occur - Rollback

```bash
ritual rollback
```

This automatically:
- Restores files from latest backup
- Rolls back database migrations
- Reverts ritual state

### 7. Clean Old Backups

```bash
ritual backup clean --keep 5
```

## See Full Documentation

For complete step-by-step instructions with examples, troubleshooting, and best practices, see:

**[Complete Deployment Workflow Guide](DEPLOYMENT_GUIDE.md)**

## Quick Reference

| Command | Purpose |
|---------|---------|
| `ritual plan` | Preview update changes |
| `ritual backup create` | Create manual backup |
| `ritual backup list` | List all backups |
| `ritual update` | Apply update |
| `ritual rollback` | Undo last update |
| `ritual backup restore <path>` | Restore from specific backup |
| `ritual backup clean` | Clean old backups |

## See Also

- [Complete Deployment Guide](DEPLOYMENT_GUIDE.md)
- [CLI Reference](../../docs/cli-reference.md)
- [Best Practices](../../docs/best-practices.md)

# Complete Deployment Workflow Guide

Comprehensive guide for production deployments using Toutago Ritual Grove.

## Table of Contents

1. [Pre-Deployment](#pre-deployment)
2. [Deployment Process](#deployment-process)
3. [Post-Deployment](#post-deployment)
4. [Rollback Procedures](#rollback-procedures)
5. [Migration Patterns](#migration-patterns)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)

## Pre-Deployment

### 1. Review the Update Plan

```bash
ritual plan --to-version 1.1.0
```

**Review checklist:**
- [ ] What files will change?
- [ ] Are there database migrations?
- [ ] Any breaking changes?
- [ ] Conflicts to resolve?
- [ ] Estimated downtime?

### 2. Test in Staging

```bash
# In staging environment
ritual update --to-version 1.1.0

# Run full test suite
go test ./...

# Manual testing
# - Test new features
# - Test existing features
# - Test edge cases

# Load testing (if needed)
# Use tools like k6, wrk, or ab
```

### 3. Prepare Rollback Plan

Document:
- Current version: v1.0.0
- Target version: v1.1.0
- Rollback command: `ritual rollback`
- Estimated rollback time: ~2 minutes
- Database rollback: Automatic (down migrations)

### 4. Schedule Maintenance Window

For production deployments:
- Schedule during low-traffic hours
- Notify users of maintenance
- Prepare monitoring dashboard
- Have team available for issues

## Deployment Process

### Step 1: Create Pre-Deployment Backup

```bash
ritual backup create --description "Pre-deployment v1.1.0 - $(date)"
```

### Step 2: Verify Current State

```bash
# Check ritual version
cat .ritual/state.yaml

# Check application health
curl http://localhost:8080/health

# Check database connection
psql -c "SELECT 1"
```

### Step 3: Put Application in Maintenance Mode (Optional)

```bash
# Option 1: Stop application
systemctl stop my-blog

# Option 2: Enable maintenance page
# Configure your load balancer or reverse proxy
```

### Step 4: Execute Update

```bash
ritual update --to-version 1.1.0 2>&1 | tee deployment-$(date +%Y%m%d-%H%M%S).log
```

Save the output for debugging if needed.

### Step 5: Resolve Conflicts (If Any)

```bash
# Edit conflicting files
vim main.go

# Look for conflict markers
<<<<<<< LOCAL
=======
>>>>>>> RITUAL

# After resolving
go test ./...
```

### Step 6: Run Post-Deployment Tests

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Smoke tests
./scripts/smoke-test.sh
```

### Step 7: Start Application

```bash
# Start service
systemctl start my-blog

# Check logs
journalctl -u my-blog -f

# Verify health
curl http://localhost:8080/health
```

## Post-Deployment

### Immediate Checks (First 5 Minutes)

```bash
# 1. Application responding?
curl http://localhost:8080

# 2. Database accessible?
psql -c "SELECT COUNT(*) FROM posts"

# 3. New features working?
curl http://localhost:8080/posts/1/comments

# 4. Error logs clean?
tail -n 100 /var/log/app.log | grep ERROR

# 5. Performance acceptable?
# Check response times in monitoring dashboard
```

### Extended Monitoring (First Hour)

- Monitor error rates
- Check memory usage
- Watch database connection pool
- Review application logs
- Monitor user feedback

### Documentation

Update deployment log:
```markdown
# Deployment Log

## 2026-01-07 17:00 - Blog v1.1.0

- **Performed by:** John Doe
- **Duration:** 5 minutes
- **Downtime:** 2 minutes
- **Issues:** None
- **Rollback:** Not required
- **Status:** Success ✓

### Changes
- Added comments feature
- Updated database schema
- Modified post display

### Metrics
- Response time: 45ms (baseline: 42ms)
- Error rate: 0.01% (baseline: 0.01%)
- Database connections: 12/100
```

## Rollback Procedures

### Automatic Rollback

```bash
ritual rollback
```

This automatically:
1. Identifies previous version
2. Restores files from backup
3. Rolls back database migrations
4. Updates ritual state

### Manual Rollback from Specific Backup

```bash
# List backups
ritual backup list

# Choose backup
ritual backup restore .ritual/backups/backup-2026-01-07-093000.tar.gz --force
```

### Database-Only Rollback

```bash
# Run down migration manually
ritual migrate down

# Or restore database from dump
psql my_database < backup.sql
```

### Rollback Decision Matrix

| Scenario | Action | Command |
|----------|--------|---------|
| App won't start | Rollback | `ritual rollback` |
| High error rate (>5%) | Rollback | `ritual rollback` |
| Data corruption | Rollback + DB restore | `ritual rollback && psql < backup.sql` |
| Minor bug | Fix forward | Deploy hotfix |
| Performance degradation | Monitor, then decide | Wait 15 min or rollback |

## Migration Patterns

### Pattern 1: Add Column (Safe)

```yaml
up:
  sql:
    - "ALTER TABLE users ADD COLUMN email VARCHAR(255)"
down:
  sql:
    - "ALTER TABLE users DROP COLUMN email"
```

### Pattern 2: Rename Column (Data Preservation)

```yaml
up:
  sql:
    - "ALTER TABLE users RENAME COLUMN old_name TO new_name"
down:
  sql:
    - "ALTER TABLE users RENAME COLUMN new_name TO old_name"
```

### Pattern 3: Add Table with Foreign Key

```yaml
up:
  sql:
    - "CREATE TABLE comments (id SERIAL PRIMARY KEY, post_id INT REFERENCES posts(id))"
down:
  sql:
    - "DROP TABLE comments"
```

### Pattern 4: Data Migration

```yaml
up:
  sql:
    - "CREATE TABLE users_new AS TABLE users"
    - "ALTER TABLE users_new ADD COLUMN verified BOOLEAN DEFAULT FALSE"
    - "UPDATE users_new SET verified = TRUE WHERE status = 'active'"
    - "DROP TABLE users"
    - "ALTER TABLE users_new RENAME TO users"
down:
  sql:
    - "CREATE TABLE users_old AS SELECT * FROM users"
    - "ALTER TABLE users_old DROP COLUMN verified"
    - "DROP TABLE users"
    - "ALTER TABLE users_old RENAME TO users"
```

## Best Practices

### 1. Always Use Version Control

```bash
# Before deployment
git status
git log -1

# Tag releases
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0
```

### 2. Automate Testing

```bash
#!/bin/bash
# scripts/pre-deploy-test.sh

set -e

echo "Running pre-deployment tests..."

# Unit tests
go test ./...

# Linting
golangci-lint run

# Build check
go build

# Migration validation
ritual validate

echo "✓ All pre-deployment checks passed"
```

### 3. Use Blue-Green Deployment (Advanced)

```bash
# Deploy to green (new version)
ritual update --to-version 1.1.0

# Test green environment
./scripts/smoke-test.sh green

# Switch traffic from blue to green
# (Use your load balancer)

# Keep blue running for quick rollback
# Stop blue after monitoring period
```

### 4. Implement Health Checks

```go
// handlers/health.go
func HealthCheck(w http.ResponseWriter, r *http.Request) {
    // Check database
    err := db.Ping()
    if err != nil {
        http.Error(w, "Database unhealthy", 503)
        return
    }
    
    // Check dependencies
    // ...
    
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
        "version": "1.1.0",
    })
}
```

### 5. Monitor Metrics

Key metrics to track:
- Response time (p50, p95, p99)
- Error rate
- Request rate
- Database query time
- Memory usage
- CPU usage

## Troubleshooting

### Issue: Migration Failed

```
Error: migration failed: relation "users" does not exist
```

**Solution:**
```bash
# Check database state
psql -c "\d users"

# Check migration history
ritual migrate status

# Manual intervention may be needed
# Fix database manually, then:
ritual rollback
```

### Issue: File Conflicts

```
⚠ Conflict in main.go
```

**Solution:**
```bash
# 1. View the conflict
cat main.go | grep -A5 -B5 "<<<<<<<"

# 2. Edit and resolve
vim main.go

# 3. Test
go build && go test

# 4. Continue
```

### Issue: Application Won't Start After Update

```bash
# Check logs
journalctl -u my-blog -n 50

# Common causes:
# - Missing dependency
# - Configuration error
# - Database migration issue

# Quick fix: Rollback
ritual rollback
```

### Issue: High Memory Usage After Deployment

```bash
# Monitor memory
watch -n 1 'ps aux | grep my-app'

# If increasing rapidly:
# 1. Check for memory leaks in new code
# 2. Profile: pprof
# 3. Consider rollback if critical

# If stable but higher:
# New features may need more memory
# Monitor for 24h before deciding
```

## Emergency Procedures

### Complete System Failure

```bash
# 1. Stop application
systemctl stop my-blog

# 2. Restore from backup
ritual backup list
ritual backup restore <latest-good-backup> --force

# 3. Restore database
psql my_database < /backups/db-backup-last-good.sql

# 4. Verify files
ls -la

# 5. Restart
systemctl start my-blog

# 6. Verify
curl http://localhost:8080/health
```

### Data Corruption Detected

```bash
# 1. STOP ALL WRITES IMMEDIATELY
systemctl stop my-blog

# 2. Assess damage
psql -c "SELECT COUNT(*) FROM posts"
psql -c "SELECT * FROM posts ORDER BY id DESC LIMIT 10"

# 3. Restore from last good backup
ritual backup restore <last-good-backup>
psql my_database < /backups/db-before-corruption.sql

# 4. Root cause analysis
# - Check migration logs
# - Check application logs
# - Review recent code changes

# 5. Fix and redeploy
# After fix is verified
```

## Summary

Key takeaways:
1. **Always plan first** - `ritual plan`
2. **Test in staging** - Catch issues early
3. **Create backups** - Before every deployment
4. **Monitor closely** - First hour is critical
5. **Rollback quickly** - Don't hesitate if issues arise
6. **Document everything** - Learn from each deployment

## See Also

- [CLI Reference](../../docs/cli-reference.md)
- [Best Practices Guide](../../docs/best-practices.md)
- [Deployment Management](../../docs/deployment-management.md)

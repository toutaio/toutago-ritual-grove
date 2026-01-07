# Cache Management

## Overview

Toutago Ritual Grove caches extracted rituals in `~/.toutago/ritual-cache/` to improve performance. This document explains how to manage the cache.

## Cache Directory Structure

```
~/.toutago/ritual-cache/
├── embedded/          # Cached embedded rituals from binary
│   ├── hello-world/
│   ├── minimal/
│   ├── basic-site/
│   └── blog/
└── git/              # Cached rituals from git repositories
```

## Commands

### Clear Embedded Cache

This is the most common operation. Use it when you've rebuilt the ritual binary with updated rituals:

```bash
ritual clean
# or explicitly:
ritual clean --embedded
```

This will:
- Remove all cached embedded rituals
- Force re-extraction on next scan
- Preserve git-cloned rituals

### Clear All Cache

To remove all cached rituals (embedded and git):

```bash
ritual clean --all
```

This will:
- Show current cache size
- Remove entire cache directory
- Free up disk space
- Recreate empty cache directory

## Automatic Version Checking

The registry automatically detects when embedded rituals have changed versions and re-extracts them. This means:

1. You rebuild the binary with updated ritual versions
2. Next time you run `ritual list` or `ritual create`
3. The system detects version mismatch
4. Outdated rituals are automatically re-extracted

## Manual Cache Location

If you need to access the cache manually:

```bash
# View cache location
ls ~/.toutago/ritual-cache/

# Check cache size
du -sh ~/.toutago/ritual-cache/

# Manually remove cache
rm -rf ~/.toutago/ritual-cache/
```

## Troubleshooting

### Old Rituals Persist After Rebuild

If you're still seeing old ritual versions after rebuilding:

1. Run `ritual clean --embedded`
2. Verify with `ritual list`
3. If issue persists, use `ritual clean --all`

### Cache Growing Too Large

Check cache size:

```bash
du -sh ~/.toutago/ritual-cache/
```

Clean old git clones and re-download:

```bash
ritual clean --all
```

## Best Practices

1. **After Development**: Run `ritual clean` after modifying embedded rituals
2. **Disk Space**: Periodically run `ritual clean --all` to free space
3. **CI/CD**: Cache directory can be safely deleted in CI environments
4. **Debugging**: Check `~/.toutago/ritual-cache/embedded/` to see cached versions

## Environment Variables

The cache directory location can be overridden (advanced):

```bash
# Not recommended - for testing only
export TOUTA_CACHE_DIR=/tmp/ritual-cache
```

## Cache Performance

- **First scan**: ~100-200ms to extract embedded rituals
- **Subsequent scans**: ~10-20ms (using cache)
- **With version check**: ~50ms (loads manifest to compare versions)

The automatic version checking adds minimal overhead while ensuring you always have the latest ritual versions.

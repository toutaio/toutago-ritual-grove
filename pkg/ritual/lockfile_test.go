package ritual

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockFile_Load(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *LockFile
		wantErr bool
	}{
		{
			name: "valid lock file",
			content: `ritual:
  name: test-ritual
  version: 1.0.0
  resolved_at: 2024-01-01T00:00:00Z

dependencies:
  - name: github.com/example/dep1
    version: 1.2.3
    resolved: 1.2.3
    checksum: abc123
  - name: github.com/example/dep2
    version: ^2.0.0
    resolved: 2.1.0
    checksum: def456

rituals:
  - name: base-ritual
    version: 1.0.0
    source: https://github.com/example/base-ritual
    checksum: xyz789
`,
			want: &LockFile{
				Ritual: RitualLock{
					Name:       "test-ritual",
					Version:    "1.0.0",
					ResolvedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				Dependencies: []DependencyLock{
					{
						Name:     "github.com/example/dep1",
						Version:  "1.2.3",
						Resolved: "1.2.3",
						Checksum: "abc123",
					},
					{
						Name:     "github.com/example/dep2",
						Version:  "^2.0.0",
						Resolved: "2.1.0",
						Checksum: "def456",
					},
				},
				Rituals: []RitualDependencyLock{
					{
						Name:     "base-ritual",
						Version:  "1.0.0",
						Source:   "https://github.com/example/base-ritual",
						Checksum: "xyz789",
					},
				},
			},
		},
		{
			name: "minimal lock file",
			content: `ritual:
  name: minimal
  version: 0.1.0
  resolved_at: 2024-01-01T00:00:00Z
`,
			want: &LockFile{
				Ritual: RitualLock{
					Name:       "minimal",
					Version:    "0.1.0",
					ResolvedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name:    "invalid yaml",
			content: `invalid: [yaml`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			lockPath := filepath.Join(tmpDir, "ritual.lock")

			err := os.WriteFile(lockPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			got, err := LoadLockFile(lockPath)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Ritual.Name, got.Ritual.Name)
			assert.Equal(t, tt.want.Ritual.Version, got.Ritual.Version)
			assert.Equal(t, len(tt.want.Dependencies), len(got.Dependencies))
			assert.Equal(t, len(tt.want.Rituals), len(got.Rituals))
		})
	}
}

func TestLockFile_Save(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "ritual.lock")

	lock := &LockFile{
		Ritual: RitualLock{
			Name:       "test",
			Version:    "1.0.0",
			ResolvedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		Dependencies: []DependencyLock{
			{
				Name:     "github.com/example/dep",
				Version:  "1.0.0",
				Resolved: "1.0.0",
				Checksum: "abc123",
			},
		},
	}

	err := lock.Save(lockPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(lockPath)
	require.NoError(t, err)

	// Load and verify
	loaded, err := LoadLockFile(lockPath)
	require.NoError(t, err)
	assert.Equal(t, lock.Ritual.Name, loaded.Ritual.Name)
	assert.Equal(t, lock.Ritual.Version, loaded.Ritual.Version)
}

func TestLockFile_Verify(t *testing.T) {
	tests := []struct {
		name     string
		lock     *LockFile
		manifest *Manifest
		wantErr  bool
	}{
		{
			name: "matching versions",
			lock: &LockFile{
				Ritual: RitualLock{
					Name:    "test",
					Version: "1.0.0",
				},
			},
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name:    "test",
					Version: "1.0.0",
				},
			},
			wantErr: false,
		},
		{
			name: "version mismatch",
			lock: &LockFile{
				Ritual: RitualLock{
					Name:    "test",
					Version: "1.0.0",
				},
			},
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name:    "test",
					Version: "2.0.0",
				},
			},
			wantErr: true,
		},
		{
			name: "name mismatch",
			lock: &LockFile{
				Ritual: RitualLock{
					Name:    "test1",
					Version: "1.0.0",
				},
			},
			manifest: &Manifest{
				Ritual: RitualMeta{
					Name:    "test2",
					Version: "1.0.0",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.lock.Verify(tt.manifest)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLockFile_DependencyCheck(t *testing.T) {
	lock := &LockFile{
		Dependencies: []DependencyLock{
			{Name: "dep1", Version: "1.0.0", Resolved: "1.0.0"},
			{Name: "dep2", Version: "2.0.0", Resolved: "2.1.0"},
		},
	}

	dep, found := lock.GetDependency("dep1")
	assert.True(t, found)
	assert.Equal(t, "1.0.0", dep.Resolved)

	_, found = lock.GetDependency("nonexistent")
	assert.False(t, found)
}

func TestLockFile_RitualDependencyCheck(t *testing.T) {
	lock := &LockFile{
		Rituals: []RitualDependencyLock{
			{Name: "ritual1", Version: "1.0.0"},
			{Name: "ritual2", Version: "2.0.0"},
		},
	}

	ritual, found := lock.GetRitualDependency("ritual1")
	assert.True(t, found)
	assert.Equal(t, "1.0.0", ritual.Version)

	_, found = lock.GetRitualDependency("nonexistent")
	assert.False(t, found)
}

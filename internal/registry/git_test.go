package registry

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGenerateCacheName(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "https url",
			url:  "https://github.com/user/repo.git",
			want: "github_com_user_repo",
		},
		{
			name: "http url",
			url:  "http://gitlab.com/user/project.git",
			want: "gitlab_com_user_project",
		},
		{
			name: "ssh url",
			url:  "git@github.com:user/repo.git",
			want: "github_com_user_repo",
		},
		{
			name: "url without .git",
			url:  "https://github.com/user/repo",
			want: "github_com_user_repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateCacheName(tt.url)
			if got != tt.want {
				t.Errorf("generateCacheName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitSource(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ritual-git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test Git repository
	testRepoDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(testRepoDir, 0750); err != nil {
		t.Fatalf("Failed to create test repo dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git (required for commits)
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = testRepoDir
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = testRepoDir
	_ = cmd.Run()

	// Create ritual.yaml
	ritualContent := `ritual:
  name: test-git-ritual
  version: 1.0.0
  description: Test ritual from Git
  author: Test Author
  tags:
    - test
    - git

compatibility:
  min_touta_version: 0.1.0
  max_touta_version: 1.0.0

dependencies:
  go_packages: []
  other_rituals: []

questions: []

files:
  templates: []
  static: []
  protected: []
`
	ritualPath := filepath.Join(testRepoDir, "ritual.yaml")
	if err := os.WriteFile(ritualPath, []byte(ritualContent), 0600); err != nil {
		t.Fatalf("Failed to write ritual.yaml: %v", err)
	}

	// Commit the file
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	// Test cloning
	cacheDir := filepath.Join(tmpDir, "cache")
	registry := &Registry{
		rituals:    make(map[string]*RitualMetadata),
		cacheDir: cacheDir,
	}

	source := &GitSource{
		URL: testRepoDir, // Use local path as URL for testing
	}

	clonePath, err := registry.CloneGitRitual(source)
	if err != nil {
		t.Fatalf("Failed to clone git ritual: %v", err)
	}

	// Verify clone path exists
	if _, err := os.Stat(clonePath); os.IsNotExist(err) {
		t.Errorf("Clone path does not exist: %s", clonePath)
	}

	// Verify ritual.yaml exists in cloned repo
	clonedRitualPath := filepath.Join(clonePath, "ritual.yaml")
	if _, err := os.Stat(clonedRitualPath); os.IsNotExist(err) {
		t.Errorf("ritual.yaml not found in cloned repo")
	}

	// Test scanning the cloned repo
	if err := registry.ScanGitRepo(clonePath); err != nil {
		t.Fatalf("Failed to scan git repo: %v", err)
	}

	// Verify ritual was indexed
	meta, err := registry.Get("test-git-ritual")
	if err != nil {
		t.Errorf("Failed to get indexed ritual: %v", err)
	}

	if meta.Source != SourceGit {
		t.Errorf("Expected source to be SourceGit, got %v", meta.Source)
	}

	// Test cloning again (should use existing clone)
	clonePath2, err := registry.CloneGitRitual(source)
	if err != nil {
		t.Fatalf("Failed to clone git ritual second time: %v", err)
	}

	if clonePath != clonePath2 {
		t.Errorf("Expected same clone path, got different paths")
	}
}

func TestGitSourceWithBranch(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir, err := os.MkdirTemp("", "ritual-git-branch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test repo with branches
	testRepoDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(testRepoDir, 0750); err != nil {
		t.Fatalf("Failed to create test repo dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = testRepoDir
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = testRepoDir
	_ = cmd.Run()

	// Create ritual.yaml on main branch
	ritualContent := `ritual:
  name: test-branch-ritual
  version: 1.0.0
  description: Main branch ritual
`
	if err := os.WriteFile(filepath.Join(testRepoDir, "ritual.yaml"), []byte(ritualContent), 0600); err != nil {
		t.Fatalf("Failed to write ritual.yaml: %v", err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = testRepoDir
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Main commit")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create a dev branch
	cmd = exec.Command("git", "checkout", "-b", "dev")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create dev branch: %v", err)
	}

	// Modify ritual.yaml on dev branch
	ritualContent = `ritual:
  name: test-branch-ritual
  version: 2.0.0
  description: Dev branch ritual
`
	if err := os.WriteFile(filepath.Join(testRepoDir, "ritual.yaml"), []byte(ritualContent), 0600); err != nil {
		t.Fatalf("Failed to write ritual.yaml: %v", err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = testRepoDir
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Dev commit")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Test cloning dev branch
	cacheDir := filepath.Join(tmpDir, "cache")
	registry := &Registry{
		rituals:    make(map[string]*RitualMetadata),
		cacheDir: cacheDir,
	}

	source := &GitSource{
		URL:    testRepoDir,
		Branch: "dev",
	}

	clonePath, err := registry.CloneGitRitual(source)
	if err != nil {
		t.Fatalf("Failed to clone git ritual: %v", err)
	}

	// Scan and verify it's the dev version
	if err := registry.ScanGitRepo(clonePath); err != nil {
		t.Fatalf("Failed to scan git repo: %v", err)
	}

	meta, err := registry.Get("test-branch-ritual")
	if err != nil {
		t.Fatalf("Failed to get ritual: %v", err)
	}

	if meta.Version != "2.0.0" {
		t.Errorf("Expected version 2.0.0, got %s", meta.Version)
	}
}

func TestScanGitRepoMonoRepo(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir, err := os.MkdirTemp("", "ritual-git-monorepo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mono-repo with multiple rituals
	testRepoDir := filepath.Join(tmpDir, "monorepo")
	if err := os.MkdirAll(testRepoDir, 0750); err != nil {
		t.Fatalf("Failed to create test repo dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Create two rituals
	ritual1Dir := filepath.Join(testRepoDir, "ritual1")
	if err := os.MkdirAll(ritual1Dir, 0750); err != nil {
		t.Fatalf("Failed to create ritual1 dir: %v", err)
	}

	ritual1Content := `ritual:
  name: ritual-one
  version: 1.0.0
  description: First ritual in monorepo
`
	if err := os.WriteFile(filepath.Join(ritual1Dir, "ritual.yaml"), []byte(ritual1Content), 0600); err != nil {
		t.Fatalf("Failed to write ritual1.yaml: %v", err)
	}

	ritual2Dir := filepath.Join(testRepoDir, "ritual2")
	if err := os.MkdirAll(ritual2Dir, 0750); err != nil {
		t.Fatalf("Failed to create ritual2 dir: %v", err)
	}

	ritual2Content := `ritual:
  name: ritual-two
  version: 1.0.0
  description: Second ritual in monorepo
`
	if err := os.WriteFile(filepath.Join(ritual2Dir, "ritual.yaml"), []byte(ritual2Content), 0600); err != nil {
		t.Fatalf("Failed to write ritual2.yaml: %v", err)
	}

	// Scan the mono-repo
	registry := &Registry{
		rituals:    make(map[string]*RitualMetadata),
		cacheDir: tmpDir,
	}

	if err := registry.ScanGitRepo(testRepoDir); err != nil {
		t.Fatalf("Failed to scan monorepo: %v", err)
	}

	// Verify both rituals were indexed
	meta1, err := registry.Get("ritual-one")
	if err != nil {
		t.Errorf("Failed to get ritual-one: %v", err)
	} else if meta1.Source != SourceGit {
		t.Errorf("Expected SourceGit, got %v", meta1.Source)
	}

	meta2, err := registry.Get("ritual-two")
	if err != nil {
		t.Errorf("Failed to get ritual-two: %v", err)
	} else if meta2.Source != SourceGit {
		t.Errorf("Expected SourceGit, got %v", meta2.Source)
	}
}

func TestLoadFromGit(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir, err := os.MkdirTemp("", "ritual-load-git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test repo
	testRepoDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(testRepoDir, 0750); err != nil {
		t.Fatalf("Failed to create test repo dir: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = testRepoDir
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = testRepoDir
	_ = cmd.Run()

	ritualContent := `ritual:
  name: load-test-ritual
  version: 1.0.0
  description: Test LoadFromGit
`
	if err := os.WriteFile(filepath.Join(testRepoDir, "ritual.yaml"), []byte(ritualContent), 0600); err != nil {
		t.Fatalf("Failed to write ritual.yaml: %v", err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = testRepoDir
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Test LoadFromGit
	cacheDir := filepath.Join(tmpDir, "cache")
	registry := &Registry{
		rituals:    make(map[string]*RitualMetadata),
		cacheDir: cacheDir,
	}

	source := &GitSource{
		URL: testRepoDir,
	}

	if err := registry.LoadFromGit(source); err != nil {
		t.Fatalf("LoadFromGit failed: %v", err)
	}

	// Verify ritual was loaded
	meta, err := registry.Get("load-test-ritual")
	if err != nil {
		t.Fatalf("Failed to get ritual: %v", err)
	}

	if meta.Name != "load-test-ritual" {
		t.Errorf("Expected name 'load-test-ritual', got %s", meta.Name)
	}

	if meta.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", meta.Version)
	}

	if meta.Source != SourceGit {
		t.Errorf("Expected SourceGit, got %v", meta.Source)
	}
}

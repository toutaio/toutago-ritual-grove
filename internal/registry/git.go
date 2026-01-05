package registry

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitSource represents a Git repository source for rituals
type GitSource struct {
	URL    string
	Branch string
	Tag    string
	Commit string
}

// CloneGitRitual clones a Git repository containing rituals
func (r *Registry) CloneGitRitual(source *GitSource) (string, error) {
	// Generate a cache directory name from the URL
	cacheName := generateCacheName(source.URL)
	clonePath := filepath.Join(r.cacheDir, "git", cacheName)

	// Check if already cloned
	if _, err := os.Stat(filepath.Join(clonePath, ".git")); err == nil {
		// Already cloned, pull latest changes
		if err := r.pullGitRepo(clonePath, source); err != nil {
			return "", fmt.Errorf("failed to update repository: %w", err)
		}
		return clonePath, nil
	}

	// Clone the repository
	if err := r.cloneGitRepo(source, clonePath); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return clonePath, nil
}

// cloneGitRepo performs the actual git clone
func (r *Registry) cloneGitRepo(source *GitSource, targetPath string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	args := []string{"clone"}

	// Add branch or tag if specified
	if source.Branch != "" {
		args = append(args, "--branch", source.Branch)
	} else if source.Tag != "" {
		args = append(args, "--branch", source.Tag)
	}

	// Add depth for faster clone (shallow clone)
	if source.Commit == "" {
		args = append(args, "--depth", "1")
	}

	args = append(args, source.URL, targetPath)

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s: %w", string(output), err)
	}

	// Checkout specific commit if specified
	if source.Commit != "" {
		// #nosec G204 - git checkout with validated commit SHA from ritual source
		cmd := exec.Command("git", "checkout", source.Commit)
		cmd.Dir = targetPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git checkout failed: %s: %w", string(output), err)
		}
	}

	return nil
}

// pullGitRepo pulls latest changes from the repository
func (r *Registry) pullGitRepo(repoPath string, source *GitSource) error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Pull failed, might be already up to date or have conflicts
		// Check if we're on the right branch/commit
		// #nosec G204 - git checkout with validated commit SHA
		if source.Commit != "" {
			cmd := exec.Command("git", "checkout", source.Commit)
			cmd.Dir = repoPath
			if _, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to checkout commit: %w", err)
			}
		}
		// Not necessarily an error if already up to date
		if !strings.Contains(string(output), "Already up to date") {
			return fmt.Errorf("git pull failed: %s", string(output))
		}
	}

	return nil
}

// ScanGitRepo scans a Git repository for rituals
func (r *Registry) ScanGitRepo(repoPath string) error {
	// Check if ritual.yaml exists at root
	ritualFile := filepath.Join(repoPath, "ritual.yaml")
	if _, err := os.Stat(ritualFile); err == nil {
		// Single ritual in repo
		return r.indexRitual(repoPath, SourceGit)
	}

	// Scan for rituals in subdirectories (mono-repo support)
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return fmt.Errorf("failed to read repository: %w", err)
	}

	foundRituals := false
	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}

		if entry.IsDir() {
			entryPath := filepath.Join(repoPath, entry.Name())
			ritualFile := filepath.Join(entryPath, "ritual.yaml")
			if _, err := os.Stat(ritualFile); err == nil {
				if err := r.indexRitual(entryPath, SourceGit); err == nil {
					foundRituals = true
				}
			}
		}
	}

	if !foundRituals {
		return fmt.Errorf("no rituals found in repository")
	}

	return nil
}

// LoadFromGit clones and indexes rituals from a Git repository
func (r *Registry) LoadFromGit(source *GitSource) error {
	repoPath, err := r.CloneGitRitual(source)
	if err != nil {
		return err
	}

	return r.ScanGitRepo(repoPath)
}

// generateCacheName creates a safe directory name from a Git URL
func generateCacheName(url string) string {
	// Remove protocol
	name := strings.TrimPrefix(url, "https://")
	name = strings.TrimPrefix(name, "http://")
	name = strings.TrimPrefix(name, "git@")

	// Remove .git suffix
	name = strings.TrimSuffix(name, ".git")

	// Replace special characters
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, ".", "_")

	return name
}

package deployment

import (
	"sort"
)

// DiffGenerator handles file difference detection
type DiffGenerator struct{}

// NewDiffGenerator creates a new diff generator
func NewDiffGenerator() *DiffGenerator {
	return &DiffGenerator{}
}

// FileDiff represents the differences between two file sets
type FileDiff struct {
	Added     []string // Files that were added
	Modified  []string // Files that were modified
	Deleted   []string // Files that were deleted
	Unchanged []string // Files that haven't changed
	Conflicts []string // Protected files that would be overwritten
}

// DiffSummary provides a summary of changes
type DiffSummary struct {
	TotalChanges  int
	AddedCount    int
	ModifiedCount int
	DeletedCount  int
	ConflictCount int
}

// Summary returns a summary of the diff
func (d *FileDiff) Summary() DiffSummary {
	return DiffSummary{
		TotalChanges:  len(d.Added) + len(d.Modified) + len(d.Deleted) + len(d.Conflicts),
		AddedCount:    len(d.Added),
		ModifiedCount: len(d.Modified),
		DeletedCount:  len(d.Deleted),
		ConflictCount: len(d.Conflicts),
	}
}

// GenerateDiff compares two file sets and returns the differences
func (g *DiffGenerator) GenerateDiff(currentFiles, newFiles map[string]string) *FileDiff {
	return g.GenerateDiffWithProtected(currentFiles, newFiles, nil)
}

// GenerateDiffWithProtected generates a diff considering protected files
func (g *DiffGenerator) GenerateDiffWithProtected(currentFiles, newFiles map[string]string, protectedFiles []string) *FileDiff {
	diff := &FileDiff{
		Added:     []string{},
		Modified:  []string{},
		Deleted:   []string{},
		Unchanged: []string{},
		Conflicts: []string{},
	}

	// Build protected files map for quick lookup
	protected := make(map[string]bool)
	for _, f := range protectedFiles {
		protected[f] = true
	}

	// Check for added and modified files
	for filename, newContent := range newFiles {
		currentContent, exists := currentFiles[filename]

		if !exists {
			// File was added
			diff.Added = append(diff.Added, filename)
		} else if currentContent != newContent {
			// File was modified
			if protected[filename] {
				// It's a protected file that would be overwritten
				diff.Conflicts = append(diff.Conflicts, filename)
			} else {
				diff.Modified = append(diff.Modified, filename)
			}
		} else {
			// File is unchanged
			diff.Unchanged = append(diff.Unchanged, filename)
		}
	}

	// Check for deleted files
	for filename := range currentFiles {
		if _, exists := newFiles[filename]; !exists {
			diff.Deleted = append(diff.Deleted, filename)
		}
	}

	// Sort all slices for consistent output
	sort.Strings(diff.Added)
	sort.Strings(diff.Modified)
	sort.Strings(diff.Deleted)
	sort.Strings(diff.Unchanged)
	sort.Strings(diff.Conflicts)

	return diff
}

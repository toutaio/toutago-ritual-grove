package registry

import "strings"

// FilterOptions specifies criteria for filtering rituals
type FilterOptions struct {
	Tags         []string // Filter by tags (OR logic - matches if any tag matches)
	NamePattern  string   // Filter by name substring (case-insensitive)
	DatabaseType string   // Filter by required database type (postgres, mysql, sqlite)
	Author       string   // Filter by author name
}

// Filter returns rituals matching the filter criteria
func (r *Registry) Filter(opts FilterOptions) []*RitualMetadata {
	var results []*RitualMetadata

	for _, meta := range r.rituals {
		if opts.matches(meta) {
			results = append(results, meta)
		}
	}

	return results
}

// matches checks if a ritual metadata matches the filter options
func (opts FilterOptions) matches(meta *RitualMetadata) bool {
	// Check name pattern
	if !opts.matchesNamePattern(meta.Name) {
		return false
	}

	// Check tags
	if !opts.matchesTags(meta.Tags) {
		return false
	}

	// Check author
	if opts.Author != "" {
		if !strings.Contains(strings.ToLower(meta.Author), strings.ToLower(opts.Author)) {
			return false
		}
	}

	// Database type filtering would require loading the manifest
	// For now, we skip it if specified (could be enhanced later)
	// if opts.DatabaseType != "" {
	//     // Would need to load manifest and check dependencies.database.types
	// }

	return true
}

// matchesNamePattern checks if the ritual name matches the pattern
func (opts FilterOptions) matchesNamePattern(name string) bool {
	if opts.NamePattern == "" {
		return true
	}
	return strings.Contains(strings.ToLower(name), strings.ToLower(opts.NamePattern))
}

// matchesTags checks if the ritual has any of the specified tags
func (opts FilterOptions) matchesTags(tags []string) bool {
	// Empty filter matches all
	if len(opts.Tags) == 0 {
		return true
	}

	// Check if any filter tag matches any ritual tag (OR logic)
	for _, filterTag := range opts.Tags {
		for _, ritualTag := range tags {
			if strings.EqualFold(filterTag, ritualTag) {
				return true
			}
		}
	}

	return false
}

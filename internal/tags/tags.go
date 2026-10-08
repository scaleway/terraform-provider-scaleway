// Package tags provides reusable helpers for managing Scaleway resource tags,
// including provider-level default tags that are automatically merged into
// every tag-compatible resource.
//
// The design mirrors the AWS provider's default_tags feature but adapted for
// Scaleway's native tag model where tags are a list of strings ([]string)
// rather than key-value pairs.
package tags

// DefaultConfig holds the provider-level default tags that are merged into
// every tag-compatible resource.
type DefaultConfig struct {
	Tags []string
}

// MergeTags returns the union of defaultTags and resourceTags.
// Default tags appear first; resource-specific tags that are not already
// present in the default set are appended. Duplicate entries are removed.
//
// If defaultTags is empty, resourceTags is returned as-is.
func MergeTags(defaultTags, resourceTags []string) []string {
	if len(defaultTags) == 0 {
		return resourceTags
	}

	seen := make(map[string]struct{}, len(defaultTags)+len(resourceTags))
	result := make([]string, 0, len(defaultTags)+len(resourceTags))

	for _, t := range defaultTags {
		if _, ok := seen[t]; !ok {
			result = append(result, t)
			seen[t] = struct{}{}
		}
	}

	for _, t := range resourceTags {
		if _, ok := seen[t]; !ok {
			result = append(result, t)
			seen[t] = struct{}{}
		}
	}

	return result
}

// MergeTags returns the result of MergeTags using the DefaultConfig's tags.
func (dc *DefaultConfig) MergeTags(resourceTags []string) []string {
	if dc == nil {
		return resourceTags
	}

	return MergeTags(dc.Tags, resourceTags)
}

// RemoveDefaultTags strips every tag that is present in defaultTags from
// allTags. This is used during Read to split the full tag set returned by
// the API into the resource-specific portion (stored in the tags attribute)
// and the default portion (which is recomputed from the provider config).
//
// If defaultTags is empty, allTags is returned as-is.
func RemoveDefaultTags(allTags, defaultTags []string) []string {
	if len(defaultTags) == 0 {
		return allTags
	}

	defaults := make(map[string]struct{}, len(defaultTags))
	for _, t := range defaultTags {
		defaults[t] = struct{}{}
	}

	result := make([]string, 0, len(allTags))
	for _, t := range allTags {
		if _, ok := defaults[t]; !ok {
			result = append(result, t)
		}
	}

	return result
}

// RemoveDefaultTags returns the result of RemoveDefaultTags using the
// DefaultConfig's tags.
func (dc *DefaultConfig) RemoveDefaultTags(allTags []string) []string {
	if dc == nil {
		return allTags
	}

	return RemoveDefaultTags(allTags, dc.Tags)
}

// TagsEqual returns true if two string slices contain the same elements
// (order-insensitive). nil and empty slices are considered equal.
func TagsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	set := make(map[string]int, len(a))
	for _, t := range a {
		set[t]++
	}

	for _, t := range b {
		if count, ok := set[t]; !ok || count == 0 {
			return false
		}
		set[t]--
	}

	return true
}

package scraper

import (
	"fmt"
	"regexp"
	"strings"
)

// PathFilter provides glob-based filtering for URL paths.
// It supports include and exclude patterns with the following rules:
// - Empty include patterns means "include all"
// - Exclude patterns take priority over include patterns
// - Patterns support glob-style wildcards (*)
type PathFilter struct {
	includePaths    []string
	excludePaths    []string
	compiledInclude []*regexp.Regexp
	compiledExclude []*regexp.Regexp
}

// NewPathFilter creates a new PathFilter with the given include and exclude patterns.
// Patterns support glob-style wildcards:
// - "*" matches any sequence of characters
// - "/blog/*" matches "/blog/foo", "/blog/bar/baz", etc.
// - "/docs/v1" matches exactly "/docs/v1"
// Patterns are case-insensitive.
func NewPathFilter(includePaths, excludePaths []string) (*PathFilter, error) {
	filter := &PathFilter{
		includePaths: includePaths,
		excludePaths: excludePaths,
	}

	// Compile include patterns
	for _, pattern := range includePaths {
		regex, err := globToRegex(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid include pattern %q: %w", pattern, err)
		}
		filter.compiledInclude = append(filter.compiledInclude, regex)
	}

	// Compile exclude patterns
	for _, pattern := range excludePaths {
		regex, err := globToRegex(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
		}
		filter.compiledExclude = append(filter.compiledExclude, regex)
	}

	return filter, nil
}

// Match returns true if the given URL path should be included based on the filter rules.
// Rules:
// 1. If path matches any exclude pattern, return false
// 2. If include patterns are empty, return true (include all)
// 3. If path matches any include pattern, return true
// 4. Otherwise, return false
func (f *PathFilter) Match(urlPath string) bool {
	// Normalize path (remove trailing slash for consistency, except for root)
	normalizedPath := strings.TrimSuffix(urlPath, "/")
	if normalizedPath == "" {
		normalizedPath = "/"
	}

	// Check exclude patterns first (they take priority)
	for _, regex := range f.compiledExclude {
		if regex.MatchString(normalizedPath) {
			return false
		}
	}

	// If no include patterns, include everything (that wasn't excluded)
	if len(f.compiledInclude) == 0 {
		return true
	}

	// Check include patterns
	for _, regex := range f.compiledInclude {
		if regex.MatchString(normalizedPath) {
			return true
		}
	}

	// No include pattern matched
	return false
}

// FilterURLs filters a list of URLs based on the path filter rules.
// Returns a new slice containing only the URLs that match the filter.
func (f *PathFilter) FilterURLs(urls []string) []string {
	if f == nil {
		return urls
	}

	filtered := make([]string, 0, len(urls))
	for _, url := range urls {
		// Extract path from URL
		// Simple extraction: find everything after the domain
		// Format: scheme://domain/path
		parts := strings.SplitN(url, "://", 2)
		if len(parts) != 2 {
			continue // Invalid URL format
		}

		afterScheme := parts[1]
		slashIdx := strings.Index(afterScheme, "/")

		var path string
		if slashIdx == -1 {
			path = "/" // Root path
		} else {
			path = afterScheme[slashIdx:]
		}

		if f.Match(path) {
			filtered = append(filtered, url)
		}
	}

	return filtered
}

// globToRegex converts a glob pattern to a regular expression.
// Supports:
// - "*" for any sequence of characters
// - Case-insensitive matching
func globToRegex(pattern string) (*regexp.Regexp, error) {
	// Escape special regex characters except *
	escaped := regexp.QuoteMeta(pattern)

	// Replace escaped \* with .*
	regexPattern := strings.ReplaceAll(escaped, `\*`, `.*`)

	// Anchor the pattern to match the entire path
	regexPattern = "^" + regexPattern + "$"

	// Compile as case-insensitive
	return regexp.Compile("(?i)" + regexPattern)
}

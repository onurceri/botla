// Package urlutil provides URL normalization and utility functions.
package urlutil

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeURL normalizes URLs for consistent storage and comparison.
// Rules:
//  1. Trim whitespace
//  2. Lowercase the scheme and host
//  3. Remove trailing slashes from paths (except for root path)
//  4. Preserve query parameters and fragments as-is
//
// Returns an error if the URL is malformed.
func NormalizeURL(rawURL string) (string, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "", nil
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("parse url failed: %w", err)
	}

	// Lowercase scheme and host
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Remove trailing slash from path, but keep root "/" as empty
	// e.g., "https://example.com/" -> "https://example.com"
	// e.g., "https://example.com/page/" -> "https://example.com/page"
	if u.Path != "" && u.Path != "/" {
		u.Path = strings.TrimSuffix(u.Path, "/")
	} else {
		u.Path = ""
	}

	return u.String(), nil
}

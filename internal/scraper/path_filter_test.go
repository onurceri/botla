package scraper

import (
	"testing"
)

func TestPathFilter_Match(t *testing.T) {
	testCases := []struct {
		name     string
		include  []string
		exclude  []string
		path     string
		expected bool
	}{
		// Include tests
		{
			name:     "include all when empty",
			include:  nil,
			exclude:  nil,
			path:     "/any/path",
			expected: true,
		},
		{
			name:     "include glob match",
			include:  []string{"/blog/*"},
			exclude:  nil,
			path:     "/blog/post-1",
			expected: true,
		},
		{
			name:     "include glob deep match",
			include:  []string{"/blog/*"},
			exclude:  nil,
			path:     "/blog/category/post-1",
			expected: true,
		},
		{
			name:     "include no match",
			include:  []string{"/blog/*"},
			exclude:  nil,
			path:     "/docs/intro",
			expected: false,
		},
		{
			name:     "include exact match",
			include:  []string{"/about"},
			exclude:  nil,
			path:     "/about",
			expected: true,
		},
		{
			name:     "include multiple patterns - first matches",
			include:  []string{"/blog/*", "/docs/*"},
			exclude:  nil,
			path:     "/blog/post",
			expected: true,
		},
		{
			name:     "include multiple patterns - second matches",
			include:  []string{"/blog/*", "/docs/*"},
			exclude:  nil,
			path:     "/docs/intro",
			expected: true,
		},
		{
			name:     "include multiple patterns - no match",
			include:  []string{"/blog/*", "/docs/*"},
			exclude:  nil,
			path:     "/about",
			expected: false,
		},

		// Exclude tests
		{
			name:     "exclude takes priority over include all",
			include:  nil,
			exclude:  []string{"/admin/*"},
			path:     "/admin/users",
			expected: false,
		},
		{
			name:     "exclude takes priority over explicit include",
			include:  []string{"/*"},
			exclude:  []string{"/admin/*"},
			path:     "/admin/users",
			expected: false,
		},
		{
			name:     "exclude partial match",
			include:  nil,
			exclude:  []string{"/tag/*"},
			path:     "/tag/golang",
			expected: false,
		},
		{
			name:     "exclude no match",
			include:  nil,
			exclude:  []string{"/tag/*"},
			path:     "/blog/post",
			expected: true,
		},
		{
			name:     "exclude multiple patterns",
			include:  nil,
			exclude:  []string{"/admin/*", "/tag/*"},
			path:     "/tag/python",
			expected: false,
		},

		// Combined include and exclude
		{
			name:     "include and exclude - both match, exclude wins",
			include:  []string{"/blog/*"},
			exclude:  []string{"/blog/private/*"},
			path:     "/blog/private/secret",
			expected: false,
		},
		{
			name:     "include and exclude - only include matches",
			include:  []string{"/blog/*"},
			exclude:  []string{"/blog/private/*"},
			path:     "/blog/public/post",
			expected: true,
		},
		{
			name:     "include and exclude - neither matches",
			include:  []string{"/blog/*"},
			exclude:  []string{"/admin/*"},
			path:     "/docs/intro",
			expected: false,
		},

		// Edge cases
		{
			name:     "trailing slash - with wildcard pattern",
			include:  []string{"/blog/*"},
			exclude:  nil,
			path:     "/blog/post",
			expected: true,
		},
		{
			name:     "trailing slash - exact match",
			include:  []string{"/blog"},
			exclude:  nil,
			path:     "/blog/",
			expected: true,
		},
		{
			name:     "root path",
			include:  []string{"/*"},
			exclude:  nil,
			path:     "/",
			expected: true,
		},
		{
			name:     "case insensitive - uppercase path",
			include:  []string{"/blog/*"},
			exclude:  nil,
			path:     "/BLOG/post",
			expected: true,
		},
		{
			name:     "case insensitive - uppercase pattern",
			include:  []string{"/BLOG/*"},
			exclude:  nil,
			path:     "/blog/post",
			expected: true,
		},
		{
			name:     "wildcard in middle",
			include:  []string{"/api/*/users"},
			exclude:  nil,
			path:     "/api/v1/users",
			expected: true,
		},
		{
			name:     "multiple wildcards",
			include:  []string{"/api/*/v*/users"},
			exclude:  nil,
			path:     "/api/public/v2/users",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filter, err := NewPathFilter(tc.include, tc.exclude)
			if err != nil {
				t.Fatalf("Failed to create filter: %v", err)
			}

			result := filter.Match(tc.path)
			if result != tc.expected {
				t.Errorf("Match(%q) = %v, expected %v\nInclude: %v\nExclude: %v",
					tc.path, result, tc.expected, tc.include, tc.exclude)
			}
		})
	}
}

func TestPathFilter_FilterURLs(t *testing.T) {
	testCases := []struct {
		name     string
		include  []string
		exclude  []string
		urls     []string
		expected []string
	}{
		{
			name:    "filter with include pattern",
			include: []string{"/blog/*"},
			exclude: nil,
			urls: []string{
				"https://example.com/blog/post-1",
				"https://example.com/blog/post-2",
				"https://example.com/docs/intro",
				"https://example.com/about",
			},
			expected: []string{
				"https://example.com/blog/post-1",
				"https://example.com/blog/post-2",
			},
		},
		{
			name:    "filter with exclude pattern",
			include: nil,
			exclude: []string{"/admin/*"},
			urls: []string{
				"https://example.com/blog/post-1",
				"https://example.com/admin/users",
				"https://example.com/admin/settings",
				"https://example.com/about",
			},
			expected: []string{
				"https://example.com/blog/post-1",
				"https://example.com/about",
			},
		},
		{
			name:    "filter with both include and exclude",
			include: []string{"/blog/*"},
			exclude: []string{"/blog/private/*"},
			urls: []string{
				"https://example.com/blog/post-1",
				"https://example.com/blog/private/secret",
				"https://example.com/blog/public/article",
				"https://example.com/docs/intro",
			},
			expected: []string{
				"https://example.com/blog/post-1",
				"https://example.com/blog/public/article",
			},
		},
		{
			name:    "no filter - include all",
			include: nil,
			exclude: nil,
			urls: []string{
				"https://example.com/blog/post-1",
				"https://example.com/docs/intro",
				"https://example.com/about",
			},
			expected: []string{
				"https://example.com/blog/post-1",
				"https://example.com/docs/intro",
				"https://example.com/about",
			},
		},
		{
			name:    "root path handling",
			include: []string{"/*"},
			exclude: []string{"/admin"},
			urls: []string{
				"https://example.com/",
				"https://example.com/admin",
				"https://example.com/blog",
			},
			expected: []string{
				"https://example.com/",
				"https://example.com/blog",
			},
		},
		{
			name:    "different domains",
			include: []string{"/blog/*"},
			exclude: nil,
			urls: []string{
				"https://example.com/blog/post-1",
				"https://other.com/blog/post-2",
				"https://example.com/docs/intro",
			},
			expected: []string{
				"https://example.com/blog/post-1",
				"https://other.com/blog/post-2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filter, err := NewPathFilter(tc.include, tc.exclude)
			if err != nil {
				t.Fatalf("Failed to create filter: %v", err)
			}

			result := filter.FilterURLs(tc.urls)

			if len(result) != len(tc.expected) {
				t.Errorf("FilterURLs returned %d URLs, expected %d\nGot: %v\nExpected: %v",
					len(result), len(tc.expected), result, tc.expected)
				return
			}

			for i, url := range result {
				if url != tc.expected[i] {
					t.Errorf("FilterURLs[%d] = %q, expected %q", i, url, tc.expected[i])
				}
			}
		})
	}
}

func TestPathFilter_NilFilter(t *testing.T) {
	var filter *PathFilter
	urls := []string{
		"https://example.com/blog/post-1",
		"https://example.com/docs/intro",
	}

	result := filter.FilterURLs(urls)

	if len(result) != len(urls) {
		t.Errorf("Nil filter should return all URLs, got %d, expected %d", len(result), len(urls))
	}
}

func TestNewPathFilter_InvalidPattern(t *testing.T) {
	// Note: Our current implementation uses regexp.QuoteMeta which escapes most special chars,
	// so it's hard to create an invalid pattern. This test is here for completeness.
	// If we add more complex pattern validation in the future, this test will be useful.

	// For now, test that valid patterns work
	_, err := NewPathFilter([]string{"/blog/*"}, []string{"/admin/*"})
	if err != nil {
		t.Errorf("Valid patterns should not error: %v", err)
	}
}

func TestGlobToRegex(t *testing.T) {
	testCases := []struct {
		pattern  string
		testPath string
		expected bool
	}{
		{"/blog/*", "/blog/post", true},
		{"/blog/*", "/BLOG/post", true}, // case insensitive
		{"/blog/*", "/docs/post", false},
		{"/api/*/users", "/api/v1/users", true},
		{"/api/*/users", "/api/public/users", true},
		{"/api/*/users", "/api/v1/posts", false},
		{"/*", "/anything", true},
		{"/*", "/", true},
		{"/exact", "/exact", true},
		{"/exact", "/exact/more", false},
	}

	for _, tc := range testCases {
		t.Run(tc.pattern+"_"+tc.testPath, func(t *testing.T) {
			regex, err := globToRegex(tc.pattern)
			if err != nil {
				t.Fatalf("Failed to compile pattern %q: %v", tc.pattern, err)
			}

			result := regex.MatchString(tc.testPath)
			if result != tc.expected {
				t.Errorf("Pattern %q matching %q = %v, expected %v",
					tc.pattern, tc.testPath, result, tc.expected)
			}
		})
	}
}

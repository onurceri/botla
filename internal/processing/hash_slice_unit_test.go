package processing

import "testing"

func TestSafeHashPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "hash longer than max length is truncated",
			input:    "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			maxLen:   16,
			expected: "abcdef1234567890",
		},
		{
			name:     "hash exactly max length returns full hash",
			input:    "abcdef1234567890",
			maxLen:   16,
			expected: "abcdef1234567890",
		},
		{
			name:     "hash shorter than max length returns full hash (edge case: 3 chars)",
			input:    "123",
			maxLen:   16,
			expected: "123",
		},
		{
			name:     "empty hash returns empty string",
			input:    "",
			maxLen:   16,
			expected: "",
		},
		{
			name:     "single character hash",
			input:    "a",
			maxLen:   16,
			expected: "a",
		},
		{
			name:     "hash with max length 0 returns empty",
			input:    "abcdef1234567890",
			maxLen:   0,
			expected: "",
		},
		{
			name:     "hash with max length 1",
			input:    "abcdef1234567890",
			maxLen:   1,
			expected: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := safeHashPrefix(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("safeHashPrefix(%q, %d) = %q, expected %q",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// Test that safeHashPrefix does not panic with various edge cases
func TestSafeHashPrefix_NoPanic(t *testing.T) {
	t.Parallel()

	// These should not panic
	panicTests := []struct {
		name   string
		input  string
		maxLen int
	}{
		{"empty string", "", 16},
		{"short string", "abc", 16},
		{"negative max length should not panic", "abc", -1},
	}

	for _, tt := range panicTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("safeHashPrefix(%q, %d) panicked: %v",
						tt.input, tt.maxLen, r)
				}
			}()
			_ = safeHashPrefix(tt.input, tt.maxLen)
		})
	}
}

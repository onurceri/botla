package testdb

import "testing"

func TestExtractDirtyVersion(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected int
	}{
		{
			name:     "dirty version 16",
			output:   "error: Dirty database version 16. Fix and force version.",
			expected: 16,
		},
		{
			name:     "dirty version 1",
			output:   "Dirty database version 1.",
			expected: 1,
		},
		{
			name:     "dirty version 100",
			output:   "error: Dirty database version 100. Fix and force version.",
			expected: 100,
		},
		{
			name:     "no dirty version",
			output:   "migration failed: some other error",
			expected: 0,
		},
		{
			name:     "empty string",
			output:   "",
			expected: 0,
		},
		{
			name:     "partial match no number",
			output:   "Dirty database version ",
			expected: 0,
		},
		{
			name:     "multiline output with dirty version",
			output:   "Running migrations...\nerror: Dirty database version 42. Fix and force version.\nMigration aborted.",
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDirtyVersion(tt.output)
			if result != tt.expected {
				t.Errorf("extractDirtyVersion(%q) = %d, want %d", tt.output, result, tt.expected)
			}
		})
	}
}

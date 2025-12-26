package errors

import (
	"errors"
	"testing"
)

func TestWrapf(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		format      string
		args        []interface{}
		wantNil     bool
		wantContain string
	}{
		{
			name:        "nil error returns nil",
			err:         nil,
			format:      "context message",
			args:        nil,
			wantNil:     true,
			wantContain: "",
		},
		{
			name:        "wraps error with simple message",
			err:         errors.New("original error"),
			format:      "simple context",
			args:        nil,
			wantNil:     false,
			wantContain: "simple context: original error",
		},
		{
			name:        "wraps error with formatted message",
			err:         errors.New("connection failed"),
			format:      "fetching from %s:%d",
			args:        []interface{}{"example.com", 80},
			wantNil:     false,
			wantContain: "fetching from example.com:80: connection failed",
		},
		{
			name:        "preserves original error for unwrapping",
			err:         errors.New("original"),
			format:      "wrapped",
			args:        nil,
			wantNil:     false,
			wantContain: "wrapped: original",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrapf(tt.err, tt.format, tt.args...)

			if tt.wantNil {
				if got != nil {
					t.Errorf("Wrapf() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("Wrapf() = nil, want non-nil error")
			}

			gotStr := got.Error()
			if gotStr != tt.wantContain {
				t.Errorf("Wrapf().Error() = %q, want %q", gotStr, tt.wantContain)
			}

			// Verify error can be unwrapped to original
			if tt.err != nil && !errors.Is(got, tt.err) {
				t.Errorf("Wrapf() error chain does not contain original error")
			}
		})
	}
}

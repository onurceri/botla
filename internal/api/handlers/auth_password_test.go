package handlers

import "testing"

// isStrongPassword: >= 8 chars with an upper, lower, digit and one of @$!%*?&.
func TestIsStrongPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{name: "valid", password: "Test@123", want: true},
		{name: "valid long", password: "C0rrect-Horse$Battery", want: true},
		{name: "every special char accepted", password: "Aa1@$!%*?&", want: true},
		{name: "too short", password: "Te@1", want: false},
		{name: "seven chars", password: "Tes@123", want: false},
		{name: "empty", password: "", want: false},
		{name: "no upper", password: "test@123", want: false},
		{name: "no lower", password: "TEST@123", want: false},
		{name: "no digit", password: "Test@abc", want: false},
		{name: "no special", password: "Test1234", want: false},
		{name: "special outside allowed set", password: "Test#1234", want: false},
		{name: "whitespace only", password: "        ", want: false},
		{name: "leading space still needs all classes", password: " test1234", want: false},
		{name: "leading space with all classes", password: " Test@123", want: true},
		{name: "non-ascii letters do not count as upper/lower", password: "ÇĞÜ@1234", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStrongPassword(tt.password); got != tt.want {
				t.Fatalf("isStrongPassword(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

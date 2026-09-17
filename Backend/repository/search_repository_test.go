package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEscapeLikePattern memverifikasi bahwa input pencarian diperlakukan
// sebagai teks biasa: wildcard LIKE dari user (% _) di-escape sehingga
// query seperti "q=%25" tidak lagi match semua baris (bug audit Week 5).
func TestEscapeLikePattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "teks biasa", input: "ali", expected: "ali"},
		{name: "percent", input: "100%", expected: `100\%`},
		{name: "underscore", input: "user_name", expected: `user\_name`},
		{name: "backslash", input: `a\b`, expected: `a\\b`},
		{name: "kombinasi wildcard", input: "%_", expected: `\%\_`},
		{name: "unicode utuh", input: "nama café ☕", expected: "nama café ☕"},
		{name: "string kosong", input: "", expected: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, escapeLikePattern(tc.input))
		})
	}
}

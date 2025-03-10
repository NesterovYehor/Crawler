package tests

import (
	"testing"

	"github.com/NesterovYehor/Crawler/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		expected string
	}{
		{
			name:     "remove scheme",
			inputURL: "https://blog.boot.dev/path",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "remove trailing slash",
			inputURL: "https://blog.boot.dev/path/",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "keep root path",
			inputURL: "https://blog.boot.dev/",
			expected: "blog.boot.dev",
		},
		{
			name:     "no path",
			inputURL: "https://blog.boot.dev",
			expected: "blog.boot.dev",
		},
		{
			name:     "handle subdomains",
			inputURL: "https://sub.blog.boot.dev/path",
			expected: "sub.blog.boot.dev/path",
		},
	}
	for _, tc := range tests {
		actual, err := utils.NormalizeURL(tc.inputURL)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, actual)
	}
}

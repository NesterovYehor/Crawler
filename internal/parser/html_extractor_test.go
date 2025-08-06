package parser_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/NesterovYehor/Crawler/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestGetURLsFromHTML(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		inputBody string
		expected  []string
	}{
		{
			name:    "absolute and relative URLs",
			baseURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a href="/path/one">
			<span>Boot.dev</span>
		</a>
		<a href="https://other.com/path/one">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://blog.boot.dev/path/one", "https://other.com/path/one"},
		},
		{
			name:    "trailing slash in base URL",
			baseURL: "https://blog.boot.dev/",
			inputBody: `
<html>
	<body>
		<a href="path/two">Link</a>
	</body>
</html>
`,
			expected: []string{"https://blog.boot.dev/path/two"},
		},
		{
			name:    "URL with query parameters",
			baseURL: "https://example.com",
			inputBody: `
<html>
	<body>
		<a href="/search?q=golang">Search</a>
	</body>
</html>
`,
			expected: []string{"https://example.com/search?q=golang"},
		},
		{
			name:    "empty input",
			baseURL: "https://example.com",
			inputBody: `
<html>
	<body>
	</body>
</html>
`,
			expected: []string{}, // No links at all
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader
            base, err := url.Parse(tc.baseURL)
            assert.NoError(t, err)
			actual, err := parser.GetURLsFromHTML(reader(tc.inputBody), base)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.expected, actual)
		})
	}
}

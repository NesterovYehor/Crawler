package utils

import (
	"net/url"
	"strings"
)

func NormalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Build the normalized URL
	normalized := parsedURL.Host + parsedURL.Path

	// Remove trailing slash if it's not the root "/"
	normalized = strings.TrimSuffix(normalized, "/")

	return normalized, nil
}

package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func NormalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	normalized := parsedURL.Host + parsedURL.Path

	normalized = strings.TrimSuffix(normalized, "/")

	return normalized, nil
}

func GetDomain(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
        return "", fmt.Errorf("FAILED TO PARSE URL:%v ERR:%v", rawURL, err)
	}

	domain := parsedURL.Host

	domain = strings.TrimSuffix(domain, "/")

	return domain, nil
}

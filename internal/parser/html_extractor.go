package parser

import (
	"fmt"
	"io"
	"net/url"

	"golang.org/x/net/html"
)

func GetURLsFromHTML(htmlBody io.Reader, base *url.URL) ([]string, error) {
	node, err := html.Parse(htmlBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	urls := make([]string, 0)

	var extractLinks func(*html.Node)
	extractLinks = func(n *html.Node) {
		if n == nil {
			return
		}

		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					parsedURL, err := url.Parse(attr.Val)
					if err != nil {
						fmt.Printf("Skipping invalid URL: %v\n", err)
						continue
					}

					if !parsedURL.IsAbs() {
						parsedURL = base.ResolveReference(parsedURL)
					}

					if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
						continue
					}

					if parsedURL.Host == "" {
						continue
					}

					urls = append(urls, parsedURL.String())
					break
				}
			}
		}

		extractLinks(n.FirstChild)
		extractLinks(n.NextSibling)
	}

	extractLinks(node)

	return urls, nil
}

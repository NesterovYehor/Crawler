package parser

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func GetURLsFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	reader := strings.NewReader(htmlBody)

	node, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}

	var urls []string

	var dfs func(node *html.Node)
	dfs = func(node *html.Node) {
		if node == nil {
			return
		}

		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					parsedURL, err := url.Parse(attr.Val)
					if err == nil {
						resolvedURL := baseURL.ResolveReference(parsedURL)
						urls = append(urls, resolvedURL.String())
					}
					break
				}
			}
		}

		dfs(node.FirstChild)
		dfs(node.NextSibling)
	}

	dfs(node)
	return urls, nil
}

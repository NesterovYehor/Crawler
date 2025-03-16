package parser

import (
	"fmt"
	"io"
	"net/url"

	"golang.org/x/net/html"
)

func GetURLsFromHTML(htmlBody io.Reader, baseURL *url.URL) ([]string, error) {
	node, err := html.Parse(htmlBody)
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
					if err != nil {
                        fmt.Println(err)
                        continue
					}
                    resolvedURL := baseURL.ResolveReference(parsedURL)
                    if resolvedURL.Host == baseURL.Host {
                        urls = append(urls, resolvedURL.Path)
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

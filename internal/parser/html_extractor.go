package parser

import (
	"fmt"
	"io"
	"net/url"

	"github.com/NesterovYehor/Crawler/internal/utils"
	"golang.org/x/net/html"
)

func GetURLsFromHTML(htmlBody io.Reader, rawUrl string) ([]string, error) {
	baseRaw, err := utils.NormalizeURL(rawUrl)
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(baseRaw)
	if err != nil {
		return nil, err
	}
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
					resolvedURL := base.ResolveReference(parsedURL)
					urls = append(urls, resolvedURL.Path)
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

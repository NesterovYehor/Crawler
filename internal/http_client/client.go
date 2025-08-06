package httpclient

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/NesterovYehor/Crawler/internal/utils"
	"github.com/temoto/robotstxt"
)

type Urlset struct {
	URLs []URL `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
}

type Interface interface {
	FetchBody(rawURL string) (io.ReadCloser, error)
	FetchRules(baseURL string) ([]byte, bool, []string, error)
}

type HTTP struct {
	client    *http.Client
	Benchmark []time.Duration
	Count     int
}

func NewHTTPClient(idleConns int) Interface {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          idleConns,
		MaxIdleConnsPerHost:   idleConns,
		IdleConnTimeout:       120 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second, // Keep this, as your overall client timeout is 15s
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		DisableKeepAlives:     false,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
	}

	return &HTTP{
		client: &http.Client{
			Timeout:   15 * time.Second, // Keep this at 15 seconds
			Transport: transport,
		},
	}
}

func (c *HTTP) FetchBody(rawURL string) (io.ReadCloser, error) {
	start := time.Now()
	req, err := http.NewRequest("GET", rawURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", rawURL, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", rawURL, err)
	}

	c.Benchmark = append(c.Benchmark, time.Since(start))

	if resp.StatusCode == 408 || resp.StatusCode == 429 {
		return nil, utils.ErrRetryLater
	}
	if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
		return nil, utils.ErrInvalidStatusCode(resp.StatusCode)
	}

	if time.Since(start) >= 500*time.Millisecond {
		fmt.Println("URL TOOK:", time.Since(start))
		c.Count++
	}

	return resp.Body, nil
}

func (c *HTTP) FetchRules(baseURL string) ([]byte, bool, []string, error) {
	rulesUrl, err := url.Parse(baseURL + "/robots.txt")
	if err != nil {
		return nil, false, nil, err
	}
	rulesUrl.Scheme = "https"

	resp, err := c.FetchBody(rulesUrl.String())
	if err != nil {
		fmt.Println(rulesUrl.String())
		if errors.Is(err, utils.ErrRetryLater) {
			return nil, true, nil, nil
		}
		return nil, false, nil, err
	}
	defer func() {
		if err := resp.Close(); err != nil {
			slog.Error(err.Error())
		}
	}()
	body, err := io.ReadAll(resp)
	if err != nil {
		return nil, false, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	rules, err := robotstxt.FromBytes(body)
	if err != nil {
		return nil, false, nil, err
	}
	if len(rules.Sitemaps) != 0 {
		urls, err := c.fetchSitemapURLs(rules.Sitemaps)
		if err != nil {
			return nil, false, nil, err
		}
		return body, false, urls, nil
	}

	return body, false, nil, nil
}

func (c *HTTP) fetchSitemapURLs(urls []string) ([]string, error) {
	if len(urls) == 0 {
		return nil, utils.ErrNoURLsProvided
	}

	var allURLs []string

	for _, sitemapURL := range urls {
		resp, err := c.FetchBody(sitemapURL)
		if err != nil {
			return nil, err
		}
		defer func() {
			if closer, ok := resp.(io.Closer); ok {
				if err := closer.Close(); err != nil {
					slog.Error(err.Error())
				}
			}
		}()

		body, err := io.ReadAll(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var urlset Urlset
		if err := xml.Unmarshal(body, &urlset); err != nil {
			fmt.Printf("error decoding XML from %s: %v\n", sitemapURL, err)
			return nil, nil
		}

		for _, u := range urlset.URLs {
			allURLs = append(allURLs, u.Loc)
		}
	}

	return allURLs, nil
}

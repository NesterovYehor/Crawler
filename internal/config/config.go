package config

import (
	"net/url"
	"sync"
)

type Config struct {
	Pages              map[string]int
	BaseURL            *url.URL
	Mu                 *sync.Mutex
	ConcurrencyControl chan struct{}
	MaxPages           int
	Wg                 *sync.WaitGroup
}

func NewConfig(rawBaseURL string, concurrencyLimit, maxPages int) (*Config, error) {
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	return &Config{
		Pages:              make(map[string]int),
		BaseURL:            baseURL,
		Mu:                 &sync.Mutex{},
		ConcurrencyControl: make(chan struct{}, concurrencyLimit),
		MaxPages:           maxPages,
		Wg:                 &sync.WaitGroup{},
	}, nil
}

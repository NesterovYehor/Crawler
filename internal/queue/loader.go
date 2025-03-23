package queue

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

func FillQueue(filePath string, q *Queue) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}

	urls, err := loadURLs(filePath)
	if err != nil {
		return fmt.Errorf("failed to load URLs from file: %w", err)
	}

	for _, url := range urls {
		values := map[string]any{
			"topic":   "fetch_rules",
			"retries": 0,
			"url":     url,
		}

		if err := q.AddToQueue(values); err != nil {
			return fmt.Errorf("failed to add task to queue for URL %s: %w", url, err)
		}
	}

	return nil
}

func loadURLs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	urls := []string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	return urls, nil
}

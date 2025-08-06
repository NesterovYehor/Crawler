package loader

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/NesterovYehor/Crawler/internal/models"
	"github.com/NesterovYehor/Crawler/internal/queue"
)

const initialTopic = "fetch_rules"

func FillQueue(filePath string, q queue.Interface) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}

	urls, err := loadURLs(filePath)
	if err != nil {
		return fmt.Errorf("failed to load URLs from file: %w", err)
	}

	msgs := make([]*models.Task, len(urls))
	for i, url := range urls {
		m := models.NewTask(initialTopic, url, queue.HighPriorityQueue, "")
		msgs[i] = m
	}

	if err := q.Add(msgs); err != nil {
		return fmt.Errorf("failed to add tasks to queue %w", err)
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

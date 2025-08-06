package mocks

import (
	"context"

	"github.com/NesterovYehor/Crawler/internal/models"
	"github.com/NesterovYehor/Crawler/internal/queue"
)

type Source []*models.Task

type MockQueue struct {
	Sources         map[string]Source
	DeletedMessages []*models.Task
	Err             error
}

func NewMockQueue(err error) *MockQueue {
	return &MockQueue{
		Sources: make(map[string]Source),
		Err:     err,
	}
}

func (m *MockQueue) Add(tasks []*models.Task) error {
	for _, task := range tasks {
		m.Sources[task.SourceName] = tasks
	}
	return nil
}

func (m *MockQueue) GetMessages(ctx context.Context, count int, source queue.Source) ([]*models.Task, error) {
	return m.Sources[source.Curr][0:count], m.Err
}

func (m *MockQueue) AcknowledgeMessage(tasks []*models.Task) error {
	for _, task := range tasks {
		for i := range len(m.Sources[task.SourceName]) - 1 {
			if m.Sources[task.SourceName][i].ID == task.ID {
				m.DeletedMessages = append(m.DeletedMessages, task)
				temp := m.Sources[task.SourceName][:i]
				temp = append(temp, m.Sources[task.SourceName][i+1:]...)
				m.Sources[task.SourceName] = temp
				continue
			}
		}
	}
	return m.Err
}

func (m *MockQueue) Retry(task *models.Task) error {
	m.Sources[task.SourceName] = append(m.Sources[queue.RetryPriorityQueue], task)

	return m.Err
}

func (m *MockQueue) Close() error {
	m.Sources = nil
	return m.Err
}

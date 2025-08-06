package mocks

import (
	"context"
	"errors"

	"github.com/NesterovYehor/Crawler/internal/models"
)

type StorageMock struct {
	bloomBuffer map[string]bool
	cacheBuff   map[string]map[string]any
}

func (m *StorageMock) ExistsInBF(key string) (bool, error) {
	if key == "" {
		return false, errors.New("BloomFilter ERROR: Empty key")
	}
	return m.bloomBuffer[key], nil
}

func (m *StorageMock) AddToBF(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("BloomFilter ERROR: Empty key")
	}
	m.bloomBuffer[key] = true
	return nil
}

func (m *StorageMock) SaveToCache(ctx context.Context, key string, values map[string]any) error {
	if key == "" {
		return errors.New("cache ERROR: Tried to store key is empty")
	}
	if key == "" {
		return errors.New("cache ERROR: Tried to store values equeal nil")
	}

	m.cacheBuff[key] = values

	return nil
}

func (m *StorageMock) RunScript(key, sriptHash string) (any, error) {
    return nil, nil
}

func (m *StorageMock) SaveIfNew(ctx context.Context, data *models.PageDataModel) error {
    return nil
}

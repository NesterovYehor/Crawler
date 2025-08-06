package mocks

import (
	"errors"
	"io"
)

type MockHTTPClient struct {
	FetchBodyFn  func(rawURL string) (io.ReadCloser, error)
	FetchRulesFn func(baseURL string) ([]byte, bool, []string, error)
}

// FetchBody mocks the FetchBody method.
func (m *MockHTTPClient) FetchBody(rawURL string) (io.ReadCloser, error) {
	if m.FetchBodyFn != nil {
		body, err := m.FetchBodyFn(rawURL)
		return body, err
	}
	return nil, errors.New("FetchBodyFn not set")
}

// FetchRules mocks the FetchRules method.
func (m *MockHTTPClient) FetchRules(baseURL string) ([]byte, bool, []string, error) {
	if m.FetchRulesFn != nil {
		return m.FetchRulesFn(baseURL)
	}
	return nil, false, nil, errors.New("FetchRulesFn not set")
}

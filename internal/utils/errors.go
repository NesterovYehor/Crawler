package utils

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidContentType = errors.New("invalid response content-type header")
	ErrNoURLsProvided     = errors.New("no URLs provided")
	ErrInValidUrl         = errors.New("provided url is not valid")
	ErrRetryLater         = errors.New("retry Later")
	ErrEmptyDomain        = errors.New("empty domain")
	ErrNoTasks            = errors.New("no tasks available")
	ErrInValidPageData    = errors.New("page data is invalid")
	ErrChanIsClosed       = errors.New("channel is closed, cannot get task")
	ErrDomainRateLimited  = errors.New("domain is currently rate limited")
)

func ErrInvalidTaskFormat(msg any) error {
	return fmt.Errorf("invalid message format: %v", msg)
}

func ErrInvalidRetriesValue(err error) error {
	return fmt.Errorf("invalid retries value: %w", err)
}

func ErrInvalidStatusCode(statusCode int) error{
	return fmt.Errorf("invalid response status code: %v", statusCode)
}

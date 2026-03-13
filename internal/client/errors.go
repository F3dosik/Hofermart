package client

import (
	"errors"
	"fmt"
	"time"
)

type ErrRateLimit struct {
	RetryAfter time.Duration
}

func (e *ErrRateLimit) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %s", e.RetryAfter)
}

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrBuildURL      = errors.New("build url error")
	ErrRequestExec   = errors.New("request execution error")
)

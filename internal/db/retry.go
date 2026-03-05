package db

import (
	"context"
	"fmt"
	"time"
)

func WithRetry(ctx context.Context, op func() error) error {
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	maxAttempts := len(delays) + 1

	var err error
	for i := 0; i < maxAttempts; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err = op()
		if err == nil {
			return nil
		}

		if !IsRetriable(err) || i == len(delays) {
			return fmt.Errorf("operation failed after %d attempt(s): %w", i+1, err)
		}

		select {
		case <-time.After(delays[i]):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("operation failed after %d attempt(s): %w", maxAttempts, err)
}

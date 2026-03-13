package worker

import (
	"context"
	"time"
)

func (w *Worker) waitIfPaused(ctx context.Context) {
	for {
		until := time.Unix(0, w.pauseUntil.Load())
		if time.Now().After(until) {
			return
		}

		select {
		case <-time.After(time.Until(until)):
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) setPause(until time.Time) {
	newUntil := until.UnixNano()

	for {
		current := w.pauseUntil.Load()
		if newUntil <= current {
			return
		}
		if w.pauseUntil.CompareAndSwap(current, newUntil) {
			return
		}
	}
}

func (w *Worker) backoffDelay(attempt int) time.Duration {
	delay := w.pollInterval * time.Duration(attempt)
	return min(delay, w.maxDelay)
}

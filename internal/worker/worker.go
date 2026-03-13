package worker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/F3dosik/Hofermart/internal/client"
	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/F3dosik/Hofermart/internal/repository"
	"go.uber.org/zap"
)

type Worker struct {
	repo         repository.Repository
	client       client.AccrualClient
	scheduler    Scheduler
	jobChan      chan *ScheduleJob
	count        int
	pollInterval time.Duration
	pauseUntil   atomic.Int64
	wg           sync.WaitGroup
	maxDelay     time.Duration
	logger       *zap.SugaredLogger
}

func New(
	repo repository.Repository,
	client client.AccrualClient,
	scheduler Scheduler,
	jobChan chan *ScheduleJob,
	count int,
	pollInterval time.Duration,
	maxDelay time.Duration,
	logger *zap.SugaredLogger,
) *Worker {
	return &Worker{
		repo:         repo,
		client:       client,
		scheduler:    scheduler,
		jobChan:      jobChan,
		count:        count,
		pollInterval: pollInterval,
		maxDelay:     maxDelay,
		logger:       logger,
	}
}

func (w *Worker) Run(ctx context.Context) {
	for i := 0; i < w.count; i++ {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			w.runWorker(ctx)
		}()
	}
	w.wg.Wait()
}

func (w *Worker) runWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-w.jobChan:
			w.waitIfPaused(ctx)

			resp, err := w.client.GetAccrual(ctx, job.Number)

			if err != nil {
				var errRateLimit *client.ErrRateLimit
				switch {
				case errors.Is(err, client.ErrRequestExec),
					errors.Is(err, client.ErrOrderNotFound):
					w.logger.Error(err)

				case errors.As(err, &errRateLimit):
					until := time.Now().Add(errRateLimit.RetryAfter)
					w.setPause(until)
					job.RunAt = until
					w.scheduler.Schedule(job)
				default:
					w.logger.Error(err)
				}
				continue
			}
			switch resp.Status {
			case model.AccrualStatusProcessed:
				if err := w.repo.UpdateOrder(ctx, job.Number, mapAccrualStatus(resp.Status), resp.Accrual); err != nil {
					w.logger.Errorw("worker: failed to update order", "order", job.Number, "error", err)
				} else {
					w.logger.Debugw("worker: order processed", "order", job.Number, "accrual", resp.Accrual)
				}

			case model.AccrualStatusInvalid:
				if err := w.repo.UpdateOrderStatus(ctx, job.Number, mapAccrualStatus(resp.Status)); err != nil {
					w.logger.Errorw("worker: failed to update order status", "order", job.Number, "error", err)
				} else {
					w.logger.Debugw("worker: order invalid", "order", job.Number)
				}

			default:
				if err := w.repo.UpdateOrderStatus(ctx, job.Number, mapAccrualStatus(resp.Status)); err != nil {
					w.logger.Errorw("worker: failed to update order status", "order", job.Number, "error", err)
				} else {
					w.logger.Debugw("worker: order not ready, requeuing", "order", job.Number, "status", resp.Status)
					job.Attempt++
					job.RunAt = time.Now().Add(w.backoffDelay(job.Attempt))
					w.scheduler.Schedule(job)
				}

			}

		}
	}
}

func (w *Worker) LoadPendingOrders(ctx context.Context) error {
	orders, err := w.repo.GetPendingOrders(ctx)
	if err != nil {
		return err
	}

	for _, number := range orders {
		w.scheduler.Schedule(&ScheduleJob{
			Number:  number,
			RunAt:   time.Now(),
			Attempt: 0,
		})
	}
	return nil
}

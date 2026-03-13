package worker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/F3dosik/Hofermart/internal/client"
	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func newTestWorker(repo *mockRepository, c *mockAccrualClient, s *mockScheduler, jobChan chan *ScheduleJob) *Worker {
	logger, _ := zap.NewDevelopment()
	return New(repo, c, s, jobChan,
		1,
		100*time.Millisecond,
		5*time.Second,
		logger.Sugar(),
	)
}

func TestWorker_RunWorker_Processed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accrual := 500.0
	jobChan := make(chan *ScheduleJob, 1)

	repo := new(mockRepository)
	c := new(mockAccrualClient)
	s := new(mockScheduler)

	c.On("GetAccrual", mock.Anything, "12345678903").
		Return(&model.AccrualResponse{
			Order:   "12345678903",
			Status:  model.AccrualStatusProcessed,
			Accrual: &accrual,
		}, nil)

	repo.On("UpdateOrder", mock.Anything, "12345678903", model.OrderStatusProcessed, &accrual).
		Return(nil)

	w := newTestWorker(repo, c, s, jobChan)

	jobChan <- &ScheduleJob{Number: "12345678903", RunAt: time.Now(), Attempt: 0}

	go w.runWorker(ctx)

	time.Sleep(50 * time.Millisecond)
	cancel()

	repo.AssertExpectations(t)
	c.AssertExpectations(t)
}

func TestWorker_RunWorker_Invalid(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobChan := make(chan *ScheduleJob, 1)
	repo := new(mockRepository)
	c := new(mockAccrualClient)
	s := new(mockScheduler)

	c.On("GetAccrual", mock.Anything, "12345678903").
		Return(&model.AccrualResponse{
			Order:  "12345678903",
			Status: model.AccrualStatusInvalid,
		}, nil)

	repo.On("UpdateOrderStatus", mock.Anything, "12345678903", model.OrderStatusInvalid).
		Return(nil)

	w := newTestWorker(repo, c, s, jobChan)
	jobChan <- &ScheduleJob{Number: "12345678903", RunAt: time.Now()}

	go w.runWorker(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()

	repo.AssertExpectations(t)
	c.AssertExpectations(t)
}

func TestWorker_RunWorker_Processing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobChan := make(chan *ScheduleJob, 1)
	repo := new(mockRepository)
	c := new(mockAccrualClient)
	s := new(mockScheduler)

	c.On("GetAccrual", mock.Anything, "12345678903").
		Return(&model.AccrualResponse{
			Order:  "12345678903",
			Status: model.AccrualStatusProcessing,
		}, nil)

	repo.On("UpdateOrderStatus", mock.Anything, "12345678903", model.OrderStatusProcessing).
		Return(nil)

	s.On("Schedule", mock.MatchedBy(func(job *ScheduleJob) bool {
		return job.Number == "12345678903" && job.Attempt == 1
	})).Return()

	w := newTestWorker(repo, c, s, jobChan)
	jobChan <- &ScheduleJob{Number: "12345678903", RunAt: time.Now(), Attempt: 0}

	go w.runWorker(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()

	repo.AssertExpectations(t)
	c.AssertExpectations(t)
	s.AssertExpectations(t)
}

func TestWorker_RunWorker_RateLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobChan := make(chan *ScheduleJob, 1)
	repo := new(mockRepository)
	c := new(mockAccrualClient)
	s := new(mockScheduler)

	c.On("GetAccrual", mock.Anything, "12345678903").
		Return(nil, &client.ErrRateLimit{RetryAfter: time.Second})

	s.On("Schedule", mock.MatchedBy(func(job *ScheduleJob) bool {
		return job.Number == "12345678903"
	})).Return()

	w := newTestWorker(repo, c, s, jobChan)
	jobChan <- &ScheduleJob{Number: "12345678903", RunAt: time.Now()}

	go w.runWorker(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()

	c.AssertExpectations(t)
	s.AssertExpectations(t)
}

func TestWorker_RunWorker_RequestError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobChan := make(chan *ScheduleJob, 1)
	repo := new(mockRepository)
	c := new(mockAccrualClient)
	s := new(mockScheduler)

	c.On("GetAccrual", mock.Anything, "12345678903").
		Return(nil, fmt.Errorf("get accrual: %w: connection refused", client.ErrRequestExec))

	w := newTestWorker(repo, c, s, jobChan)
	jobChan <- &ScheduleJob{Number: "12345678903", RunAt: time.Now()}

	go w.runWorker(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()

	repo.AssertExpectations(t)
	s.AssertExpectations(t)
	c.AssertExpectations(t)
}

func TestWorker_RunWorker_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	jobChan := make(chan *ScheduleJob)
	repo := new(mockRepository)
	c := new(mockAccrualClient)
	s := new(mockScheduler)

	w := newTestWorker(repo, c, s, jobChan)

	done := make(chan struct{})
	go func() {
		w.runWorker(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("runWorker did not stop after context cancellation")
	}
}

func TestWorker_LoadPendingOrders(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		setupMock     func(*mockRepository, *mockScheduler)
		wantErr       bool
		wantScheduled int
	}{
		{
			name: "success: schedules all pending orders",
			setupMock: func(repo *mockRepository, s *mockScheduler) {
				repo.On("GetPendingOrders", ctx).
					Return([]string{"12345678903", "9278923470"}, nil)
				s.On("Schedule", mock.MatchedBy(func(job *ScheduleJob) bool {
					return job.Number == "12345678903" && job.Attempt == 0
				})).Return()
				s.On("Schedule", mock.MatchedBy(func(job *ScheduleJob) bool {
					return job.Number == "9278923470" && job.Attempt == 0
				})).Return()
			},
			wantScheduled: 2,
		},
		{
			name: "empty list: nothing scheduled",
			setupMock: func(repo *mockRepository, s *mockScheduler) {
				repo.On("GetPendingOrders", ctx).
					Return([]string{}, nil)
			},
		},
		{
			name: "repository error",
			setupMock: func(repo *mockRepository, s *mockScheduler) {
				repo.On("GetPendingOrders", ctx).
					Return(nil, errors.New("db is down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			s := new(mockScheduler)
			tt.setupMock(repo, s)

			jobChan := make(chan *ScheduleJob, 10)
			w := newTestWorker(repo, new(mockAccrualClient), s, jobChan)

			err := w.LoadPendingOrders(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
			s.AssertExpectations(t)
		})
	}
}

func TestWorker_BackoffDelay(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	w := New(nil, nil, nil, nil, 1,
		100*time.Millisecond,
		500*time.Millisecond,
		logger.Sugar(),
	)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 0},
		{1, 100 * time.Millisecond},
		{3, 300 * time.Millisecond},
		{10, 500 * time.Millisecond},
	}

	for _, tt := range tests {
		result := w.backoffDelay(tt.attempt)
		assert.Equal(t, tt.expected, result, "attempt=%d", tt.attempt)
	}
}

func TestWorker_SetPause(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	w := New(nil, nil, nil, nil, 1, time.Second, time.Second, logger.Sugar())

	future := time.Now().Add(time.Minute)
	w.setPause(future)
	assert.Equal(t, future.UnixNano(), w.pauseUntil.Load())

	earlier := time.Now().Add(10 * time.Second)
	w.setPause(earlier)
	assert.Equal(t, future.UnixNano(), w.pauseUntil.Load())
}

func TestWorker_WaitIfPaused_NoPause(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	w := New(nil, nil, nil, nil, 1, time.Second, time.Second, logger.Sugar())

	done := make(chan struct{})
	go func() {
		w.waitIfPaused(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("waitIfPaused blocked when no pause was set")
	}
}

func TestWorker_WaitIfPaused_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	logger, _ := zap.NewDevelopment()
	w := New(nil, nil, nil, nil, 1, time.Second, time.Second, logger.Sugar())

	w.setPause(time.Now().Add(time.Hour))

	done := make(chan struct{})
	go func() {
		w.waitIfPaused(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("waitIfPaused did not respect context cancellation")
	}
}

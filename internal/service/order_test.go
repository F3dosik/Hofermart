// service/order_service_test.go
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/F3dosik/Hofermart/internal/repository"
	"github.com/F3dosik/Hofermart/internal/worker"
	"github.com/stretchr/testify/assert"
)

func newTestOrderService(repo *mockRepository, jobChan chan *worker.ScheduleJob) OrderService {
	return NewOrderService(repo, jobChan)
}

func TestOrderService_UploadOrder(t *testing.T) {
	ctx := context.Background()
	validOrder := "12345678903"
	invalidOrder := "12345678900"

	tests := []struct {
		name        string
		orderNumber string
		setupMock   func(*mockRepository)
		wantErr     error
		wantJob     bool
	}{
		{
			name:        "success: job sent to channel",
			orderNumber: validOrder,
			setupMock: func(m *mockRepository) {
				m.On("UploadOrder", ctx, validOrder, testUserID).
					Return(nil)
			},
			wantJob: true,
		},
		{
			name:        "invalid order number: repo not called, no job",
			orderNumber: invalidOrder,
			setupMock:   func(m *mockRepository) {},
			wantErr:     ErrInvalidOrderNumber,
			wantJob:     false,
		},
		{
			name:        "order already exists for this user: no job",
			orderNumber: validOrder,
			setupMock: func(m *mockRepository) {
				m.On("UploadOrder", ctx, validOrder, testUserID).
					Return(repository.ErrOrderAlreadyExist)
			},
			wantErr: ErrOrderAlreadyExist,
			wantJob: false,
		},
		{
			name:        "order already exists for another user: no job",
			orderNumber: validOrder,
			setupMock: func(m *mockRepository) {
				m.On("UploadOrder", ctx, validOrder, testUserID).
					Return(repository.ErrOrderAlreadyExistForAnotherUser)
			},
			wantErr: ErrOrderAlreadyExistForAnotherUser,
			wantJob: false,
		},
		{
			name:        "repository unexpected error: no job",
			orderNumber: validOrder,
			setupMock: func(m *mockRepository) {
				m.On("UploadOrder", ctx, validOrder, testUserID).
					Return(errors.New("db is down"))
			},
			wantErr: errors.New("any"),
			wantJob: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			tt.setupMock(repo)

			jobChan := make(chan *worker.ScheduleJob, 1)

			svc := newTestOrderService(repo, jobChan)
			err := svc.UploadOrder(ctx, tt.orderNumber, testUserID)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, ErrInvalidOrderNumber) ||
					errors.Is(tt.wantErr, ErrOrderAlreadyExist) ||
					errors.Is(tt.wantErr, ErrOrderAlreadyExistForAnotherUser) {
					assert.ErrorIs(t, err, tt.wantErr)
				}
			}

			if tt.wantJob {
				select {
				case job := <-jobChan:
					assert.Equal(t, tt.orderNumber, job.Number)
					assert.Equal(t, 0, job.Attempt)
					assert.WithinDuration(t, time.Now(), job.RunAt, time.Second)
				default:
					t.Error("expected job in channel, but channel is empty")
				}
			} else {
				select {
				case job := <-jobChan:
					t.Errorf("expected no job in channel, but got job for order %s", job.Number)
				default:
				}
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestOrderService_GetOrders(t *testing.T) {
	ctx := context.Background()

	accrual := 500.0
	testOrders := []*model.Order{
		{Number: "12345678903", Status: model.OrderStatusProcessed, Accrual: &accrual},
		{Number: "9278923470", Status: model.OrderStatusNew},
	}

	tests := []struct {
		name       string
		setupMock  func(*mockRepository)
		wantOrders []*model.Order
		wantErr    error
	}{
		{
			name: "success with orders",
			setupMock: func(m *mockRepository) {
				m.On("GetOrders", ctx, testUserID).
					Return(testOrders, nil)
			},
			wantOrders: testOrders,
		},
		{
			name: "empty list",
			setupMock: func(m *mockRepository) {
				m.On("GetOrders", ctx, testUserID).
					Return([]*model.Order{}, nil)
			},
			wantOrders: []*model.Order{},
		},
		{
			name: "repository unexpected error",
			setupMock: func(m *mockRepository) {
				m.On("GetOrders", ctx, testUserID).
					Return(nil, errors.New("db is down"))
			},
			wantErr: errors.New("any"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			tt.setupMock(repo)

			jobChan := make(chan *worker.ScheduleJob, 1)
			svc := newTestOrderService(repo, jobChan)
			orders, err := svc.GetOrders(ctx, testUserID)

			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOrders, orders)
			} else {
				assert.Error(t, err)
				assert.Nil(t, orders)
			}

			repo.AssertExpectations(t)
		})
	}
}

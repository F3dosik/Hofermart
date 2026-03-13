package worker

import (
	"context"

	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) UpdateOrder(ctx context.Context, number string, status model.OrderStatus, accrual *float64) error {
	args := m.Called(ctx, number, status, accrual)
	return args.Error(0)
}

func (m *mockRepository) UpdateOrderStatus(ctx context.Context, number string, status model.OrderStatus) error {
	args := m.Called(ctx, number, status)
	return args.Error(0)
}

func (m *mockRepository) GetPendingOrders(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockRepository) CreateUserWithBalance(ctx context.Context, login, password string) (*model.User, error) {
	return nil, nil
}
func (m *mockRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	return nil, nil
}
func (m *mockRepository) UploadOrder(ctx context.Context, number string, userID uuid.UUID) error {
	return nil
}
func (m *mockRepository) GetOrders(ctx context.Context, userID uuid.UUID) ([]*model.Order, error) {
	return nil, nil
}
func (m *mockRepository) GetBalance(ctx context.Context, userID uuid.UUID) (*model.Balance, error) {
	return nil, nil
}
func (m *mockRepository) CreateWithdrawal(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) error {
	return nil
}
func (m *mockRepository) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]*model.Withdrawal, error) {
	return nil, nil
}
func (m *mockRepository) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return nil
}

type mockAccrualClient struct {
	mock.Mock
}

func (m *mockAccrualClient) GetAccrual(ctx context.Context, orderNumber string) (*model.AccrualResponse, error) {
	args := m.Called(ctx, orderNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AccrualResponse), args.Error(1)
}

type mockScheduler struct {
	mock.Mock
}

func (m *mockScheduler) Schedule(job *ScheduleJob) {
	m.Called(job)
}

func (m *mockScheduler) Run(ctx context.Context) {
	m.Called(ctx)
}

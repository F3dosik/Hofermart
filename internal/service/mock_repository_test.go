package service

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

func newTestUserService(repo *mockRepository) UserService {
	return NewUserService(repo, "test-secret-key")
}

func (m *mockRepository) CreateUserWithBalance(ctx context.Context, login, hash string) (*model.User, error) {
	args := m.Called(ctx, login, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockRepository) UploadOrder(ctx context.Context, number string, userID uuid.UUID) error {
	args := m.Called(ctx, number, userID)
	return args.Error(0)
}

func (m *mockRepository) GetOrders(ctx context.Context, userID uuid.UUID) ([]*model.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
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

func (m *mockRepository) GetBalance(ctx context.Context, userID uuid.UUID) (*model.Balance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Balance), args.Error(1)
}

func (m *mockRepository) CreateWithdrawal(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) error {
	args := m.Called(ctx, userID, orderNumber, sum)
	return args.Error(0)
}

func (m *mockRepository) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]*model.Withdrawal, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Withdrawal), args.Error(1)
}

func (m *mockRepository) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

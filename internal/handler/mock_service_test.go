package handler

import (
	"context"

	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Register(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}

func (m *mockUserService) Login(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}

type mockOrderService struct {
	mock.Mock
}

func (m *mockOrderService) UploadOrder(ctx context.Context, number string, userID uuid.UUID) error {
	args := m.Called(ctx, number, userID)
	return args.Error(0)
}

func (m *mockOrderService) GetOrders(ctx context.Context, userID uuid.UUID) ([]*model.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

type mockBalanceService struct {
	mock.Mock
}

func (m *mockBalanceService) GetBalance(ctx context.Context, userID uuid.UUID) (*model.Balance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Balance), args.Error(1)
}

func (m *mockBalanceService) CreateWithdrawal(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) error {
	args := m.Called(ctx, userID, orderNumber, sum)
	return args.Error(0)
}

func (m *mockBalanceService) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]*model.Withdrawal, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Withdrawal), args.Error(1)
}

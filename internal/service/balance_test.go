// service/balance_service_test.go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/F3dosik/Hofermart/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func newTestBalanceService(repo *mockRepository) BalanceService {
	return NewBalanceService(repo)
}

var testUserID = uuid.New()

func TestBalanceService_GetBalance(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		setupMock   func(*mockRepository)
		wantBalance *model.Balance
		wantErr     error
	}{
		{
			name: "success",
			setupMock: func(m *mockRepository) {
				m.On("GetBalance", ctx, testUserID).
					Return(&model.Balance{Current: 100.5, Withdrawn: 42}, nil)
			},
			wantBalance: &model.Balance{Current: 100.5, Withdrawn: 42},
		},
		{
			name: "balance not found",
			setupMock: func(m *mockRepository) {
				m.On("GetBalance", ctx, testUserID).
					Return(nil, pgx.ErrNoRows)
			},
			wantErr: ErrBalanceNotFound,
		},
		{
			name: "repository unexpected error",
			setupMock: func(m *mockRepository) {
				m.On("GetBalance", ctx, testUserID).
					Return(nil, errors.New("db is down"))
			},
			wantErr: errors.New("any"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			tt.setupMock(repo)

			svc := newTestBalanceService(repo)
			balance, err := svc.GetBalance(ctx, testUserID)

			if tt.wantBalance != nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBalance.Current, balance.Current)
				assert.Equal(t, tt.wantBalance.Withdrawn, balance.Withdrawn)
			} else {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, ErrBalanceNotFound) {
					assert.ErrorIs(t, err, ErrBalanceNotFound)
				}
				assert.Nil(t, balance)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestBalanceService_CreateWithdrawal(t *testing.T) {
	ctx := context.Background()
	validOrder := "12345678903"
	invalidOrder := "12345678904"

	tests := []struct {
		name        string
		orderNumber string
		sum         float64
		setupMock   func(*mockRepository)
		wantErr     error
	}{
		{
			name:        "success",
			orderNumber: validOrder,
			sum:         100,
			setupMock: func(m *mockRepository) {
				m.On("CreateWithdrawal", ctx, testUserID, validOrder, 100.0).
					Return(nil)
			},
		},
		{
			name:        "invalid order number",
			orderNumber: invalidOrder,
			sum:         100,
			setupMock:   func(m *mockRepository) {},
			wantErr:     ErrInvalidOrderNumber,
		},
		{
			name:        "not enough balance",
			orderNumber: validOrder,
			sum:         9999,
			setupMock: func(m *mockRepository) {
				m.On("CreateWithdrawal", ctx, testUserID, validOrder, 9999.0).
					Return(repository.ErrNotEnoughBalance)
			},
			wantErr: ErrNotEnoughBalance,
		},
		{
			name:        "repository unexpected error",
			orderNumber: validOrder,
			sum:         100,
			setupMock: func(m *mockRepository) {
				m.On("CreateWithdrawal", ctx, testUserID, validOrder, 100.0).
					Return(errors.New("db is down"))
			},
			wantErr: errors.New("any"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			tt.setupMock(repo)

			svc := newTestBalanceService(repo)
			err := svc.CreateWithdrawal(ctx, testUserID, tt.orderNumber, tt.sum)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, ErrInvalidOrderNumber) ||
					errors.Is(tt.wantErr, ErrNotEnoughBalance) {
					assert.ErrorIs(t, err, tt.wantErr)
				}
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestBalanceService_GetWithdrawals(t *testing.T) {
	ctx := context.Background()

	testWithdrawals := []*model.Withdrawal{
		{OrderNumber: "12345678903", Sum: 100},
		{OrderNumber: "9278923470", Sum: 50},
	}

	tests := []struct {
		name            string
		setupMock       func(*mockRepository)
		wantWithdrawals []*model.Withdrawal
		wantErr         error
	}{
		{
			name: "success with withdrawals",
			setupMock: func(m *mockRepository) {
				m.On("GetWithdrawals", ctx, testUserID).
					Return(testWithdrawals, nil)
			},
			wantWithdrawals: testWithdrawals,
		},
		{
			name: "empty list",
			setupMock: func(m *mockRepository) {
				m.On("GetWithdrawals", ctx, testUserID).
					Return([]*model.Withdrawal{}, nil)
			},
			wantWithdrawals: []*model.Withdrawal{},
		},
		{
			name: "repository unexpected error",
			setupMock: func(m *mockRepository) {
				m.On("GetWithdrawals", ctx, testUserID).
					Return(nil, errors.New("db is down"))
			},
			wantErr: errors.New("any"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			tt.setupMock(repo)

			svc := newTestBalanceService(repo)
			withdrawals, err := svc.GetWithdrawals(ctx, testUserID)

			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantWithdrawals, withdrawals)
			} else {
				assert.Error(t, err)
				assert.Nil(t, withdrawals)
			}

			repo.AssertExpectations(t)
		})
	}
}

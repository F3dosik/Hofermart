package service

import (
	"context"
	"errors"

	"github.com/F3dosik/Hofermart/internal/db"
	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/F3dosik/Hofermart/internal/repository"
	"github.com/google/uuid"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (*model.Balance, error)
	CreateWithdrawal(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) error
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]*model.Withdrawal, error)
}

type balanceService struct {
	repository repository.Repository
}

func NewBalanceService(repo repository.Repository) BalanceService {
	return &balanceService{repository: repo}
}

func (s *balanceService) GetBalance(ctx context.Context, userID uuid.UUID) (*model.Balance, error) {
	balance, err := s.repository.GetBalance(ctx, userID)
	if err != nil {
		if db.IsNoRows(err) {
			return nil, ErrBalanceNotFound
		}
		return nil, err
	}
	return balance, nil
}

func (s *balanceService) CreateWithdrawal(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) error {
	if err := validateOrder(orderNumber); err != nil {
		return err
	}

	if err := s.repository.CreateWithdrawal(ctx, userID, orderNumber, sum); err != nil {
		if errors.Is(err, repository.ErrNotEnoughBalance) {
			return ErrNotEnoughBalance
		}
		return err
	}

	return nil
}

func (s *balanceService) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]*model.Withdrawal, error) {
	withdrawals, err := s.repository.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}

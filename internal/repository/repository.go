package repository

import (
	"context"
	"fmt"

	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateUserWithBalance(ctx context.Context, login, password string) (*model.User, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)

	UploadOrder(ctx context.Context, number string, userID uuid.UUID) error
	GetOrders(ctx context.Context, userID uuid.UUID) ([]*model.Order, error)
	UpdateOrder(ctx context.Context, number string, status model.OrderStatus, accrual *float64) error
	UpdateOrderStatus(ctx context.Context, number string, status model.OrderStatus) error
	GetPendingOrders(ctx context.Context) ([]string, error)

	GetBalance(ctx context.Context, userID uuid.UUID) (*model.Balance, error)

	CreateWithdrawal(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) error
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]*model.Withdrawal, error)

	WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type postgresRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

package repository

import (
	"context"
	"fmt"

	"github.com/F3dosik/Hofermart/internal/db"
	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *postgresRepository) createBalance(ctx context.Context, q db.Querier, userID uuid.UUID) error {
	_, err := q.Exec(ctx, `
			INSERT INTO balance (user_id)
			VALUES ($1)
		`, userID)

	if err != nil {
		return fmt.Errorf("create balance: %w", err)
	}

	return nil
}

func (r *postgresRepository) GetBalance(ctx context.Context, userID uuid.UUID) (*model.Balance, error) {
	var balance model.Balance
	err := r.pool.QueryRow(ctx, `
		SELECT current, withdrawn FROM balance
		WHERE user_id = $1
	`, userID).Scan(&balance.Current, &balance.Withdrawn)

	if err != nil {
		return nil, err
	}

	return &balance, nil
}

func (r *postgresRepository) CreateWithdrawal(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) error {
	return db.WithRetry(ctx, func() error {
		return r.WithTx(ctx, func(tx pgx.Tx) error {
			if err := r.updateBalance(ctx, tx, userID, sum); err != nil {
				return err
			}
			return r.insertWithdrawal(ctx, tx, userID, orderNumber, sum)
		})
	})
}

func (r *postgresRepository) updateBalance(ctx context.Context, q db.Querier, userID uuid.UUID, sum float64) error {
	result, err := q.Exec(ctx, `
		UPDATE balance
		SET current = current - $1,
			withdrawn = withdrawn + $1,
			updated_at = now()
		WHERE user_id = $2
			AND current >= $1
	`, sum, userID)

	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotEnoughBalance
	}

	return nil
}

func (r *postgresRepository) insertWithdrawal(ctx context.Context, q db.Querier, userID uuid.UUID, orderNumber string, sum float64) error {
	_, err := q.Exec(ctx, `
		INSERT INTO withdrawals (user_id, order_number, sum)
		VALUES ($1, $2, $3)
	`, userID, orderNumber, sum)

	if err != nil {
		return fmt.Errorf("insert withdrawal: %w", err)
	}

	return nil
}

func (r *postgresRepository) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]*model.Withdrawal, error) {
	var withdrawals []*model.Withdrawal
	err := db.WithRetry(ctx, func() error {
		rows, err := r.pool.Query(ctx, `
			SELECT order_number, sum, processed_at
			FROM withdrawals
			WHERE user_id = $1
			ORDER BY processed_at DESC
		`, userID)

		if err != nil {
			return fmt.Errorf("query withdrawals: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var withdraw model.Withdrawal
			if err := rows.Scan(
				&withdraw.OrderNumber,
				&withdraw.Sum,
				&withdraw.ProcessedAt,
			); err != nil {
				return fmt.Errorf("scan withdrawal: %w", err)
			}

			withdrawals = append(withdrawals, &withdraw)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

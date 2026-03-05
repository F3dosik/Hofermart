package repository

import (
	"context"
	"fmt"

	"github.com/F3dosik/Hofermart/internal/db"
	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *postgresRepository) UploadOrder(ctx context.Context, number string, userID uuid.UUID) error {
	var existingUserID uuid.UUID
	var inserted bool
	err := db.WithRetry(ctx, func() error {
		return r.pool.QueryRow(ctx, `
			WITH ins AS (
				INSERT INTO orders (user_id, number)
				VALUES ($1, $2)
				ON CONFLICT (number) DO NOTHING
				RETURNING user_id
			)
			SELECT user_id, true AS inserted FROM ins
			UNION ALL 
			SELECT user_id, false FROM orders WHERE number = $2
			LIMIT 1
		`, userID, number).Scan(&existingUserID, &inserted)
	})

	if err != nil {
		return fmt.Errorf("upload order: %w", err)
	}

	if inserted {
		return nil
	}

	if existingUserID != userID {
		return ErrOrderAlreadyExistForAnotherUser
	}
	return ErrOrderAlreadyExist
}

func (r *postgresRepository) GetOrders(ctx context.Context, userID uuid.UUID) ([]*model.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT number, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

func (r *postgresRepository) UpdateOrder(ctx context.Context, number string, status model.OrderStatus, accrual *float64) error {
	return db.WithRetry(ctx, func() error {
		return r.WithTx(ctx, func(tx pgx.Tx) error {
			userID, err := r.updateOrder(ctx, tx, number, status, accrual)
			if err != nil {
				return err
			}
			return r.accrualBalance(ctx, tx, userID, accrual)
		})
	})
}

func (r *postgresRepository) updateOrder(ctx context.Context, q db.Querier, number string, status model.OrderStatus, accrual *float64) (uuid.UUID, error) {
	var userID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE orders
		SET status = $1,
			accrual = $2
		WHERE number = $3
		RETURNING user_id
	`, status, accrual, number).Scan(&userID)

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("update order: %w", err)
	}

	return userID, nil
}

func (r *postgresRepository) accrualBalance(ctx context.Context, q db.Querier, userID uuid.UUID, accrual *float64) error {
	_, err := q.Exec(ctx, `
		UPDATE balance 
		SET current = current + $1, 
			updated_at = now()
		WHERE user_id = $2
	`, accrual, userID)
	if err != nil {
		return fmt.Errorf("accrual balance: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateOrderStatus(ctx context.Context, number string, status model.OrderStatus) error {
	return db.WithRetry(ctx, func() error {
		_, err := r.pool.Exec(ctx, `
		UPDATE orders
		SET status = $1
		WHERE number = $2
	`, status, number)
		if err != nil {
			return fmt.Errorf("update order status: %w", err)
		}
		return nil
	})
}

func (r *postgresRepository) GetPendingOrders(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT number FROM orders
		WHERE status IN ('NEW', 'PROCESSING')
	`)
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()
	var orders []string
	for rows.Next() {
		var order string
		if err := rows.Scan(&order); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

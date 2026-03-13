package repository

import (
	"context"
	"fmt"

	"github.com/F3dosik/Hofermart/internal/db"
	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/jackc/pgx/v5"
)

func (r *postgresRepository) createUser(ctx context.Context, q db.Querier, login, password string) (*model.User, error) {
	var user model.User
	err := q.QueryRow(ctx, `
			INSERT INTO users (login, password)
			VALUES ($1, $2)
			RETURNING id, created_at
		`, login, password).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrLoginAlreadyExist
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	user.Login = login
	return &user, nil
}

func (r *postgresRepository) CreateUserWithBalance(ctx context.Context, login, password string) (*model.User, error) {
	var user *model.User

	err := db.WithRetry(ctx, func() error {
		return r.WithTx(ctx, func(tx pgx.Tx) error {
			var err error
			user, err = r.createUser(ctx, tx, login, password)
			if err != nil {
				return err
			}
			return r.createBalance(ctx, tx, user.ID)
		})
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *postgresRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	user := model.User{Login: login}
	err := db.WithRetry(ctx, func() error {
		row := r.pool.QueryRow(ctx, `
			SELECT id, password, created_at FROM users
			WHERE login = $1
		`, login)

		return row.Scan(&user.ID, &user.Password, &user.CreatedAt)
	})

	if err != nil {
		if db.IsNoRows(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by login: %w", err)
	}

	return &user, nil
}

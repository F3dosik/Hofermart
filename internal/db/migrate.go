package db

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"

	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(dsn string, logger *zap.SugaredLogger) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("iofs source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return fmt.Errorf("create migrate (dsn = %s): %w", dsn, err)
	}

	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil || dbErr != nil {
			logger.Warnw("migrate close:",
				"source error", zap.Error(srcErr),
				"database error", dbErr,
			)
		}
	}()

	err = WithRetry(context.Background(), func() error {
		if e := m.Up(); e != nil && !errors.Is(e, migrate.ErrNoChange) {
			return e
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("migrations failed: %w", err)
	}

	return nil
}

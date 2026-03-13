package model

import (
	"time"

	uuid "github.com/jackc/pgx/pgtype/ext/gofrs-uuid"
)

type Balance struct {
	UserID    uuid.UUID
	Current   float64
	Withdrawn float64
	UpdatedAt time.Time
}

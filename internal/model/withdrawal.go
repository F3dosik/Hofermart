package model

import (
	"time"

	uuid "github.com/jackc/pgx/pgtype/ext/gofrs-uuid"
)

type Withdrawal struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}

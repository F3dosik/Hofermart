package model

import (
	"time"

	uuid "github.com/jackc/pgx/pgtype/ext/gofrs-uuid"
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Number     string
	Status     OrderStatus
	Accrual    *float64
	UploadedAt time.Time
}

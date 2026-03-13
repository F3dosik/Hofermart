package model

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        uuid.UUID
	Login     string
	Password  string
	CreatedAt time.Time
}

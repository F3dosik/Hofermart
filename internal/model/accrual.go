package model

type AccrualStatus string

const (
	AccrualStatusRegistered AccrualStatus = "REGISTERED"
	AccrualStatusProcessing AccrualStatus = "PROCESSING"
	AccrualStatusInvalid    AccrualStatus = "INVALID"
	AccrualStatusProcessed  AccrualStatus = "PROCESSED"
)

type AccrualResponse struct {
	Order   string        `json:"order"`
	Status  AccrualStatus `json:"status"`
	Accrual *float64      `json:"accrual,omitempty"`
}

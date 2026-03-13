package worker

import "github.com/F3dosik/Hofermart/internal/model"

func mapAccrualStatus(status model.AccrualStatus) model.OrderStatus {
	switch status {
	case model.AccrualStatusRegistered:
		return model.OrderStatusProcessing
	case model.AccrualStatusProcessing:
		return model.OrderStatusProcessing
	case model.AccrualStatusInvalid:
		return model.OrderStatusInvalid
	case model.AccrualStatusProcessed:
		return model.OrderStatusProcessed
	default:
		return model.OrderStatusProcessing
	}
}

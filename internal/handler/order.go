package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/F3dosik/Hofermart/internal/ctxkey"
	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/F3dosik/Hofermart/internal/service"
	"github.com/google/uuid"
)

func parseOrder(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	number := strings.TrimSpace(string(body))
	if number == "" {
		return "", errors.New("empty order number")
	}
	return number, nil
}

func (h *Handler) uploadOrder(w http.ResponseWriter, r *http.Request) {
	number, err := parseOrder(r)
	if err != nil {
		h.logger.Errorw("create order", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(ctxkey.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Error("upload order: can't take userID")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err = h.orderService.UploadOrder(r.Context(), number, userID); err != nil {
		switch {
		case errors.Is(err, service.ErrOrderAlreadyExist):
			h.logger.Debugw("upload order", "error", err)
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, service.ErrOrderAlreadyExistForAnotherUser):
			h.logger.Debugw("upload order", "error", err)
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		case errors.Is(err, service.ErrInvalidOrderNumber):
			h.logger.Debugw("upload order", "error", err)
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		default:
			h.logger.Errorw("upload order: internal error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

type orderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

func newOrderResponse(o *model.Order) orderResponse {
	return orderResponse{
		Number:     o.Number,
		Status:     string(o.Status),
		Accrual:    o.Accrual,
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
	}
}

func (h *Handler) getOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Debug("get order: can't take userID")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	orders, err := h.orderService.GetOrders(r.Context(), userID)
	if err != nil {
		h.logger.Errorw("get orders: internal error", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ordersResponse := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		ordersResponse = append(ordersResponse, newOrderResponse(order))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(ordersResponse); err != nil {
		h.logger.Errorw("get orders: encode error", "error", err)
		return
	}
}

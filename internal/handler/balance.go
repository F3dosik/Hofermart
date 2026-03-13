package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/F3dosik/Hofermart/internal/ctxkey"
	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/F3dosik/Hofermart/internal/service"
	"github.com/google/uuid"
)

type balanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func newBalanceResponse(b *model.Balance) *balanceResponse {
	return &balanceResponse{
		Current:   b.Current,
		Withdrawn: b.Withdrawn,
	}
}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Error("get balance: can't take userID")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.Errorw("get balance: internal error", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	br := newBalanceResponse(balance)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(br); err != nil {
		h.logger.Errorw("get balance: encode error", "error", err)
		return
	}

}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *Handler) withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Error("withdraw: can't take userID")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var withdraw withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&withdraw); err != nil {
		h.logger.Debugw("can't decode JSON body", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := h.balanceService.CreateWithdrawal(r.Context(), userID, withdraw.Order, withdraw.Sum); err != nil {
		switch {
		case errors.Is(err, service.ErrNotEnoughBalance):
			h.logger.Debugw("withdraw", "error", err)
			http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
		case errors.Is(err, service.ErrInvalidOrderNumber):
			h.logger.Debugw("withdraw", "error", err)
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		default:
			h.logger.Errorw("withdraw: internal error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return

	}

	w.WriteHeader(http.StatusOK)
}

type withdrawResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func newWithdrawResponse(w *model.Withdrawal) withdrawResponse {
	return withdrawResponse{
		Order:       w.OrderNumber,
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
	}
}

func (h *Handler) getWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Error("get withdrawals: can't take userID")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	withdrawals, err := h.balanceService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		switch {
		default:
			h.logger.Errorw("get withdrawals: internal error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	if len(withdrawals) == 0 {
		h.logger.Debugw("get withdrawals: no content")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	wr := make([]withdrawResponse, 0, len(withdrawals))
	for _, withdraw := range withdrawals {
		wr = append(wr, newWithdrawResponse(withdraw))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(wr); err != nil {
		h.logger.Errorw("get withdrawals: encode error", "error", err)
		return
	}
}

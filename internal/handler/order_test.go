// handler/order_test.go
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/F3dosik/Hofermart/internal/ctxkey"
	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/F3dosik/Hofermart/internal/service"
)

var testHandlerUserID = uuid.New()

func injectUserID(r *http.Request, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), ctxkey.UserIDKey, userID)
	return r.WithContext(ctx)
}

func TestHandler_UploadOrder(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		injectUser bool
		setupMock  func(*mockOrderService)
		wantStatus int
	}{
		{
			name:       "success: 202",
			body:       "12345678903",
			injectUser: true,
			setupMock: func(m *mockOrderService) {
				m.On("UploadOrder", mock.Anything, "12345678903", testHandlerUserID).
					Return(nil)
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "order already exists for this user: 200",
			body:       "12345678903",
			injectUser: true,
			setupMock: func(m *mockOrderService) {
				m.On("UploadOrder", mock.Anything, "12345678903", testHandlerUserID).
					Return(service.ErrOrderAlreadyExist)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "order already exists for another user: 409",
			body:       "12345678903",
			injectUser: true,
			setupMock: func(m *mockOrderService) {
				m.On("UploadOrder", mock.Anything, "12345678903", testHandlerUserID).
					Return(service.ErrOrderAlreadyExistForAnotherUser)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "invalid order number: 422",
			body:       "12345678903",
			injectUser: true,
			setupMock: func(m *mockOrderService) {
				m.On("UploadOrder", mock.Anything, "12345678903", testHandlerUserID).
					Return(service.ErrInvalidOrderNumber)
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "empty body: 400",
			body:       "",
			injectUser: true,
			setupMock:  func(m *mockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no userID in context: 500",
			body:       "12345678903",
			injectUser: false,
			setupMock:  func(m *mockOrderService) {},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "internal error: 500",
			body:       "12345678903",
			injectUser: true,
			setupMock: func(m *mockOrderService) {
				m.On("UploadOrder", mock.Anything, "12345678903", testHandlerUserID).
					Return(errors.New("db is down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderSvc := new(mockOrderService)
			tt.setupMock(orderSvc)

			h := newTestHandler(new(mockUserService), orderSvc, new(mockBalanceService))

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders",
				bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "text/plain")

			if tt.injectUser {
				req = injectUserID(req, testHandlerUserID)
			}

			rr := httptest.NewRecorder()
			h.uploadOrder(rr, req) // вызываем напрямую, минуя middleware

			assert.Equal(t, tt.wantStatus, rr.Code)
			orderSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_GetOrders(t *testing.T) {
	accrual := 500.0
	uploadedAt := time.Now().Truncate(time.Second)

	testOrders := []*model.Order{
		{
			Number:     "12345678903",
			Status:     model.OrderStatusProcessed,
			Accrual:    &accrual,
			UploadedAt: uploadedAt,
		},
		{
			Number:     "9278923470",
			Status:     model.OrderStatusNew,
			UploadedAt: uploadedAt,
		},
	}

	tests := []struct {
		name       string
		injectUser bool
		setupMock  func(*mockOrderService)
		wantStatus int
		checkBody  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "success with orders: 200",
			injectUser: true,
			setupMock: func(m *mockOrderService) {
				m.On("GetOrders", mock.Anything, testHandlerUserID).
					Return(testOrders, nil)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response []orderResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Len(t, response, 2)
				assert.Equal(t, "12345678903", response[0].Number)
				assert.Equal(t, "PROCESSED", response[0].Status)
				assert.Equal(t, &accrual, response[0].Accrual)
				assert.Equal(t, "9278923470", response[1].Number)
				assert.Equal(t, "NEW", response[1].Status)
				assert.Nil(t, response[1].Accrual)
			},
		},
		{
			name:       "empty orders: 204",
			injectUser: true,
			setupMock: func(m *mockOrderService) {
				m.On("GetOrders", mock.Anything, testHandlerUserID).
					Return([]*model.Order{}, nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "no userID in context: 500",
			injectUser: false,
			setupMock:  func(m *mockOrderService) {},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "internal error: 500",
			injectUser: true,
			setupMock: func(m *mockOrderService) {
				m.On("GetOrders", mock.Anything, testHandlerUserID).
					Return(nil, errors.New("db is down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderSvc := new(mockOrderService)
			tt.setupMock(orderSvc)

			h := newTestHandler(new(mockUserService), orderSvc, new(mockBalanceService))

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			if tt.injectUser {
				req = injectUserID(req, testHandlerUserID)
			}

			rr := httptest.NewRecorder()
			h.getOrders(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.checkBody != nil {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				tt.checkBody(t, rr)
			}

			orderSvc.AssertExpectations(t)
		})
	}
}

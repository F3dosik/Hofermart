package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/F3dosik/Hofermart/internal/service"
)

func TestHandler_GetBalance(t *testing.T) {
	tests := []struct {
		name       string
		injectUser bool
		setupMock  func(*mockBalanceService)
		wantStatus int
		checkBody  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "success: 200",
			injectUser: true,
			setupMock: func(m *mockBalanceService) {
				m.On("GetBalance", mock.Anything, testHandlerUserID).
					Return(&model.Balance{Current: 100.5, Withdrawn: 42}, nil)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response balanceResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, 100.5, response.Current)
				assert.Equal(t, 42.0, response.Withdrawn)
			},
		},
		{
			name:       "no userID in context: 500",
			injectUser: false,
			setupMock:  func(m *mockBalanceService) {},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "internal error: 500",
			injectUser: true,
			setupMock: func(m *mockBalanceService) {
				m.On("GetBalance", mock.Anything, testHandlerUserID).
					Return(nil, errors.New("db is down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balanceSvc := new(mockBalanceService)
			tt.setupMock(balanceSvc)

			h := newTestHandler(new(mockUserService), new(mockOrderService), balanceSvc)

			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			if tt.injectUser {
				req = injectUserID(req, testHandlerUserID)
			}

			rr := httptest.NewRecorder()
			h.getBalance(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.checkBody != nil {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				tt.checkBody(t, rr)
			}

			balanceSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Withdraw(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		injectUser bool
		setupMock  func(*mockBalanceService)
		wantStatus int
	}{
		{
			name:       "success: 200",
			body:       `{"order":"12345678903","sum":100}`,
			injectUser: true,
			setupMock: func(m *mockBalanceService) {
				m.On("CreateWithdrawal", mock.Anything, testHandlerUserID, "12345678903", 100.0).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json: 400",
			body:       `not-json`,
			injectUser: true,
			setupMock:  func(m *mockBalanceService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not enough balance: 402",
			body:       `{"order":"12345678903","sum":9999}`,
			injectUser: true,
			setupMock: func(m *mockBalanceService) {
				m.On("CreateWithdrawal", mock.Anything, testHandlerUserID, "12345678903", 9999.0).
					Return(service.ErrNotEnoughBalance)
			},
			wantStatus: http.StatusPaymentRequired,
		},
		{
			name:       "invalid order number: 422",
			body:       `{"order":"00000000001","sum":100}`,
			injectUser: true,
			setupMock: func(m *mockBalanceService) {
				m.On("CreateWithdrawal", mock.Anything, testHandlerUserID, "00000000001", 100.0).
					Return(service.ErrInvalidOrderNumber)
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "no userID in context: 500",
			body:       `{"order":"12345678903","sum":100}`,
			injectUser: false,
			setupMock:  func(m *mockBalanceService) {},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "internal error: 500",
			body:       `{"order":"12345678903","sum":100}`,
			injectUser: true,
			setupMock: func(m *mockBalanceService) {
				m.On("CreateWithdrawal", mock.Anything, testHandlerUserID, "12345678903", 100.0).
					Return(errors.New("db is down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balanceSvc := new(mockBalanceService)
			tt.setupMock(balanceSvc)

			h := newTestHandler(new(mockUserService), new(mockOrderService), balanceSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw",
				bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.injectUser {
				req = injectUserID(req, testHandlerUserID)
			}

			rr := httptest.NewRecorder()
			h.withdraw(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			balanceSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_GetWithdrawals(t *testing.T) {
	processedAt := time.Now().Truncate(time.Second)

	testWithdrawals := []*model.Withdrawal{
		{OrderNumber: "12345678903", Sum: 100, ProcessedAt: processedAt},
		{OrderNumber: "9278923470", Sum: 50, ProcessedAt: processedAt},
	}

	tests := []struct {
		name       string
		injectUser bool
		setupMock  func(*mockBalanceService)
		wantStatus int
		checkBody  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "success with withdrawals: 200",
			injectUser: true,
			setupMock: func(m *mockBalanceService) {
				m.On("GetWithdrawals", mock.Anything, testHandlerUserID).
					Return(testWithdrawals, nil)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response []withdrawResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Len(t, response, 2)
				assert.Equal(t, "12345678903", response[0].Order)
				assert.Equal(t, 100.0, response[0].Sum)
				assert.Equal(t, processedAt.Format(time.RFC3339), response[0].ProcessedAt)
			},
		},
		{
			name:       "empty withdrawals: 204",
			injectUser: true,
			setupMock: func(m *mockBalanceService) {
				m.On("GetWithdrawals", mock.Anything, testHandlerUserID).
					Return([]*model.Withdrawal{}, nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "no userID in context: 500",
			injectUser: false,
			setupMock:  func(m *mockBalanceService) {},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "internal error: 500",
			injectUser: true,
			setupMock: func(m *mockBalanceService) {
				m.On("GetWithdrawals", mock.Anything, testHandlerUserID).
					Return(nil, errors.New("db is down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balanceSvc := new(mockBalanceService)
			tt.setupMock(balanceSvc)

			h := newTestHandler(new(mockUserService), new(mockOrderService), balanceSvc)

			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			if tt.injectUser {
				req = injectUserID(req, testHandlerUserID)
			}

			rr := httptest.NewRecorder()
			h.getWithdrawals(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.checkBody != nil {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				tt.checkBody(t, rr)
			}

			balanceSvc.AssertExpectations(t)
		})
	}
}

package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/F3dosik/Hofermart/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

const testSecretKey = "test-secret-key"

func newTestHandler(userSvc *mockUserService, orderSvc *mockOrderService, balanceSvc *mockBalanceService) *Handler {
	logger, _ := zap.NewDevelopment()
	return New(userSvc, orderSvc, balanceSvc, testSecretKey, logger.Sugar())
}

func TestHandler_Register(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func(*mockUserService)
		wantStatus int
		wantCookie bool
	}{
		{
			name: "success: 200 + cookie",
			body: `{"login":"alice","password":"secret123"}`,
			setupMock: func(m *mockUserService) {
				m.On("Register", mock.Anything, "alice", "secret123").
					Return("token_abc", nil)
			},
			wantStatus: http.StatusOK,
			wantCookie: true,
		},
		{
			name:       "invalid json: 400",
			body:       `not-json`,
			setupMock:  func(m *mockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty login: 400",
			body: `{"login":"","password":"secret123"}`,
			setupMock: func(m *mockUserService) {
				m.On("Register", mock.Anything, "", "secret123").
					Return("", service.ErrEmptyLogin)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "password too short: 400",
			body: `{"login":"alice","password":"x"}`,
			setupMock: func(m *mockUserService) {
				m.On("Register", mock.Anything, "alice", "x").
					Return("", service.ErrPasswordTooShort)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "login taken: 409",
			body: `{"login":"alice","password":"secret123"}`,
			setupMock: func(m *mockUserService) {
				m.On("Register", mock.Anything, "alice", "secret123").
					Return("", service.ErrLoginAlreadyExist)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "internal error: 500",
			body: `{"login":"alice","password":"secret123"}`,
			setupMock: func(m *mockUserService) {
				m.On("Register", mock.Anything, "alice", "secret123").
					Return("", errors.New("db is down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userSvc := new(mockUserService)
			tt.setupMock(userSvc)

			h := newTestHandler(userSvc, new(mockOrderService), new(mockBalanceService))

			req := httptest.NewRequest(http.MethodPost, "/api/user/register",
				bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantCookie {
				resp := rr.Result()
				defer resp.Body.Close()
				cookies := resp.Cookies()
				found := false
				for _, c := range cookies {
					if c.Name == "token" {
						found = true
						assert.NotEmpty(t, c.Value)
					}
				}
				assert.True(t, found, "cookie 'token' not found")
			}

			userSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func(*mockUserService)
		wantStatus int
		wantCookie bool
	}{
		{
			name: "success: 200 + cookie",
			body: `{"login":"alice","password":"secret123"}`,
			setupMock: func(m *mockUserService) {
				m.On("Login", mock.Anything, "alice", "secret123").
					Return("token_abc", nil)
			},
			wantStatus: http.StatusOK,
			wantCookie: true,
		},
		{
			name:       "invalid json: 400",
			body:       `not-json`,
			setupMock:  func(m *mockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty login: 400",
			body: `{"login":"","password":"secret123"}`,
			setupMock: func(m *mockUserService) {
				m.On("Login", mock.Anything, "", "secret123").
					Return("", service.ErrEmptyLogin)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "wrong credentials: 401",
			body: `{"login":"alice","password":"wrongpass"}`,
			setupMock: func(m *mockUserService) {
				m.On("Login", mock.Anything, "alice", "wrongpass").
					Return("", service.ErrInvalidCredentials)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "internal error: 500",
			body: `{"login":"alice","password":"secret123"}`,
			setupMock: func(m *mockUserService) {
				m.On("Login", mock.Anything, "alice", "secret123").
					Return("", errors.New("db is down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userSvc := new(mockUserService)
			tt.setupMock(userSvc)

			h := newTestHandler(userSvc, new(mockOrderService), new(mockBalanceService))

			req := httptest.NewRequest(http.MethodPost, "/api/user/login",
				bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantCookie {
				resp := rr.Result()
				defer resp.Body.Close()
				cookies := resp.Cookies()
				found := false
				for _, c := range cookies {
					if c.Name == "token" {
						found = true
						assert.NotEmpty(t, c.Value)
					}
				}
				assert.True(t, found, "cookie 'token' not found")
			}

			userSvc.AssertExpectations(t)
		})
	}
}

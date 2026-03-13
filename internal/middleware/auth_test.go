package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/F3dosik/Hofermart/internal/ctxkey"
	"github.com/F3dosik/Hofermart/internal/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key"

func TestRequireAuth(t *testing.T) {
	userID := uuid.New()

	validToken, _ := jwt.GenerateToken(userID, testSecret)

	tests := []struct {
		name       string
		setupReq   func(*http.Request)
		wantStatus int
		checkCtx   func(*testing.T, *http.Request)
	}{
		{
			name: "valid token: passes through with userID in context",
			setupReq: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: "token", Value: validToken})
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing cookie: 401",
			setupReq:   func(r *http.Request) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid token: 401",
			setupReq: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: "token", Value: "invalid.token.value"})
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "token with wrong secret: 401",
			setupReq: func(r *http.Request) {
				wrongToken, _ := jwt.GenerateToken(userID, "wrong-secret")
				r.AddCookie(&http.Cookie{Name: "token", Value: wrongToken})
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedCtx context.Context
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedCtx = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			handler := RequireAuth(newLogger(), testSecret)(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			tt.setupReq(req)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantStatus == http.StatusOK {
				require.NotNil(t, capturedCtx)
				gotID, ok := capturedCtx.Value(ctxkey.UserIDKey).(uuid.UUID)
				assert.True(t, ok, "userID not found in context")
				assert.Equal(t, userID, gotID)
			}
		})
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func newLogger() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	return logger.Sugar()
}

func TestRequireJSON(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantStatus  int
	}{
		{"valid: application/json", "application/json", http.StatusOK},
		{"valid: application/json with charset", "application/json; charset=utf-8", http.StatusOK},
		{"invalid: text/plain", "text/plain", http.StatusUnsupportedMediaType},
		{"invalid: empty", "", http.StatusUnsupportedMediaType},
		{"invalid: multipart", "multipart/form-data", http.StatusUnsupportedMediaType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequireJSON(newLogger())(okHandler())

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

func TestRequirePlainText(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantStatus  int
	}{
		{"valid: text/plain", "text/plain", http.StatusOK},
		{"valid: text/plain with charset", "text/plain; charset=utf-8", http.StatusOK},
		{"invalid: application/json", "application/json", http.StatusUnsupportedMediaType},
		{"invalid: empty", "", http.StatusUnsupportedMediaType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequirePlainText(newLogger())(okHandler())

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

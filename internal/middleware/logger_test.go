package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithLogging(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.Handler
		wantStatus int
		wantSize   bool
	}{
		{
			name: "logs 200 response",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			wantStatus: http.StatusOK,
		},
		{
			name: "logs 404 response with body",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "not found", http.StatusNotFound)
			}),
			wantStatus: http.StatusNotFound,
			wantSize:   true,
		},
		{
			name: "default status 200 when WriteHeader not called",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("hello"))
			}),
			wantStatus: http.StatusOK,
			wantSize:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := WithLogging(newLogger())(tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantSize {
				assert.Greater(t, rr.Body.Len(), 0)
			}
		})
	}
}

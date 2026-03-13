package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestAccrualClient_GetAccrual(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		handler      http.HandlerFunc
		wantErr      error
		wantStatus   model.AccrualStatus
		checkErrType func(*testing.T, error)
	}{
		{
			name: "success: 200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"order":"12345678903","status":"PROCESSED","accrual":500}`))
			},
			wantStatus: model.AccrualStatusProcessed,
		},
		{
			name: "order not found: 204",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			wantErr: ErrOrderNotFound,
		},
		{
			name: "rate limit: 429 with Retry-After seconds",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
			},
			checkErrType: func(t *testing.T, err error) {
				var rateLimitErr *ErrRateLimit
				require.ErrorAs(t, err, &rateLimitErr)
				assert.Equal(t, 60*time.Second, rateLimitErr.RetryAfter)
			},
		},
		{
			name: "rate limit: 429 with Retry-After http date",
			handler: func(w http.ResponseWriter, r *http.Request) {
				future := time.Now().Add(2 * time.Minute)
				w.Header().Set("Retry-After", future.UTC().Format(http.TimeFormat))
				w.WriteHeader(http.StatusTooManyRequests)
			},
			checkErrType: func(t *testing.T, err error) {
				var rateLimitErr *ErrRateLimit
				require.ErrorAs(t, err, &rateLimitErr)
				assert.Greater(t, rateLimitErr.RetryAfter, time.Duration(0))
			},
		},
		{
			name: "rate limit: 429 without Retry-After header",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
			},
			checkErrType: func(t *testing.T, err error) {
				assert.Error(t, err)
				var rateLimitErr *ErrRateLimit
				assert.False(t, errors.As(err, &rateLimitErr))
			},
		},
		{
			name: "unexpected status: 500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			checkErrType: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTestServer(tt.handler)
			defer srv.Close()

			client := NewAccrual(srv.URL)
			result, err := client.GetAccrual(ctx, "12345678903")

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
				return
			}

			if tt.checkErrType != nil {
				tt.checkErrType(t, err)
				return
			}

			assert.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantStatus, result.Status)
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name        string
		headerValue string
		wantErr     bool
		wantMin     time.Duration
	}{
		{
			name:        "seconds format",
			headerValue: "30",
			wantMin:     30 * time.Second,
		},
		{
			name:        "http date format",
			headerValue: time.Now().Add(time.Minute).UTC().Format(http.TimeFormat),
			wantMin:     50 * time.Second,
		},
		{
			name:        "empty header",
			headerValue: "",
			wantErr:     true,
		},
		{
			name:        "invalid format",
			headerValue: "not-a-date-or-number",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.headerValue != "" {
					w.Header().Set("Retry-After", tt.headerValue)
				}
				w.WriteHeader(http.StatusTooManyRequests)
			}))
			defer srv.Close()

			c := resty.New()
			resp, err := c.R().Get(srv.URL)
			require.NoError(t, err)

			duration, err := parseRetryAfter(resp)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Zero(t, duration)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, duration, tt.wantMin)
			}
		})
	}
}

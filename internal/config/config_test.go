package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResolveFunctions(t *testing.T) {
	tests := []struct {
		name    string
		envStr  string
		flagStr string
		envInt  int
		flagInt int
		envDur  time.Duration
		flagDur time.Duration
	}{
		{
			name:    "env has priority",
			envStr:  "env",
			flagStr: "flag",
			envInt:  5,
			flagInt: 2,
			envDur:  time.Second,
			flagDur: time.Minute,
		},
		{
			name:    "fallback to flag",
			flagStr: "flag",
			flagInt: 3,
			flagDur: time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, firstNonEmpty(tt.envStr, tt.flagStr), resolveString(tt.envStr, tt.flagStr))

			expectedInt := tt.flagInt
			if tt.envInt != 0 {
				expectedInt = tt.envInt
			}
			assert.Equal(t, expectedInt, resolveInt(tt.envInt, tt.flagInt))

			expectedDur := tt.flagDur
			if tt.envDur != 0 {
				expectedDur = tt.envDur
			}
			assert.Equal(t, expectedDur, resolveDuration(tt.envDur, tt.flagDur))
		})
	}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name  string
		addr  string
		valid bool
	}{
		{"valid host:port", "localhost:8080", true},
		{"valid url", "http://localhost:8080", true},
		{"invalid address", "localhost", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, validateAddress(tt.addr))
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				ServiceAddress: "localhost:8081",
				DatabaseURI:    "postgres://test",
				AccrualAddress: "localhost:8080",
				LogLevel:       "development",
				JWTSecret:      "secret",
				WorkerCount:    3,
			},
		},
		{
			name: "invalid service address",
			config: Config{
				ServiceAddress: "bad",
				DatabaseURI:    "postgres://test",
				AccrualAddress: "localhost:8080",
				LogLevel:       "development",
				JWTSecret:      "secret",
			},
			wantErr: true,
		},
		{
			name: "empty database",
			config: Config{
				ServiceAddress: "localhost:8081",
				AccrualAddress: "localhost:8080",
				LogLevel:       "development",
				JWTSecret:      "secret",
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: Config{
				ServiceAddress: "localhost:8081",
				DatabaseURI:    "postgres://test",
				AccrualAddress: "localhost:8080",
				LogLevel:       "invalid",
				JWTSecret:      "secret",
			},
			wantErr: true,
		},
		{
			name: "negative workers",
			config: Config{
				ServiceAddress: "localhost:8081",
				DatabaseURI:    "postgres://test",
				AccrualAddress: "localhost:8080",
				LogLevel:       "development",
				JWTSecret:      "secret",
				WorkerCount:    -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

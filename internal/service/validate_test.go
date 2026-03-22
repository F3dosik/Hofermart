package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name    string
		login   string
		wantErr error
	}{
		{"valid login", "alice", nil},
		{"empty login", "", ErrEmptyLogin},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, validateLogin(tt.login), tt.wantErr)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"valid password", "secret123", nil},
		{"too short", "abc", ErrPasswordTooShort},
		{"exactly 8 chars", "12345678", nil},
		{"7 chars", "1234567", ErrPasswordTooShort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, validatePassword(tt.password), tt.wantErr)
		})
	}
}

func TestValidateOrder(t *testing.T) {
	tests := []struct {
		name        string
		orderNumber string
		wantErr     error
	}{
		{"valid order", "12345678903", nil},
		{"invalid order", "12345678900", ErrInvalidOrderNumber},
		{"empty string", "", ErrInvalidOrderNumber},
		{"non-digit chars", "1234abcdefg", ErrInvalidOrderNumber},
		{"single zero", "0", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, validateOrder(tt.orderNumber), tt.wantErr)
		})
	}
}

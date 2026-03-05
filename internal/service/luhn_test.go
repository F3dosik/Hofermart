package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckLuhn(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"valid number", "12345678903", true},
		{"invalid number", "12345678900", false},
		{"empty string", "", false},
		{"non-digit chars", "1234abcdefg", false},
		{"single zero", "0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkLuhn(tt.number)
			assert.Equal(t, tt.want, got)
		})
	}
}

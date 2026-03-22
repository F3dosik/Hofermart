package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key"

func TestGenerateToken(t *testing.T) {
	userID := uuid.New()

	token, err := GenerateToken(userID, testSecret)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseToken(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		makeToken func() string
		wantErr   bool
		wantID    uuid.UUID
	}{
		{
			name: "valid token",
			makeToken: func() string {
				token, _ := GenerateToken(userID, testSecret)
				return token
			},
			wantID: userID,
		},
		{
			name: "wrong secret",
			makeToken: func() string {
				token, _ := GenerateToken(userID, "other-secret")
				return token
			},
			wantErr: true,
		},
		{
			name:      "malformed token",
			makeToken: func() string { return "not.a.token" },
			wantErr:   true,
		},
		{
			name:      "empty token",
			makeToken: func() string { return "" },
			wantErr:   true,
		},
		{
			name: "expired token",
			makeToken: func() string {
				claims := Claims{
					UserID: userID,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				signed, _ := token.SignedString([]byte(testSecret))
				return signed
			},
			wantErr: true,
		},
		{
			name: "wrong signing method",
			makeToken: func() string {
				claims := Claims{UserID: userID}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				return token.Raw
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseToken(tt.makeToken(), testSecret)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)
				assert.Equal(t, tt.wantID, claims.UserID)
			}
		})
	}
}

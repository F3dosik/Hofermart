package middleware

import (
	"context"
	"net/http"

	"github.com/F3dosik/Hofermart/internal/ctxkey"
	"github.com/F3dosik/Hofermart/internal/jwt"
	"go.uber.org/zap"
)

func RequireAuth(logger *zap.SugaredLogger, secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil {
				logger.Errorw("auth: missing token cookie", "error", err)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			tokenString := cookie.Value
			claims, err := jwt.ParseToken(tokenString, secretKey)
			if err != nil {
				logger.Debugw("auth error", "error", err)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxkey.UserIDKey, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

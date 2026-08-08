package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/TheAmgadX/moltaqa-backend/shared/env"
	"github.com/golang-jwt/jwt/v5"
)

const (
	UserIDKey = "user_id"
)

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized access", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			secret := env.GetString("JWT_SIGNING_KEY", "auth-secret")
			return []byte(secret), nil
		})

		if err != nil {
			http.Error(w, "unauthorized access", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

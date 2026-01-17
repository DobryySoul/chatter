package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userKey contextKey = "authUser"

type TokenParser interface {
	Parse(token string) (string, error)
}

func UserFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(userKey).(string)
	return value, ok
}

func RequireAuth(parser TokenParser, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization", http.StatusUnauthorized)
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid authorization", http.StatusUnauthorized)
			return
		}

		username, err := parser.Parse(parts[1])
		if err != nil || username == "" {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

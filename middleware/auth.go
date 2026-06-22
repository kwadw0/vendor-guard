package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	jwtutil "vendor-guard/auth/jwt"
	"vendor-guard/utils"
)

var errUnauthorized = errors.New("unauthorized")

type contextKey string

const UserIDKey contextKey = "user_id"

// GetUserID retrieves the authenticated user's ID (as a string) from the request context.
// Returns an empty string if the key is missing.
func GetUserID(r *http.Request) string {
	val, _ := r.Context().Value(UserIDKey).(string)
	return val
}

// RequireAuth validates the Bearer token and injects the user_id into the request context.
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				utils.ErrorJSON(w, http.StatusUnauthorized, errUnauthorized, "UNAUTHORIZED")
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := jwtutil.ValidateToken(token, jwtSecret)
			if err != nil {
				utils.ErrorJSON(w, http.StatusUnauthorized, errUnauthorized, "UNAUTHORIZED")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

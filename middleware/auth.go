package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	jwtutil "preuvio/auth/jwt"
	"preuvio/utils"
)

var errUnauthorized = errors.New("unauthorized")

type contextKey string

const UserIDKey contextKey = "user_id"
const RoleIDKey contextKey = "role_id"

// GetUserID retrieves the authenticated user's ID (as a string) from the request context.
func GetUserID(r *http.Request) string {
	val, _ := r.Context().Value(UserIDKey).(string)
	return val
}

// GetRoleID retrieves the authenticated user's role ID (as a string) from the request context.
func GetRoleID(r *http.Request) string {
	val, _ := r.Context().Value(RoleIDKey).(string)
	return val
}

// RequireAuth validates the Bearer token and injects user_id and role_id into the request context.
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			token := authHeader
			// Accept "Bearer <token>" case-insensitively, or a bare token
			// (Swagger Authorize modal often pastes the raw JWT).
			if parts := strings.SplitN(authHeader, " ", 2); len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				token = strings.TrimSpace(parts[1])
			}
			if token == "" {
				utils.ErrorJSON(w, http.StatusUnauthorized, errUnauthorized, "UNAUTHORIZED")
				return
			}

			claims, err := jwtutil.ValidateToken(token, jwtSecret)
			// Empty UserID means a refresh token (or foreign JWT) was sent
			// as the access token — reject explicitly instead of injecting "".
			if err != nil || claims.UserID == "" {
				utils.ErrorJSON(w, http.StatusUnauthorized, errUnauthorized, "UNAUTHORIZED")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleIDKey, claims.RoleID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

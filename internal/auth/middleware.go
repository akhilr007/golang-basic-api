package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/akhilr007/tasks/internal/utils"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		
		if authHeader == "" {
			utils.WriteError(w, http.StatusUnauthorized, "missing token")
			return
		}
		
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.WriteError(w, http.StatusUnauthorized, "invalid token format")
			return
		}
		
		claims, err := ValidateToken(parts[1])
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
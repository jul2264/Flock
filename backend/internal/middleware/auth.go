package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
)

type contextKey string

// UserIDKey is the context key for the authenticated Clerk user ID (subject).
const UserIDKey contextKey = "userID"

// ClerkMiddleware verifies the Bearer JWT issued by Clerk and stores the
// user's Clerk ID in the request context under UserIDKey.
// Must run before RequireAuth.
func ClerkMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" || token == authHeader {
				// Header missing or not in "Bearer <token>" format
				http.Error(w, `{"error":"missing or malformed Authorization header"}`, http.StatusUnauthorized)
				return
			}

			claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{
				Token: token,
			})
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Store the Clerk user ID (subject) for downstream handlers
			ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth is a guard middleware that rejects requests where ClerkMiddleware
// has not successfully stored a user ID. Apply it after ClerkMiddleware.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clerkID, ok := r.Context().Value(UserIDKey).(string)
		if !ok || clerkID == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
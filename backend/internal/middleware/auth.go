package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
)

type contextKey string

// UserIDKey is the context key for the authenticated Clerk user ID (subject).
const UserIDKey contextKey = "userID"

// RoleKey is the context key for the authenticated user's role ("user", "organizer", "admin").
// Only present after RequireRole middleware has run successfully.
const RoleKey contextKey = "role"

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

// RequireRole returns middleware that enforces that the authenticated user's role
// is one of the provided allowed roles. It performs a single indexed DB lookup
// on clerk_id and stores the resolved role in context under RoleKey so that
// downstream handlers can read it without re-querying.
//
// Must run after ClerkMiddleware and RequireAuth.
//
// Usage:
//
//	r.Use(middleware.RequireRole(db, "organizer", "admin"))  // organizer or admin
//	r.Use(middleware.RequireRole(db, "admin"))               // admin only
func RequireRole(db *sql.DB, roles ...string) func(http.Handler) http.Handler {
	// Build a set for O(1) lookups
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clerkID, ok := r.Context().Value(UserIDKey).(string)
			if !ok || clerkID == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			var role string
			err := db.QueryRowContext(r.Context(),
				`SELECT role FROM users WHERE clerk_id = $1`, clerkID,
			).Scan(&role)
			if err != nil {
				if err == sql.ErrNoRows {
					// User has a valid Clerk token but no synced profile yet
					http.Error(w, `{"error":"user profile not found — call POST /users/sync first"}`, http.StatusForbidden)
					return
				}
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				return
			}

			if !allowed[role] {
				http.Error(w, `{"error":"forbidden: insufficient permissions"}`, http.StatusForbidden)
				return
			}

			// Store role in context so downstream handlers can read it without re-querying
			ctx := context.WithValue(r.Context(), RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RoleFromCtx is a helper to read the resolved role from the request context.
// Returns an empty string if RequireRole middleware has not run.
func RoleFromCtx(r *http.Request) string {
	role, _ := r.Context().Value(RoleKey).(string)
	return role
}
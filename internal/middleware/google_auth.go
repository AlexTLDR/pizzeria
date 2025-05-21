package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/AlexTLDR/pizzeria/internal/auth"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Context keys
const (
	userEmailKey contextKey = "user_email"
)

// Initialize OAuth config
var oauthConfig *auth.OAuthConfig

// InitializeOAuth sets up the OAuth configuration
func InitializeOAuth() error {
	var err error

	oauthConfig, err = auth.Initialize()
	if err != nil {
		return err
	}

	// Initialize cookie secret for secure sessions
	InitializeCookieSecret()

	return nil
}

// GetOAuthConfig returns the OAuth configuration
func GetOAuthConfig() *auth.OAuthConfig {
	return oauthConfig
}

// GoogleAuth is middleware that checks if the user is authenticated using Google OAuth
func GoogleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the secure session cookie
		userEmail, valid := VerifySecureSessionCookie(r)
		if !valid {
			log.Println("Invalid or expired session, redirecting to login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)

			return
		}

		// Check if the user email is in the allowed list
		if !oauthConfig.IsAllowedEmail(userEmail) {
			log.Printf("Unauthorized access attempt by: %s", userEmail)
			ClearSecureSessionCookie(w) // Clear invalid session
			http.Error(w, "Unauthorized: You do not have permission to access this page", http.StatusForbidden)

			return
		}

		// Set the email in the request context
		ctx := context.WithValue(r.Context(), userEmailKey, userEmail)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserEmail extracts the user's email from the request context
func GetUserEmail(r *http.Request) string {
	if email, ok := r.Context().Value(userEmailKey).(string); ok {
		return email
	}

	return ""
}

// SetSessionCookie sets a Google session cookie with the user's email
func SetSessionCookie(w http.ResponseWriter, email string) {
	// Use the secure cookie implementation
	SetSecureSessionCookie(w, email)
}

// ClearSessionCookie clears the Google session cookie
func ClearSessionCookie(w http.ResponseWriter) {
	// Use the secure cookie implementation
	ClearSecureSessionCookie(w)
}

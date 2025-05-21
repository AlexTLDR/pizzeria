package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/AlexTLDR/pizzeria/internal/middleware"
)

// ShowGoogleLogin renders the Google login page
func (m *Repository) ShowGoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate a random state parameter for CSRF protection
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, "Could not generate state parameter", http.StatusInternalServerError)
		return
	}

	state := base64.URLEncoding.EncodeToString(b)

	// Store state in a cookie for verification later
	stateCookie := http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300, // 5 minutes
		Expires:  time.Now().Add(5 * time.Minute),
	}
	http.SetCookie(w, &stateCookie)

	// Get login URL from the OAuth config
	loginURL := m.OAuthConfig.GetGoogleLoginURL(state)

	// Get any error message from the URL
	errorMsg := r.URL.Query().Get("error")

	// Render login template
	err = m.TemplateCache["google-login.html"].Execute(w, map[string]interface{}{
		"Title":    "Admin Login",
		"LoginURL": loginURL,
		"Error":    errorMsg,
		"Year":     time.Now().Year(),
	})

	if err != nil {
		log.Printf("Error rendering Google login template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GoogleCallback handles the OAuth callback from Google
func (m *Repository) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state parameter from cookie
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		http.Redirect(w, r, "/login?error=Invalid state parameter", http.StatusSeeOther)
		return
	}

	expectedState := stateCookie.Value
	actualState := r.URL.Query().Get("state")

	if expectedState != actualState {
		log.Printf("Invalid state parameter: expected %s, got %s", expectedState, actualState)
		http.Redirect(w, r, "/login?error=Invalid state parameter", http.StatusSeeOther)

		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	// Get the authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/login?error=Missing authorization code", http.StatusSeeOther)
		return
	}

	// Exchange the code for a token
	token, err := m.OAuthConfig.ExchangeCodeForToken(code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Redirect(w, r, "/login?error=Failed to authenticate", http.StatusSeeOther)

		return
	}

	// Get user info
	userInfo, err := m.OAuthConfig.GetUserInfo(token)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Redirect(w, r, "/login?error=Failed to get user information", http.StatusSeeOther)

		return
	}

	// Check if the email is in the allowed list
	if !m.OAuthConfig.IsAllowedEmail(userInfo.Email) {
		log.Printf("Unauthorized access attempt by: %s", userInfo.Email)
		http.Redirect(w, r, "/login?error=You are not authorized to access the admin area", http.StatusSeeOther)

		return
	}

	// Set session cookie with the authenticated email
	middleware.SetSessionCookie(w, userInfo.Email)

	// Redirect to admin dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// GoogleLogout handles user logout
func (m *Repository) GoogleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	middleware.ClearSessionCookie(w)

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

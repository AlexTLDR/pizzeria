package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlexTLDR/pizzeria/internal/middleware"
)

// ShowLoginPage displays the login page
func (m *Repository) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	// Check if there's an error message
	errorMsg := r.URL.Query().Get("error")

	// Render the login page
	err := m.TemplateCache["login.html"].Execute(w, map[string]interface{}{
		"Title": "Admin Login",
		"Error": errorMsg,
		"Year":  time.Now().Year(),
	})

	if err != nil {
		// Just log the error since template.Execute likely already wrote to the response
		log.Printf("ERROR: Template rendering failed in ShowLoginPage: %v", err)
		return
	}
}

// HandleGoogleLogin initiates the Google OAuth login flow
func (m *Repository) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate a random state token to protect against CSRF
	state, err := generateStateToken()
	if err != nil {
		m.serverError(w, err, "HandleGoogleLogin - generating state token")
		return
	}

	// Store the state token in a cookie
	stateCookie := http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 5, // 5 minutes
	}
	http.SetCookie(w, &stateCookie)

	// Redirect to Google's OAuth consent page
	url := m.OAuthConfig.GetGoogleLoginURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback processes the callback from Google OAuth
func (m *Repository) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Get the state from the URL
	state := r.URL.Query().Get("state")
	if state == "" {
		log.Println("No state token in callback")
		http.Redirect(w, r, "/login?error=Invalid+authentication+attempt", http.StatusSeeOther)
		return
	}

	// Get the state from the cookie
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		log.Printf("State cookie not found: %v", err)
		http.Redirect(w, r, "/login?error=Invalid+authentication+attempt", http.StatusSeeOther)
		return
	}

	// Verify the state matches
	if state != stateCookie.Value {
		log.Printf("State mismatch: %s vs %s", state, stateCookie.Value)
		http.Redirect(w, r, "/login?error=Invalid+authentication+attempt", http.StatusSeeOther)
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

	// Get the authorization code from the URL
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println("No code in callback")
		http.Redirect(w, r, "/login?error=Authentication+failed", http.StatusSeeOther)
		return
	}

	// Exchange the authorization code for a token
	token, err := m.OAuthConfig.ExchangeCodeForToken(code)
	if err != nil {
		m.serverError(w, err, "HandleGoogleCallback - exchanging OAuth code")
		return
	}

	// Get the user info from Google
	googleUserInfo, err := m.OAuthConfig.GetUserInfo(token)
	if err != nil {
		m.serverError(w, err, "HandleGoogleCallback - getting user info")
		return
	}

	// Check if the email is verified
	if !googleUserInfo.VerifiedEmail {
		log.Printf("Email not verified: %s", googleUserInfo.Email)
		http.Redirect(w, r, "/login?error=Email+not+verified", http.StatusSeeOther)
		return
	}

	// Check if the email is in the allowed list
	if !m.OAuthConfig.IsAllowedEmail(googleUserInfo.Email) {
		log.Printf("Unauthorized email: %s", googleUserInfo.Email)
		http.Redirect(w, r, "/login?error=Unauthorized+email", http.StatusSeeOther)
		return
	}

	// Set the session cookie with the email
	middleware.SetSessionCookie(w, googleUserInfo.Email)

	// Redirect to the admin dashboard
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// HandleLogout logs the user out
func (m *Repository) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	middleware.ClearSessionCookie(w)

	// Redirect to the login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// generateStateToken generates a random state token for CSRF protection
func generateStateToken() (string, error) {
	// Generate 32 bytes of random data
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64
	return base64.URLEncoding.EncodeToString(b), nil
}

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthConfig holds the configuration for Google OAuth
type OAuthConfig struct {
	GoogleOAuthConfig *oauth2.Config
	AllowedEmails     []string
}

// GoogleUserInfo holds the user information from Google
type GoogleUserInfo struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
	Name          string `json:"name"`
}

// Initialize loads OAuth configuration from environment
func Initialize() (*OAuthConfig, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	// (environment variables might be set via docker-compose or system)
	if err := godotenv.Load(); err != nil {
		// This is normal in Docker environments where env vars are set directly
		// Just log a debug message but don't fail
		fmt.Printf("Info: .env file not found (using environment variables): %v\n", err)
	}

	// Get Google OAuth credentials
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil, errors.New("missing Google OAuth configuration - ensure GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and GOOGLE_REDIRECT_URL environment variables are set")
	}

	// Get allowed emails
	allowedEmailsStr := os.Getenv("ALLOWED_EMAILS")

	allowedEmails := strings.Split(allowedEmailsStr, ",")
	for i := range allowedEmails {
		allowedEmails[i] = strings.TrimSpace(allowedEmails[i])
	}

	if len(allowedEmails) == 0 || (len(allowedEmails) == 1 && allowedEmails[0] == "") {
		return nil, errors.New("no allowed email addresses configured - ensure ALLOWED_EMAILS environment variable is set")
	}

	// Create OAuth config
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	return &OAuthConfig{
		GoogleOAuthConfig: oauthConfig,
		AllowedEmails:     allowedEmails,
	}, nil
}

// GetGoogleLoginURL returns the URL for the Google login
func (c *OAuthConfig) GetGoogleLoginURL(state string) string {
	return c.GoogleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GetUserInfo gets the user information from Google using an OAuth token
func (c *OAuthConfig) GetUserInfo(token *oauth2.Token) (*GoogleUserInfo, error) {
	// Create HTTP client with token
	client := c.GoogleOAuthConfig.Client(context.Background(), token)

	// Call Google API to get user info
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}

	// Parse user info
	var userInfo GoogleUserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, fmt.Errorf("failed parsing user info: %w", err)
	}

	return &userInfo, nil
}

// IsAllowedEmail checks if the email is in the allowed list
func (c *OAuthConfig) IsAllowedEmail(email string) bool {
	return slices.Contains(c.AllowedEmails, email)
}

// ExchangeCodeForToken exchanges an authorization code for an OAuth token
func (c *OAuthConfig) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	token, err := c.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}

	return token, nil
}

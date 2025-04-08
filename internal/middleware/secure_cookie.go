package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// SessionCookieName is the name of the cookie used to store the session
	SessionCookieName = "google_session"
	// SessionDuration is how long the session cookie will be valid
	SessionDuration = 24 * time.Hour
)

var (
	// cookieSecret is used to sign the session cookie
	cookieSecret []byte
)

// InitializeCookieSecret generates or loads a secret key for cookie signing
func InitializeCookieSecret() {
	// For a real application, you should use a stable secret from environment or config
	// In this example, we'll generate a random secret if one doesn't exist
	secret := make([]byte, 32) // 256 bits
	_, err := rand.Read(secret)
	if err != nil {
		log.Printf("Warning: Failed to generate secure cookie secret: %v", err)
		// Fallback to a fixed secret (not recommended for production)
		secret = []byte("default-cookie-secret-please-change-me")
	}
	cookieSecret = secret
	log.Println("Cookie signing secret initialized")
}

// SetSecureSessionCookie creates and sets a signed session cookie with the user's email
func SetSecureSessionCookie(w http.ResponseWriter, email string) {
	// Current timestamp for the cookie
	now := time.Now()
	expires := now.Add(SessionDuration)

	// Create the cookie payload: email|expiration_timestamp
	expiresStr := strconv.FormatInt(expires.Unix(), 10)
	payload := fmt.Sprintf("%s|%s", email, expiresStr)

	// Create HMAC signature
	h := hmac.New(sha256.New, cookieSecret)
	h.Write([]byte(payload))
	signature := h.Sum(nil)

	// Encode the payload and signature for the cookie
	encodedPayload := base64.URLEncoding.EncodeToString([]byte(payload))
	encodedSignature := base64.URLEncoding.EncodeToString(signature)

	// Final cookie value: base64(payload).base64(signature)
	cookieValue := fmt.Sprintf("%s.%s", encodedPayload, encodedSignature)

	// Set the cookie
	cookie := http.Cookie{
		Name:     SessionCookieName,
		Value:    cookieValue,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(SessionDuration.Seconds()),
		Expires:  expires,
	}

	http.SetCookie(w, &cookie)
	log.Printf("Set secure session cookie for %s, expires: %s", email, expires.Format(time.RFC3339))
}

// VerifySecureSessionCookie verifies the signature of a session cookie and returns the email
func VerifySecureSessionCookie(r *http.Request) (string, bool) {
	// Get the cookie
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		log.Printf("No session cookie found: %v", err)
		return "", false
	}

	// Split the cookie value into payload and signature
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		log.Printf("Invalid cookie format")
		return "", false
	}

	// Decode the payload and signature
	payloadBytes, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		log.Printf("Invalid payload encoding: %v", err)
		return "", false
	}

	signatureBytes, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		log.Printf("Invalid signature encoding: %v", err)
		return "", false
	}

	// Verify the signature
	h := hmac.New(sha256.New, cookieSecret)
	h.Write(payloadBytes)
	expectedSignature := h.Sum(nil)

	if !hmac.Equal(signatureBytes, expectedSignature) {
		log.Printf("Cookie signature verification failed")
		return "", false
	}

	// Parse the payload
	payload := string(payloadBytes)
	payloadParts := strings.Split(payload, "|")
	if len(payloadParts) != 2 {
		log.Printf("Invalid payload format")
		return "", false
	}

	email := payloadParts[0]
	expirationStr := payloadParts[1]

	// Verify expiration
	expirationUnix, err := strconv.ParseInt(expirationStr, 10, 64)
	if err != nil {
		log.Printf("Invalid expiration timestamp: %v", err)
		return "", false
	}

	expiration := time.Unix(expirationUnix, 0)
	if time.Now().After(expiration) {
		log.Printf("Session expired at %s", expiration.Format(time.RFC3339))
		return "", false
	}

	return email, true
}

// ClearSecureSessionCookie removes the session cookie
func ClearSecureSessionCookie(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
	}

	http.SetCookie(w, &cookie)
	log.Println("Cleared session cookie")
}

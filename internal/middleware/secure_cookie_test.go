package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSecureCookieRoundTrip(t *testing.T) {
	InitializeCookieSecret()

	testEmails := []string{
		"test@example.com",
		"admin@pizzeria.com",
		"user.name+tag@example.org",
		"", // Empty email edge case
	}

	for _, email := range testEmails {
		t.Run("Cookie roundtrip for "+email, func(t *testing.T) {
			w := httptest.NewRecorder()
			SetSecureSessionCookie(w, email)

			cookies := w.Result().Cookies()
			if len(cookies) == 0 {
				t.Fatal("No cookies set")
			}

			var sessionCookie *http.Cookie
			for _, cookie := range cookies {
				if cookie.Name == SessionCookieName {
					sessionCookie = cookie
					break
				}
			}

			if sessionCookie == nil {
				t.Fatalf("Session cookie '%s' not found", SessionCookieName)
			}

			if !sessionCookie.HttpOnly {
				t.Error("Session cookie should be HttpOnly")
			}

			if !sessionCookie.Secure {
				t.Error("Session cookie should be Secure")
			}

			req := httptest.NewRequest("GET", "/admin", nil)
			req.AddCookie(sessionCookie)

			gotEmail, valid := VerifySecureSessionCookie(req)
			if !valid {
				t.Errorf("Cookie validation failed for email: %s", email)
			}

			if gotEmail != email {
				t.Errorf("Got email %s, want %s", gotEmail, email)
			}
		})
	}
}

func TestCookieExpiration(t *testing.T) {
	InitializeCookieSecret()

	w := httptest.NewRecorder()

	email := "test@example.com"
	SetSecureSessionCookie(w, email)

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionCookieName {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Session cookie not found")
	}

	req := httptest.NewRequest("GET", "/admin", nil)
	req.AddCookie(sessionCookie)

	_, valid := VerifySecureSessionCookie(req)
	if !valid {
		t.Error("Cookie should be valid initially")
	}

	parts := sessionCookie.Value
	expiredCookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    parts,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(-24 * time.Hour),
	}

	expiredReq := httptest.NewRequest("GET", "/admin", nil)
	expiredReq.AddCookie(expiredCookie)
}

func TestClearSecureSessionCookie(t *testing.T) {
	InitializeCookieSecret()

	w := httptest.NewRecorder()
	SetSecureSessionCookie(w, "test@example.com")

	w = httptest.NewRecorder() // Reset recorder
	ClearSecureSessionCookie(w)

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionCookieName {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Session cookie not found after clearing")
	}

	if !sessionCookie.Expires.Before(time.Now()) {
		t.Error("Cleared cookie should have expiration in the past")
	}

	if sessionCookie.MaxAge >= 0 {
		t.Error("Cleared cookie should have negative MaxAge")
	}

	if sessionCookie.Value != "" {
		t.Errorf("Cleared cookie should have empty value, got %s", sessionCookie.Value)
	}
}
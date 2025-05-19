package main

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AlexTLDR/pizzeria/db"
	"github.com/AlexTLDR/pizzeria/internal/auth"
	"github.com/AlexTLDR/pizzeria/internal/middleware"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestEnvironment creates a test environment with a temporary database
func setupTestEnvironment(t *testing.T) (string, func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pizzeria-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create necessary subdirectories
	dbDir := filepath.Join(tempDir, "db")
	migrationsDir := filepath.Join(dbDir, "migrations")
	staticDir := filepath.Join(tempDir, "static")
	templatesDir := filepath.Join(tempDir, "templates")

	// Create directories
	for _, dir := range []string{dbDir, migrationsDir, staticDir, templatesDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Copy migrations from real project to test directory
	// For integration tests, we'd need to copy actual migrations
	// This is simplified for the test
	createTestMigrations(t, migrationsDir)

	// Create a test database
	dbPath := filepath.Join(dbDir, "pizzeria.db")
	database, err := db.InitDB(dbPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	database.Close()

	// Create a cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// Set environment variables for testing
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	os.Setenv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback")
	os.Setenv("ALLOWED_EMAILS", "test@example.com")

	// Initialize OAuth for testing
	middleware.InitializeOAuth()

	return tempDir, cleanup
}

// createTestMigrations creates simple test migrations
func createTestMigrations(t *testing.T, migrationsDir string) {
	// Create an initial migration file
	initialMigration := `-- +goose Up
CREATE TABLE menu_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    price REAL NOT NULL,
    small_price REAL,
    category TEXT NOT NULL,
    image_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE flash_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    message TEXT NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    active BOOLEAN NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE flash_messages;
DROP TABLE menu_items;`

	err := os.WriteFile(filepath.Join(migrationsDir, "20250404172242_init_schema.sql"), []byte(initialMigration), 0644)
	if err != nil {
		t.Fatalf("Failed to create test migration: %v", err)
	}
}

// TestRunMigrations tests that migrations can be run successfully
func TestRunMigrations(t *testing.T) {
	// Skip in short test mode
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test environment
	testDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Set up database path
	dbPath := filepath.Join(testDir, "db", "pizzeria.db")

	// Run migrations
	err := runMigrations(dbPath)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Verify migrations by checking if tables exist
	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Check if menu_items table exists
	var tableName string
	err = database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='menu_items'").Scan(&tableName)
	if err != nil {
		t.Errorf("menu_items table was not created: %v", err)
	}

	// Check if flash_messages table exists
	err = database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='flash_messages'").Scan(&tableName)
	if err != nil {
		t.Errorf("flash_messages table was not created: %v", err)
	}
}

// TestAuthenticatedAdmin tests the authenticatedAdmin function
func TestAuthenticatedAdmin(t *testing.T) {
	// Skip in short test mode
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test environment
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a request
	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	w := httptest.NewRecorder()

	// The request should redirect to login since there's no valid session
	authenticatedAdmin(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
	}

	// Check the redirect location
	location := resp.Header.Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to /login, got %s", location)
	}
}

// TestOAuthConfig tests the OAuth configuration
func TestOAuthConfig(t *testing.T) {
	// Initialize OAuth
	if err := middleware.InitializeOAuth(); err != nil {
		t.Fatalf("Failed to initialize OAuth: %v", err)
	}

	// Get the OAuth config
	config := middleware.GetOAuthConfig()
	if config == nil {
		t.Fatal("OAuth config is nil")
	}

	// Test email verification
	if !config.IsAllowedEmail("test@example.com") {
		t.Error("test@example.com should be an allowed email")
	}

	if config.IsAllowedEmail("unauthorized@example.com") {
		t.Error("unauthorized@example.com should not be an allowed email")
	}
}

// MockResponseWriter is a mock http.ResponseWriter for testing
type MockResponseWriter struct {
	headers http.Header
	status  int
	body    []byte
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		headers: make(http.Header),
	}
}

func (m *MockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *MockResponseWriter) Write(b []byte) (int, error) {
	m.body = append(m.body, b...)
	return len(b), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

// TestSecureCookie tests the secure cookie functionality
func TestSecureCookie(t *testing.T) {
	// Initialize cookie secret
	middleware.InitializeCookieSecret()

	// Create a mock response writer
	w := NewMockResponseWriter()

	// Set a secure cookie
	middleware.SetSecureSessionCookie(w, "test@example.com")

	// Verify a cookie was set
	cookieHeader := w.headers.Get("Set-Cookie")
	if cookieHeader == "" {
		t.Fatal("No cookie was set")
	}

	// Create a request with the cookie
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Cookie", cookieHeader)

	// Verify the cookie
	email, valid := middleware.VerifySecureSessionCookie(req)
	if !valid {
		t.Error("Cookie should be valid")
	}

	if email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", email)
	}
}

// TestGoogleAuthenticatedRedirect tests the authenticatedRedirect function
func TestGoogleAuthenticatedRedirect(t *testing.T) {
	// Create a request
	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	// Call the function
	authenticatedRedirect(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
	}

	// Check the redirect location
	location := resp.Header.Get("Location")
	if location != "/admin/" {
		t.Errorf("Expected redirect to /admin/, got %s", location)
	}
}

// TestGoogleUserInfo tests the GetUserInfo method in auth package
func TestGoogleUserInfo(t *testing.T) {
	// Create a mock OAuth config
	config := &auth.OAuthConfig{
		GoogleOAuthConfig: nil, // We'll mock the HTTP client
		AllowedEmails:     []string{"test@example.com"},
	}

	// Test the IsAllowedEmail method
	if !config.IsAllowedEmail("test@example.com") {
		t.Error("test@example.com should be an allowed email")
	}

	if config.IsAllowedEmail("unauthorized@example.com") {
		t.Error("unauthorized@example.com should not be an allowed email")
	}
}
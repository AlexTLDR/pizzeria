package handlers

import (
	"database/sql"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexTLDR/pizzeria/internal/auth"
	"github.com/AlexTLDR/pizzeria/internal/models"
)

// NewTestRepo creates a repository struct for testing
func NewTestRepo(t *testing.T) *Repository {
	// Load templates for testing
	templateCache := make(map[string]*template.Template)
	
	// Create temp database for testing
	db, err := createTestDB()
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}

	// Create sample DB model
	testDBModel := &models.DBModel{
		DB: db,
	}

	// Create test OAuth config
	testOAuthConfig := &auth.OAuthConfig{
		AllowedEmails: []string{"test@example.com", "admin@example.com"},
	}

	return &Repository{
		DB:            testDBModel,
		TemplateCache: templateCache,
		OAuthConfig:   testOAuthConfig,
	}
}

// createTestDB creates a temporary SQLite database for testing
func createTestDB() (*sql.DB, error) {
	// Create a temp directory
	tempDir, err := os.MkdirTemp("", "pizzeria-test-*")
	if err != nil {
		return nil, err
	}

	// Create a temp database file
	dbPath := filepath.Join(tempDir, "test.db")
	
	// Open the database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create test schema
	_, err = db.Exec(`
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
	`)
	
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// CreateTestRequest creates a test request with optional body
func CreateTestRequest(t *testing.T, method, url string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	
	rr := httptest.NewRecorder()
	return req, rr
}

// CleanTestDB closes the database and cleans up any temporary files
func CleanTestDB(repo *Repository) {
	if repo != nil && repo.DB != nil && repo.DB.DB != nil {
		// Get the database file path from the connection
		var dbPath string
		err := repo.DB.DB.QueryRow("PRAGMA database_list").Scan(nil, &dbPath, nil)
		if err == nil && dbPath != "" {
			// Close the database connection
			repo.DB.DB.Close()
			
			// Remove the directory containing the database file
			if filepath.Dir(dbPath) != "." {
				os.RemoveAll(filepath.Dir(dbPath))
			}
		}
	}
}
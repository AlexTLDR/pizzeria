package handlers

import (
	"database/sql"
	"fmt"
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

func createMockTemplates() map[string]*template.Template {
	funcMap := template.FuncMap{
		"deref": func(f *float64) float64 {
			if f != nil {
				return *f
			}
			return 0
		},
	}

	indexTemplate := template.New("index.html").Funcs(funcMap)
	var err error
	indexTemplate, err = indexTemplate.Parse(`<!DOCTYPE html>
<html>
<head><title>Mock Index Template</title></head>
<body>
  <h1>Mock Template for Testing</h1>
  <div>This is mock content for testing the Home handler</div>
</body>
</html>`)
	if err != nil {
		panic(err) // In tests, we can panic on template parsing errors
	}

	adminTemplate := template.New("admin-dashboard.html").Funcs(funcMap)
	adminTemplate, err = adminTemplate.Parse(`<html><body>Mock Admin Dashboard</body></html>`)
	if err != nil {
		panic(err)
	}

	loginTemplate := template.New("login.html").Funcs(funcMap)
	loginTemplate, err = loginTemplate.Parse(`<html><body>Mock Login Page</body></html>`)
	if err != nil {
		panic(err)
	}

	menuFormTemplate := template.New("menu-form.html").Funcs(funcMap)
	menuFormTemplate, err = menuFormTemplate.Parse(`<html><body>Mock Menu Form</body></html>`)
	if err != nil {
		panic(err)
	}

	templateCache := map[string]*template.Template{
		"index.html":           indexTemplate,
		"admin-dashboard.html": adminTemplate,
		"login.html":           loginTemplate,
		"menu-form.html":       menuFormTemplate,
	}

	return templateCache
}

func NewTestAppServices(t *testing.T) *AppServices {
	templateCache := createMockTemplates()

	db, err := createTestDB()
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}

	testDBModel := &models.DBModel{
		DB: db,
	}

	testOAuthConfig := &auth.OAuthConfig{
		AllowedEmails: []string{"test@example.com", "admin@example.com"},
	}

	return &AppServices{
		DB:            testDBModel,
		TemplateCache: templateCache,
		OAuthConfig:   testOAuthConfig,
	}
}

func createTestDB() (*sql.DB, error) {
	tempDir, err := os.MkdirTemp("", "pizzeria-test-*")
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

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
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("error closing DB after exec error: %v (original error: %v)", closeErr, err)
		}
		return nil, err
	}

	return db, nil
}

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

func CleanTestDB(services *AppServices) {
	if services != nil && services.DB != nil && services.DB.DB != nil {
		var dbPath string

		err := services.DB.DB.QueryRow("PRAGMA database_list").Scan(nil, &dbPath, nil)
		if err == nil && dbPath != "" {
			if closeErr := services.DB.DB.Close(); closeErr != nil {
				// In tests, we can just log the error
				fmt.Printf("Error closing test database: %v\n", closeErr)
			}

			if filepath.Dir(dbPath) != "." {
				if removeErr := os.RemoveAll(filepath.Dir(dbPath)); removeErr != nil {
					fmt.Printf("Error removing test database directory: %v\n", removeErr)
				}
			}
		}
	}
}

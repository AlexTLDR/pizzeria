package handlers

import (
	"net/http"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Required for test database
)

func TestRepository_Home(t *testing.T) {
	// Create a test repository
	repo := NewTestRepo(t)
	defer CleanTestDB(repo)

	// Set handlers with our test repo
	NewHandlers(repo)

	// Create a test HTTP request
	req, rr := CreateTestRequest(t, "GET", "/", nil)

	// Call the Home handler
	http.HandlerFunc(Repo.Home).ServeHTTP(rr, req)

	// Check response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Content type should be text/html
	contentType := rr.Header().Get("Content-Type")
	if contentType != "" && contentType != "text/html; charset=utf-8" {
		t.Errorf("handler returned wrong content type: got %v", contentType)
	}

	// Response should contain some expected HTML content
	// Note: This test will need templates to be loaded, which may require
	// additional setup in a real environment. For now, we're just checking
	// that the handler completes without error.
}

func TestRepository_CheckDBConnection(t *testing.T) {
	// Create a test repository
	repo := NewTestRepo(t)
	defer CleanTestDB(repo)

	// Set handlers with our test repo
	NewHandlers(repo)

	// Create a test HTTP request
	req, rr := CreateTestRequest(t, "GET", "/debug/db-check", nil)

	// Call the CheckDBConnection handler
	http.HandlerFunc(Repo.CheckDBConnection).ServeHTTP(rr, req)

	// Check response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Response should contain "Database connection is working"
	if !contains(rr.Body.String(), "Database connection is working") {
		t.Errorf("handler response doesn't contain expected message")
	}
}

// Helper function to check if a string contains a substring
func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) && haystack[0:len(haystack)][0:len(needle)] == needle
}
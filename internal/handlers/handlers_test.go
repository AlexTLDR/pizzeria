package handlers

import (
	"net/http"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Required for test database
)

func TestRepository_Home(t *testing.T) {
	repo := NewTestRepo(t)
	defer CleanTestDB(repo)

	NewHandlers(repo)

	req, rr := CreateTestRequest(t, "GET", "/", nil)

	http.HandlerFunc(Repo.Home).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "Mock Template for Testing") {
		t.Errorf("handler response doesn't contain expected mock template content")
	}
}

func TestRepository_CheckDBConnection(t *testing.T) {
	repo := NewTestRepo(t)
	defer CleanTestDB(repo)

	NewHandlers(repo)

	req, rr := CreateTestRequest(t, "GET", "/debug/db-check", nil)

	http.HandlerFunc(Repo.CheckDBConnection).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "Database Connection Status: OK") {
		t.Errorf("handler response doesn't contain expected database status message")
	}
	
	if !strings.Contains(rr.Body.String(), "Test Query Status: OK") {
		t.Errorf("handler response doesn't contain expected query status message")
	}
}
package handlers

import (
	"net/http"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestAppServices_Home(t *testing.T) {
	services := NewTestAppServices(t)
	defer CleanTestDB(services)

	NewHandlers(services)

	req, rr := CreateTestRequest(t, "GET", "/", nil)

	http.HandlerFunc(Services.Home).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "Mock Template for Testing") {
		t.Errorf("handler response doesn't contain expected mock template content")
	}
}

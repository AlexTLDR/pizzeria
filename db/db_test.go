package db

import (
	"testing"
	"os"
	"path/filepath"
)

func TestInitDB(t *testing.T) {
	// Create a temporary directory for our test database
	tempDir, err := os.MkdirTemp("", "pizzeria-db-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a database path in the temp directory
	dbPath := filepath.Join(tempDir, "test.db")

	// Test creating a new database
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// Verify the connection works by executing a simple query
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Errorf("Failed to create test table: %v", err)
	}

	// Insert a row
	_, err = db.Exec("INSERT INTO test (name) VALUES (?)", "test value")
	if err != nil {
		t.Errorf("Failed to insert test data: %v", err)
	}

	// Query the row
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query test data: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 row, got %d", count)
	}

	// Test with invalid path
	invalidDB, err := InitDB("/invalid/path/that/should/not/exist.db")
	if err == nil {
		invalidDB.Close()
		t.Error("Expected error with invalid path, got nil")
	}
}

func TestInitDB_WithExistingDB(t *testing.T) {
	// Create a temporary directory for our test database
	tempDir, err := os.MkdirTemp("", "pizzeria-db-test-existing-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a database path in the temp directory
	dbPath := filepath.Join(tempDir, "existing.db")

	// Create a database file
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create initial test database: %v", err)
	}

	// Create a table and add data
	_, err = db.Exec("CREATE TABLE persistent (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	_, err = db.Exec("INSERT INTO persistent (value) VALUES (?)", "original value")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Close the database
	db.Close()

	// Now test opening the existing database
	reopenedDB, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open existing database: %v", err)
	}
	defer reopenedDB.Close()

	// Verify the data persisted
	var value string
	err = reopenedDB.QueryRow("SELECT value FROM persistent WHERE id = 1").Scan(&value)
	if err != nil {
		t.Errorf("Failed to query persistent data: %v", err)
	}

	if value != "original value" {
		t.Errorf("Expected 'original value', got '%s'", value)
	}
}
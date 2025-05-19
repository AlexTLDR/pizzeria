package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// CheckDBConnection is a diagnostic handler to check DB health
func (m *Repository) CheckDBConnection(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting database connection check...")

	// Check if DB field is properly initialized
	if m.DB == nil {
		msg := "ERROR: DB is nil in Repository"
		log.Println(msg)
		m.adminError(w, r, fmt.Errorf("DB is nil"), http.StatusInternalServerError, "CheckDBConnection - DB check")
		return
	}

	// Ping database
	var testValue int
	err := m.DB.DB.QueryRow("SELECT 1").Scan(&testValue)
	if err != nil {
		m.adminError(w, r, err, http.StatusInternalServerError, "CheckDBConnection - Database ping")
		return
	}

	// Try to count menu items
	var count int
	err = m.DB.DB.QueryRow("SELECT COUNT(*) FROM menu_items").Scan(&count)
	if err != nil {
		m.adminError(w, r, err, http.StatusInternalServerError, "CheckDBConnection - Counting menu items")
		return
	}

	// Try to get menu items directly using SQL
	rows, err := m.DB.DB.Query("SELECT id, name, description, category, price, image_url FROM menu_items")
	if err != nil {
		m.adminError(w, r, err, http.StatusInternalServerError, "CheckDBConnection - Querying menu items")
		return
	}
	defer rows.Close()

	// Count the rows from query result
	var directCount int
	for rows.Next() {
		directCount++
	}

	// Generate diagnostics report
	report := fmt.Sprintf(`
Database Connection Status: OK
Test Query Status: OK (result: %d)
Menu Items Table Count: %d
Direct Query Row Count: %d
Current Time: %s
`, testValue, count, directCount, time.Now().Format(time.RFC3339))

	// Log report
	log.Println(report)

	// Write report to a file for persistence
	diagFile := "diagnostics.txt"
	err = os.WriteFile(diagFile, []byte(report), 0644)
	if err != nil {
		log.Printf("Error writing diagnostics file: %v", err)
	}

	// Return success with diagnostics
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, report)
}

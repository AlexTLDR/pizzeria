package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// TestImageOperations tests image file operations
func (m *Repository) TestImageOperations(w http.ResponseWriter, r *http.Request) {
	// Create a test file
	testDir := filepath.Join("static", "images", "menu")
	os.MkdirAll(testDir, 0755) // Ensure directory exists

	// Create unique filename based on timestamp
	timestamp := time.Now().Unix()
	testFilename := fmt.Sprintf("%d_test-image.txt", timestamp)
	testFilePath := filepath.Join(testDir, testFilename)

	// Write test content to file
	testContent := "This is a test file for image operations"
	err := os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		msg := fmt.Sprintf("ERROR: Failed to create test file: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created test file: %s", testFilePath)

	// Test deletion with and without leading slash
	testPaths := []string{
		testFilePath,       // Without leading slash
		"/" + testFilePath, // With leading slash
		filepath.Join(testDir, "nonexistent-file.txt"), // Nonexistent file
		"invalid/path/outside/menu",                    // Invalid path outside menu
	}

	results := []string{
		fmt.Sprintf("Created test file: %s", testFilePath),
	}

	// Test deletion
	for _, path := range testPaths {
		log.Printf("Testing deletion of: %s", path)
		m.deleteImageFile(path)

		// Check if file exists after deletion attempt
		var exists bool
		if fileInfo, err := os.Stat(testFilePath); err == nil && !fileInfo.IsDir() {
			exists = true
		}

		results = append(results, fmt.Sprintf("Path: %s, Still exists: %t", path, exists))
	}

	// If test file still exists after all tests, clean it up
	if fileInfo, err := os.Stat(testFilePath); err == nil && !fileInfo.IsDir() {
		os.Remove(testFilePath)
		results = append(results, "Cleaned up test file manually at end of test")
	}

	// Return results
	w.Header().Set("Content-Type", "text/plain")
	for _, result := range results {
		fmt.Fprintln(w, result)
	}
}

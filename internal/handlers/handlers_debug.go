package handlers

import (
	"encoding/json"

	"log"
	"net/http"
)

// DebugMenuItems displays raw menu items data for debugging
func (m *Repository) DebugMenuItems(w http.ResponseWriter, r *http.Request) {
	log.Println("DEBUG: Retrieving menu items from database...")

	// Query data directly from the database
	menuItems, err := m.DB.GetAllMenuItems()
	if err != nil {
		m.adminError(w, r, err, http.StatusInternalServerError, "DebugMenuItems - fetching menu items")
		return
	}

	log.Printf("DEBUG: Retrieved %d menu items", len(menuItems))

	// Return as JSON for easy debugging
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": len(menuItems),
		"items": menuItems,
	})
}

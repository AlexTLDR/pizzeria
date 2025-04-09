package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/AlexTLDR/pizzeria/internal/models"
)

// Home handles the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	// Use the model's method to get menu items - this properly handles NULL small_price values
	log.Println("Fetching menu items using model's GetAllMenuItems method")
	menuItems, err := m.DB.GetAllMenuItems()
	if err != nil {
		log.Printf("ERROR fetching menu items: %v", err)
		http.Error(w, "Error fetching menu items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Retrieved %d menu items directly via SQL", len(menuItems))

	// Get active flash messages
	flashMessages, err := m.DB.GetActiveFlashMessages()
	if err != nil {
		// Just log the error, don't fail the page load
		log.Printf("Error fetching flash messages: %v", err)
	}

	// Group menu items by category
	menuByCategory := make(map[string][]models.MenuItem)
	var categories []string
	categorySet := make(map[string]bool)

	// First, collect all unique categories
	for _, item := range menuItems {
		if !categorySet[item.Category] {
			categories = append(categories, item.Category)
			categorySet[item.Category] = true
		}
		menuByCategory[item.Category] = append(menuByCategory[item.Category], item)
	}

	log.Printf("Categories found: %v", categories)

	// Render template with categories and menu items
	err = m.TemplateCache["index.html"].Execute(w, map[string]interface{}{
		"Title":          "La Piccola Sardegna",
		"Categories":     categories,
		"MenuByCategory": menuByCategory,
		"Menu":           menuItems,
		"FlashMessages":  flashMessages,
		"Year":           time.Now().Year(),
	})

	log.Printf("Template rendered with %d menu items", len(menuItems))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

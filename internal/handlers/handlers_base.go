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

	// Create a new slice for simplified menu items
	var simplifiedMenuItems []models.MenuItem

	// First, collect all unique categories and map full category names to simple ones for the template
	for _, item := range menuItems {
		// Clone the item for modification
		simplifiedItem := item

		// Simplified category name for template conditionals
		var simpleCategory string

		// Extract simpler category name from database category
		if item.Category == "Antipasti / Vorspeisen" {
			simpleCategory = "Antipasti"
		} else if item.Category == "Insalate / Salate" {
			simpleCategory = "Insalate"
		} else if item.Category == "Carne / Fleisch" {
			simpleCategory = "Carne"
		} else if item.Category == "Pesce Fritto / Fisch fritiert" {
			simpleCategory = "Pesce Fritto"
		} else if item.Category == "Pasta al Forno / Nudelgerichte überbacken" {
			simpleCategory = "Pasta al Forno"
		} else {
			// For categories that don't need renaming (Pizza, Spaghetti, Penne, Rigatoni)
			simpleCategory = item.Category
		}

		// Update the simplified item's category
		simplifiedItem.Category = simpleCategory

		// Add to the simplified menu items slice
		simplifiedMenuItems = append(simplifiedMenuItems, simplifiedItem)

		if !categorySet[item.Category] {
			categories = append(categories, item.Category)
			categorySet[item.Category] = true
		}

		// Use the simplified category for the menu by category map
		menuByCategory[item.Category] = append(menuByCategory[item.Category], simplifiedItem)
	}

	log.Printf("Categories found: %v", categories)

	// Render template with categories and menu items
	err = m.TemplateCache["index.html"].Execute(w, map[string]interface{}{
		"Title":          "La Piccola Sardegna",
		"Categories":     categories,
		"MenuByCategory": menuByCategory,
		"Menu":           simplifiedMenuItems,
		"FlashMessages":  flashMessages,
		"Year":           time.Now().Year(),
	})

	log.Printf("Template rendered with %d menu items", len(menuItems))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

package handlers

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AlexTLDR/pizzeria/internal/models"
)

// deleteImageFile safely deletes an image file
func (m *Repository) deleteImageFile(imageURL string) {
	// Skip if empty URL
	if imageURL == "" {
		return
	}

	// Remove the leading slash if present
	if strings.HasPrefix(imageURL, "/") {
		imageURL = imageURL[1:]
	}

	// Make sure the file exists and is within the menu images directory
	if !strings.Contains(imageURL, "static/images/menu") {
		return
	}

	// Attempt to delete the file
	err := os.Remove(imageURL)
	if err != nil {
		// Just log the error but don't throw an exception
		log.Printf("Error deleting image file %s: %v", imageURL, err)
	} else {
		log.Printf("Successfully deleted old image file: %s", imageURL)
	}
}

// ShowCreateMenuItem displays the create menu item form
func (m *Repository) ShowCreateMenuItem(w http.ResponseWriter, r *http.Request) {
	// Render the menu form template
	err := m.TemplateCache["menu-form.html"].Execute(w, map[string]interface{}{
		"Title":    "Create Menu Item",
		"FormType": "create",
		"Year":     time.Now().Year(),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateMenuItem handles the create menu item form submission
func (m *Repository) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	name := r.FormValue("name")
	description := r.FormValue("description")
	categoryRaw := r.FormValue("category")
	priceStr := r.FormValue("price")

	// Basic validation
	if name == "" || description == "" || categoryRaw == "" || priceStr == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Format category (capitalize first letter, lowercase rest)
	category := strings.ToLower(categoryRaw)
	if len(category) > 0 {
		category = strings.ToUpper(category[:1]) + category[1:]
	}

	// Parse price
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}

	// Handle the uploaded image
	var imageURL string
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Create unique filename based on timestamp
		timestamp := time.Now().Unix()
		filename := fmt.Sprintf("%d_%s", timestamp, header.Filename)

		// Ensure filename contains only valid characters
		filename = strings.ReplaceAll(filename, " ", "-")

		// Save the file
		filePath := filepath.Join("static", "images", "menu", filename)
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Could not save image", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the file content
		_, err = dst.ReadFrom(file)
		if err != nil {
			http.Error(w, "Could not save image", http.StatusInternalServerError)
			return
		}

		// Set the image URL
		imageURL = "/" + filePath // Add leading slash for web URLs
	}

	// Create menu item
	item := models.MenuItem{
		Name:        name,
		Description: description,
		Category:    category,
		Price:       price,
		ImageURL:    imageURL,
	}

	// Save to database
	_, err = m.DB.InsertMenuItem(item)
	if err != nil {
		http.Error(w, "Could not save menu item", http.StatusInternalServerError)
		return
	}

	// Redirect to admin dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// ShowEditMenuItem displays the edit menu item form
func (m *Repository) ShowEditMenuItem(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	id := r.URL.Path[len("/admin/menu/edit/"):]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get the menu item
	item, err := m.DB.GetMenuItemByID(idInt)
	if err != nil {
		http.Error(w, "Menu item not found", http.StatusNotFound)
		return
	}

	// Render the menu form template
	err = m.TemplateCache["menu-form.html"].Execute(w, map[string]interface{}{
		"Title":    "Edit Menu Item",
		"FormType": "edit",
		"Item":     item,
		"Year":     time.Now().Year(),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// UpdateMenuItem handles the update menu item form submission
func (m *Repository) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	id := r.URL.Path[len("/admin/menu/update/"):]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get existing item to check for image changes
	existingItem, err := m.DB.GetMenuItemByID(idInt)
	if err != nil {
		http.Error(w, "Menu item not found", http.StatusNotFound)
		return
	}
	// Parse the form data
	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	name := r.FormValue("name")
	description := r.FormValue("description")
	categoryRaw := r.FormValue("category")
	priceStr := r.FormValue("price")
	removeImage := r.FormValue("remove_image") // New field to check if image should be removed

	// Basic validation
	if name == "" || description == "" || categoryRaw == "" || priceStr == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Format category
	category := strings.ToLower(categoryRaw)
	if len(category) > 0 {
		category = strings.ToUpper(category[:1]) + category[1:]
	}

	// Parse price
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}

	// Handle the uploaded image
	var imageURL = existingItem.ImageURL // Default to existing image

	// Check if user wants to remove the image without replacing it
	if removeImage == "yes" && existingItem.ImageURL != "" {
		// Delete the existing image
		m.deleteImageFile(existingItem.ImageURL)
		imageURL = "" // Set image URL to empty
	} else {
		// Check if a new image was uploaded
		var file multipart.File
		var header *multipart.FileHeader
		file, header, err = r.FormFile("image")

		// Only process if a new image was uploaded
		if err == nil {
			defer file.Close()

			// If we're changing the image, delete the old one
			if existingItem.ImageURL != "" {
				m.deleteImageFile(existingItem.ImageURL)
			}

			// Create unique filename based on timestamp
			timestamp := time.Now().Unix()
			filename := fmt.Sprintf("%d_%s", timestamp, header.Filename)

			// Ensure filename contains only valid characters
			filename = strings.ReplaceAll(filename, " ", "-")

			// Save the file
			filePath := filepath.Join("static", "images", "menu", filename)
			dst, err := os.Create(filePath)
			if err != nil {
				http.Error(w, "Could not save image", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			// Copy the file content
			_, err = dst.ReadFrom(file)
			if err != nil {
				http.Error(w, "Could not save image", http.StatusInternalServerError)
				return
			}

			// Set the new image URL
			imageURL = "/" + filePath // Add leading slash for web URLs
		}
	}

	// Create menu item
	item := models.MenuItem{
		ID:          idInt,
		Name:        name,
		Description: description,
		Category:    category,
		Price:       price,
		ImageURL:    imageURL,
	}

	// Update in database
	err = m.DB.UpdateMenuItem(item)
	if err != nil {
		http.Error(w, "Could not update menu item", http.StatusInternalServerError)
		return
	}

	// Redirect to admin dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// DeleteMenuItem handles the deletion of a menu item
func (m *Repository) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	id := r.URL.Path[len("/admin/menu/delete/"):]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get the menu item to check if it has an image
	item, err := m.DB.GetMenuItemByID(idInt)
	if err == nil && item.ImageURL != "" {
		m.deleteImageFile(item.ImageURL)
	}

	// Delete from database
	err = m.DB.DeleteMenuItem(idInt)
	if err != nil {
		http.Error(w, "Could not delete menu item", http.StatusInternalServerError)
		return
	}

	// Redirect to admin dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

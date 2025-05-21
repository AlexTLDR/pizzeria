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

	"github.com/google/uuid"

	"github.com/AlexTLDR/pizzeria/internal/models"
)

// deleteImageFile safely deletes an image file
func (m *Repository) deleteImageFile(imageURL string) {
	// Skip if empty URL
	if imageURL == "" {
		return
	}

	// Remove the leading slash if present
	imageURL = strings.TrimPrefix(imageURL, "/")

	// Make sure the file exists and is within the menu images directory
	if !strings.Contains(imageURL, "static/images/menu") {
		log.Printf("Warning: Attempted to delete image outside of menu directory: %s", imageURL)
		return
	}

	// Get the absolute file path
	filePath := imageURL // The path should now be relative without leading slash

	// Verify the file exists before attempting to delete
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Warning: Image file does not exist: %s", filePath)
		return
	}

	// Attempt to delete the file
	err := os.Remove(filePath)
	if err != nil {
		// Just log the error but don't throw an exception
		log.Printf("Error deleting image file %s: %v", filePath, err)
	} else {
		log.Printf("Successfully deleted old image file: %s", filePath)
	}
}

// ShowCreateMenuItem displays the create menu item form
func (m *Repository) ShowCreateMenuItem(w http.ResponseWriter, _ *http.Request) {
	// Render the menu form template
	err := m.TemplateCache["menu-form.html"].Execute(w, map[string]interface{}{
		"Title":    "Create Menu Item",
		"FormType": "create",
		"Year":     time.Now().Year(),
	})

	if err != nil {
		// Just log the error since template.Execute likely already wrote to the response
		log.Printf("ERROR: Template rendering failed in ShowCreateMenuItem: %v", err)
		return
	}
}

// ShowEditMenuItem displays the edit menu item form
func (m *Repository) ShowEditMenuItem(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	id := r.URL.Path[len("/admin/menu/edit/"):]

	idInt, err := strconv.Atoi(id)
	if err != nil {
		m.clientError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Get the menu item
	item, err := m.DB.GetMenuItemByID(idInt)
	if err != nil {
		m.clientError(w, http.StatusNotFound, "Menu item not found")
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
		// Just log the error since template.Execute likely already wrote to the response
		log.Printf("ERROR: Template rendering failed in ShowEditMenuItem: %v", err)
		return
	}
}

// isValidImageExtension checks if the file has an allowed image extension
func (m *Repository) isValidImageExtension(filename string) bool {
	extension := strings.ToLower(filepath.Ext(filename))

	// Define allowed image extensions
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
		".svg":  true,
	}

	return allowedExtensions[extension]
}

// CreateMenuItem handles the create menu item form submission
func (m *Repository) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	// Check if this is a GET request - if so, show the form instead of processing it
	if r.Method == "GET" {
		m.ShowCreateMenuItem(w, r)
		return
	}

	// For POST requests, parse the form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		m.clientError(w, http.StatusBadRequest, "Could not parse form")
		return
	}

	// Get form values
	name := r.FormValue("name")
	description := r.FormValue("description")
	categoryRaw := r.FormValue("category")
	priceStr := r.FormValue("price")
	smallPriceStr := r.FormValue("small_price")

	// Basic validation
	if name == "" || description == "" || categoryRaw == "" || priceStr == "" {
		m.clientError(w, http.StatusBadRequest, "All fields are required")
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
		m.clientError(w, http.StatusBadRequest, "Invalid price")
		return
	}

	// Parse small price if provided
	var smallPrice *float64

	if smallPriceStr != "" {
		smallPriceValue, err := strconv.ParseFloat(smallPriceStr, 64)
		if err != nil {
			m.clientError(w, http.StatusBadRequest, "Invalid small price")
			return
		}

		smallPrice = &smallPriceValue
	}

	// Handle the uploaded image
	var imageURL string

	file, header, err := r.FormFile("image_upload")
	if err == nil {
		defer file.Close()

		// Validate that the file is an image
		if !m.isValidImageExtension(header.Filename) {
			m.clientError(w, http.StatusBadRequest, "Invalid file type. Only image files (jpg, jpeg, png, gif, webp, bmp, svg) are allowed.")
			return
		}

		// Create a completely random filename with timestamp prefix
		timestamp := time.Now().Unix()
		extension := filepath.Ext(header.Filename) // Get the file extension
		randomName := fmt.Sprintf("%d_%s%s", timestamp, uuid.New().String(), extension)

		// Save the file
		filePath := filepath.Join("static", "images", "menu", randomName)

		dst, err := os.Create(filePath)
		if err != nil {
			m.adminError(w, r, err, http.StatusInternalServerError, "CreateMenuItem - saving image")
			return
		}

		defer dst.Close()

		// Copy the file content
		_, err = dst.ReadFrom(file)
		if err != nil {
			m.adminError(w, r, err, http.StatusInternalServerError, "CreateMenuItem - saving image")
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
		SmallPrice:  smallPrice,
		ImageURL:    imageURL,
	}

	// Save to database
	_, err = m.DB.InsertMenuItem(item)
	if err != nil {
		m.adminError(w, r, err, http.StatusInternalServerError, "CreateMenuItem - saving menu item")
		return
	}

	// Redirect to admin dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// UpdateMenuItem handles the update menu item form submission
func (m *Repository) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	id := r.URL.Path[len("/admin/menu/update/"):]

	idInt, err := strconv.Atoi(id)
	if err != nil {
		m.clientError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Check if this is a GET request - if so, redirect to the edit form
	if r.Method == "GET" {
		http.Redirect(w, r, "/admin/menu/edit/"+id, http.StatusSeeOther)
		return
	}

	// Get existing item to check for image changes
	existingItem, err := m.DB.GetMenuItemByID(idInt)
	if err != nil {
		m.clientError(w, http.StatusNotFound, "Menu item not found")
		return
	}

	// Parse the form data
	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		m.clientError(w, http.StatusBadRequest, "Could not parse form")
		return
	}

	// Get form values
	name := r.FormValue("name")
	description := r.FormValue("description")
	categoryRaw := r.FormValue("category")
	priceStr := r.FormValue("price")
	smallPriceStr := r.FormValue("small_price")
	removeImage := r.FormValue("remove_image")

	// Basic validation
	if name == "" || description == "" || categoryRaw == "" || priceStr == "" {
		m.clientError(w, http.StatusBadRequest, "All fields are required")
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
		m.clientError(w, http.StatusBadRequest, "Invalid price")
		return
	}

	// Parse small price if provided
	var smallPrice *float64

	if smallPriceStr != "" {
		smallPriceValue, err := strconv.ParseFloat(smallPriceStr, 64)
		if err != nil {
			m.clientError(w, http.StatusBadRequest, "Invalid small price")
			return
		}

		smallPrice = &smallPriceValue
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
		file, header, err = r.FormFile("image_upload")

		// Only process if a new image was uploaded
		if err == nil {
			defer file.Close()

			// Validate that the file is an image
			if !m.isValidImageExtension(header.Filename) {
				m.clientError(w, http.StatusBadRequest, "Invalid file type. Only image files (jpg, jpeg, png, gif, webp, bmp, svg) are allowed.")
				return
			}

			// If we're changing the image, delete the old one
			if existingItem.ImageURL != "" {
				m.deleteImageFile(existingItem.ImageURL)
			}

			// Create a completely random filename with timestamp prefix
			timestamp := time.Now().Unix()
			extension := filepath.Ext(header.Filename) // Get the file extension
			randomName := fmt.Sprintf("%d_%s%s", timestamp, uuid.New().String(), extension)

			// Save the file
			filePath := filepath.Join("static", "images", "menu", randomName)

			dst, err := os.Create(filePath)
			if err != nil {
				m.adminError(w, r, err, http.StatusInternalServerError, "UpdateMenuItem - saving image")
				return
			}

			defer dst.Close()

			// Copy the file content
			_, err = dst.ReadFrom(file)
			if err != nil {
				m.adminError(w, r, err, http.StatusInternalServerError, "UpdateMenuItem - saving image")
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
		SmallPrice:  smallPrice,
		ImageURL:    imageURL,
	}

	// Update in database
	err = m.DB.UpdateMenuItem(item)
	if err != nil {
		m.adminError(w, r, err, http.StatusInternalServerError, "UpdateMenuItem - updating menu item")
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
		m.clientError(w, http.StatusBadRequest, "Invalid ID")
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
		m.adminError(w, r, err, http.StatusInternalServerError, "DeleteMenuItem - deleting menu item")
		return
	}

	// Redirect to admin dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

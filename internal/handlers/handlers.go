package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AlexTLDR/pizzeria/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	DB            *models.DBModel
	TemplateCache map[string]*template.Template
}

var Repo *Repository

func NewRepo(db *models.DBModel, tc map[string]*template.Template) *Repository {
	return &Repository{
		DB:            db,
		TemplateCache: tc,
	}
}

func NewHandlers(r *Repository) {
	Repo = r
}

// Home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	menuItems, err := m.DB.GetAllMenuItems()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get active flash messages
	activeMessages, err := m.DB.GetActiveFlashMessages()
	if err != nil {
		log.Printf("Error getting flash messages: %v", err)
		activeMessages = []models.FlashMessage{}
	}

	data := map[string]interface{}{
		"Title":         "Pizzeria Ristorante - Authentic Italian Cuisine",
		"Menu":          menuItems,
		"Year":          time.Now().Year(),
		"FlashMessages": activeMessages,
	}

	err = m.TemplateCache["index.html"].Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Admin login handlers
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}

	// Check for error parameter in URL
	if r.URL.Query().Get("error") == "invalid" {
		data["Error"] = true
	}

	err := m.TemplateCache["login.html"].Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")

	user, err := m.DB.GetUserByUsername(username)
	if err != nil {
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "authenticated", // Use a proper session ID in production
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Admin dashboard handler
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	menuItems, err := m.DB.GetAllMenuItems()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get all flash messages for admin display
	flashMessages, err := m.DB.GetAllFlashMessages()
	if err != nil {
		// Just log the error but continue
		flashMessages = []models.FlashMessage{}
	}

	// Get success and error messages from the URL query parameters
	success := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	data := map[string]interface{}{
		"Title":         "Admin Dashboard",
		"Menu":          menuItems,
		"Year":          time.Now().Year(),
		"Success":       success,
		"Error":         errorMsg,
		"FlashMessages": flashMessages,
	}

	err = m.TemplateCache["admin-dashboard.html"].Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateFlashMessage handles the creation of a new flash message
func (m *Repository) CreateFlashMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// Parse form data
	messageType := "info" // Simplified to just use info type
	messageText := r.Form.Get("message")
	startDateStr := r.Form.Get("start_date")
	endDateStr := r.Form.Get("end_date")
	active := r.Form.Get("active") == "on"

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=Invalid start date format", http.StatusSeeOther)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=Invalid end date format", http.StatusSeeOther)
		return
	}

	// Validate dates
	if endDate.Before(startDate) {
		http.Redirect(w, r, "/admin/dashboard?error=End date cannot be before start date", http.StatusSeeOther)
		return
	}

	// Create flash message
	message := models.FlashMessage{
		Type:      messageType,
		Message:   messageText,
		StartDate: startDate,
		EndDate:   endDate,
		Active:    active,
	}

	_, err = m.DB.CreateFlashMessage(message)
	if err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=Failed to create announcement", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard?success=Announcement created successfully", http.StatusSeeOther)
}

// DeleteFlashMessage handles deleting a flash message
func (m *Repository) DeleteFlashMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the ID from the URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	err = m.DB.DeleteFlashMessage(id)
	if err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=Failed to delete announcement", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard?success=Announcement deleted successfully", http.StatusSeeOther)
}

// ShowCreateMenuItem displays the form to create a new menu item
func (m *Repository) ShowCreateMenuItem(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Add New Menu Item",
		"Item":  models.MenuItem{}, // Empty item for the form
		"Year":  time.Now().Year(),
	}

	err := m.TemplateCache["menu-form.html"].Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ShowEditMenuItem displays the form to edit an existing menu item
func (m *Repository) ShowEditMenuItem(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Get the menu item
	item, err := m.DB.GetMenuItemByID(id)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Title": "Edit Menu Item",
		"Item":  item,
		"Year":  time.Now().Year(),
	}

	err = m.TemplateCache["menu-form.html"].Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// UpdateMenuItem handles updating an existing menu item
func (m *Repository) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the ID from the URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form to handle file uploads
	err = r.ParseMultipartForm(10 << 20) // 10 MB max memory
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// Get existing item to retrieve current imageURL
	existingItem, err := m.DB.GetMenuItemByID(id)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)

	var smallPrice *float64
	if sp := r.FormValue("small_price"); sp != "" {
		spVal, _ := strconv.ParseFloat(sp, 64)
		smallPrice = &spVal
	}

	// Initialize updated item with form values
	item := models.MenuItem{
		ID:          id,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Price:       price,
		SmallPrice:  smallPrice,
		Category:    r.FormValue("category"),
		ImageURL:    existingItem.ImageURL, // Default to current image URL
	}

	// Handle file upload if provided
	file, handler, err := r.FormFile("image_upload")
	if err == nil && handler != nil {
		defer file.Close()

		// Create a unique filename using timestamp
		timestamp := time.Now().Unix()
		filename := fmt.Sprintf("%d_%s", timestamp, handler.Filename)

		// Ensure the directory exists
		uploadDir := "static/images/menu"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, 0755)
		}

		// Create file path
		filePath := filepath.Join(uploadDir, filename)

		// Create the file
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to create file on server", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the created file on the filesystem
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}

		// If we've successfully uploaded a new file, delete the old one if it exists
		if existingItem.ImageURL != "" {
			m.deleteImageFile(existingItem.ImageURL)
		}

		// Update the item's image URL
		item.ImageURL = "/" + filePath // Add leading slash for web path
	}

	// Update the item in the database
	err = m.DB.UpdateMenuItem(item)
	if err != nil {
		http.Error(w, "Could not update menu item", http.StatusInternalServerError)
		return
	}

	// Redirect to the dashboard with success message
	http.Redirect(w, r, "/admin/dashboard?success=Menu item updated successfully", http.StatusSeeOther)
}

// DeleteMenuItem handles deleting a menu item
func (m *Repository) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the ID from the URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Get the item before deletion to access its image URL
	item, err := m.DB.GetMenuItemByID(id)
	if err == nil && item.ImageURL != "" {
		m.deleteImageFile(item.ImageURL)
	}

	err = m.DB.DeleteMenuItem(id)
	if err != nil {
		http.Error(w, "Could not delete menu item", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// Menu item CRUD handlers
func (m *Repository) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form to handle file uploads
	err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)

	var smallPrice *float64
	if sp := r.FormValue("small_price"); sp != "" {
		spVal, _ := strconv.ParseFloat(sp, 64)
		smallPrice = &spVal
	}

	// Initialize item with form values
	item := models.MenuItem{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Price:       price,
		SmallPrice:  smallPrice,
		Category:    r.FormValue("category"),
		ImageURL:    r.FormValue("image_url"),
	}

	// Handle file upload if provided
	file, handler, err := r.FormFile("image_upload")
	if err == nil && handler != nil {
		defer file.Close()

		// Create a unique filename using timestamp
		timestamp := time.Now().Unix()
		filename := fmt.Sprintf("%d_%s", timestamp, handler.Filename)

		// Ensure the directory exists
		uploadDir := "static/images/menu"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, 0755)
		}

		// Create file path
		filePath := filepath.Join(uploadDir, filename)

		// Create the file
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to create file on server", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the created file on the filesystem
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}

		// Update the item's image URL
		item.ImageURL = "/" + filePath // Add leading slash for web path
	}

	// Save the menu item to the database
	_, err = m.DB.InsertMenuItem(item)
	if err != nil {
		http.Error(w, "Could not create menu item", http.StatusInternalServerError)
		return
	}

	// Redirect to the dashboard with success message
	http.Redirect(w, r, "/admin/dashboard?success=Menu item created successfully", http.StatusSeeOther)
}

func (m *Repository) AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		http.Error(w, "Password hashing error", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	_, err = m.DB.InsertUser(user)
	if err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// Helper function to safely delete image files
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

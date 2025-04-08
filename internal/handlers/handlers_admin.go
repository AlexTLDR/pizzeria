package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexTLDR/pizzeria/internal/models"
)

// AdminDashboard displays the admin dashboard
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	// Direct SQL query as a fallback/temporary solution
	log.Println("Querying menu items directly from DB in AdminDashboard handler")
	rows, err := m.DB.DB.Query("SELECT id, name, description, category, price, image_url FROM menu_items")
	if err != nil {
		log.Printf("ERROR querying menu items: %v", err)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var menuItems []models.MenuItem
	for rows.Next() {
		var item models.MenuItem
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Category, &item.Price, &item.ImageURL)
		if err != nil {
			log.Printf("ERROR scanning menu item: %v", err)
			continue
		}
		menuItems = append(menuItems, item)
	}
	log.Printf("Retrieved %d menu items directly via SQL", len(menuItems))
	menuItemCount := len(menuItems)

	// Get flash messages
	flashMessages, err := m.DB.GetAllFlashMessages() // Changed to get all messages for admin
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	flashCount := len(flashMessages)

	// Get users
	users, err := m.DB.GetAllUsers()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	userCount := len(users)

	// Render the dashboard template
	err = m.TemplateCache["admin-dashboard.html"].Execute(w, map[string]interface{}{
		"Title":         "Admin Dashboard",
		"MenuItemCount": menuItemCount,
		"FlashMsgCount": flashCount,
		"UserCount":     userCount,
		"FlashMessages": flashMessages,
		"Menu":          menuItems, // Add menu items to the template context
		"Year":          time.Now().Year(),
	})

	if err != nil {
		log.Printf("ERROR rendering admin dashboard template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Printf("Admin dashboard rendered with %d menu items", len(menuItems))
}

// CreateFlashMessage handles creation of flash messages
func (m *Repository) CreateFlashMessage(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// Get data from form
	message := r.Form.Get("message")
	startDateStr := r.Form.Get("start_date")
	endDateStr := r.Form.Get("end_date")

	// Parse dates
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	// Create flash message
	flashMsg := models.FlashMessage{
		Message:   message,
		Type:      "info", // Default type
		Active:    true,   // Active by default
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Insert the flash message directly using SQL
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO flash_messages (type, message, start_date, end_date, active, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	result, err := m.DB.DB.ExecContext(ctx, stmt,
		flashMsg.Type,
		flashMsg.Message,
		flashMsg.StartDate,
		flashMsg.EndDate,
		flashMsg.Active,
		now,
		now,
	)

	if err != nil {
		http.Error(w, "Could not save flash message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the last inserted ID
	_, err = result.LastInsertId()
	if err != nil {
		http.Error(w, "Could not get inserted ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to admin dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// DeleteFlashMessage handles deletion of flash messages
func (m *Repository) DeleteFlashMessage(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	idStr := r.URL.Path[len("/admin/flash-message/delete/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Delete the flash message
	err = m.DB.DeleteFlashMessage(id)
	if err != nil {
		http.Error(w, "Could not delete flash message", http.StatusInternalServerError)
		return
	}

	// Redirect back to admin dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

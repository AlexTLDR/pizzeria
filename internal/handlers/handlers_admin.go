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
		m.adminError(w, r, err, http.StatusInternalServerError, "AdminDashboard - querying menu items")
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

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		m.adminError(w, r, err, http.StatusInternalServerError, "AdminDashboard - iterating menu items")
		return
	}

	log.Printf("Retrieved %d menu items directly via SQL", len(menuItems))
	menuItemCount := len(menuItems)

	// Get flash messages
	flashMessages, err := m.DB.GetAllFlashMessages() // Changed to get all messages for admin
	if err != nil {
		m.adminError(w, r, err, http.StatusInternalServerError, "AdminDashboard - fetching flash messages")
		return
	}

	flashCount := len(flashMessages)

	// Render the dashboard template
	err = m.TemplateCache["admin-dashboard.html"].Execute(w, map[string]interface{}{
		"Title":         "Admin Dashboard",
		"MenuItemCount": menuItemCount,
		"FlashMsgCount": flashCount,
		"FlashMessages": flashMessages,
		"Menu":          menuItems, // Add menu items to the template context
		"Year":          time.Now().Year(),
	})

	if err != nil {
		// Just log the error since template.Execute likely already wrote to the response
		log.Printf("ERROR: Template rendering failed in AdminDashboard: %v", err)
		return
	}

	log.Printf("Admin dashboard rendered with %d menu items", len(menuItems))
}

// CreateFlashMessage handles creation of flash messages
func (m *Repository) CreateFlashMessage(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		m.clientError(w, http.StatusBadRequest, "Could not parse form")
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
		m.clientError(w, http.StatusBadRequest, "Invalid start date")
		return
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		m.clientError(w, http.StatusBadRequest, "Invalid end date")
		return
	}

	// Create flash message
	flashMsg := models.FlashMessage{
		Message:   message,
		Type:      "info",
		Active:    true,
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
		m.adminError(w, r, err, http.StatusInternalServerError, "CreateFlashMessage - saving message")
		return
	}

	// Get the last inserted ID
	_, err = result.LastInsertId()
	if err != nil {
		m.adminError(w, r, err, http.StatusInternalServerError, "CreateFlashMessage - getting inserted ID")
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
		m.clientError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Delete the flash message
	err = m.DB.DeleteFlashMessage(id)
	if err != nil {
		m.adminError(w, r, err, http.StatusInternalServerError, "DeleteFlashMessage - deleting message")
		return
	}

	// Redirect back to admin dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// AdminRoot redirects from /admin to /admin/dashboard
func (m *Repository) AdminRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

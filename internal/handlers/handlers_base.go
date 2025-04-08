package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/AlexTLDR/pizzeria/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Home handles the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	// Direct SQL query as a fallback/temporary solution
	log.Println("Querying menu items directly from DB in Home handler")
	rows, err := m.DB.DB.Query("SELECT id, name, description, category, price, image_url FROM menu_items")
	if err != nil {
		log.Printf("ERROR querying menu items: %v", err)
		http.Error(w, "Error fetching menu items: "+err.Error(), http.StatusInternalServerError)
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
		"Menu":           menuItems, // Add menuItems under the "Menu" key for the template
		"FlashMessages":  flashMessages,
		"Year":           time.Now().Year(),
	})

	log.Printf("Template rendered with %d menu items", len(menuItems))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ShowLogin shows the login page
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	// Get any error message from the URL
	errorMsg := r.URL.Query().Get("error")

	// Render the login page template
	m.TemplateCache["login.html"].Execute(w, map[string]interface{}{
		"Title": "Login",
		"Error": errorMsg,
	})
}

// Login handles the login form submission
func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// Get the username and password from the form
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	// Validate inputs
	if username == "" || password == "" {
		http.Redirect(w, r, "/login?error=Please provide both username and password", http.StatusSeeOther)
		return
	}

	// Get the user from the database
	user, err := m.DB.GetUserByUsername(username)
	if err != nil {
		http.Redirect(w, r, "/login?error=Invalid credentials", http.StatusSeeOther)
		return
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login?error=Invalid credentials", http.StatusSeeOther)
		return
	}

	// Set a cookie to indicate that the user is logged in
	cookie := http.Cookie{
		Name:     "admin",
		Value:    "true",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600, // 1 hour
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	// Redirect to the admin dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// Logout handles user logout
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the admin cookie
	cookie := http.Cookie{
		Name:     "admin",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Delete the cookie
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

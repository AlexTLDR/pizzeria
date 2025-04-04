package handlers

import (
	"html/template"
	"net/http"
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

	data := map[string]interface{}{
		"Title": "Pizzeria Ristorante - Authentic Italian Cuisine",
		"Menu":  menuItems,
		"Year":  time.Now().Year(),
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

	data := map[string]interface{}{
		"Title": "Admin Dashboard",
		"Menu":  menuItems,
	}

	err = m.TemplateCache["admin-dashboard.html"].Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	price, _ := strconv.ParseFloat(r.Form.Get("price"), 64)

	var smallPrice *float64
	if sp := r.Form.Get("small_price"); sp != "" {
		spVal, _ := strconv.ParseFloat(sp, 64)
		smallPrice = &spVal
	}

	item := models.MenuItem{
		ID:          id,
		Name:        r.Form.Get("name"),
		Description: r.Form.Get("description"),
		Price:       price,
		SmallPrice:  smallPrice,
		Category:    r.Form.Get("category"),
		ImageURL:    r.Form.Get("image_url"),
	}

	err = m.DB.UpdateMenuItem(item)
	if err != nil {
		http.Error(w, "Could not update menu item", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
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

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	price, _ := strconv.ParseFloat(r.Form.Get("price"), 64)

	var smallPrice *float64
	if sp := r.Form.Get("small_price"); sp != "" {
		spVal, _ := strconv.ParseFloat(sp, 64)
		smallPrice = &spVal
	}

	item := models.MenuItem{
		Name:        r.Form.Get("name"),
		Description: r.Form.Get("description"),
		Price:       price,
		SmallPrice:  smallPrice,
		Category:    r.Form.Get("category"),
		ImageURL:    r.Form.Get("image_url"),
	}

	_, err = m.DB.InsertMenuItem(item)
	if err != nil {
		http.Error(w, "Could not create menu item", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
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

package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/AlexTLDR/pizzeria/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// ShowUserManagement displays the user management page
func (m *Repository) ShowUserManagement(w http.ResponseWriter, r *http.Request) {
	// Get all users
	users, err := m.DB.GetAllUsers()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get current user ID from the cookie (in a real app, you would use a proper session)
	// For now, we're just getting the username from the active session
	var currentUser models.User

	// In a real app with proper sessions, you would get the user ID from the session
	// Here we're just using the first user as a placeholder for the current user
	if len(users) > 0 {
		currentUser = users[0]
	}

	// Get success and error messages from the URL query parameters
	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	// Render the user management template
	err = m.TemplateCache["user-management.html"].Execute(w, map[string]interface{}{
		"Title":       "User Management",
		"Users":       users,
		"CurrentUser": currentUser,
		"SuccessMsg":  successMsg,
		"ErrorMsg":    errorMsg,
		"Year":        time.Now().Year(),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateUser handles the creation of a new user
func (m *Repository) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Basic validation
	if username == "" || password == "" || confirmPassword == "" {
		http.Redirect(w, r, "/admin/users?error=All fields are required", http.StatusSeeOther)
		return
	}

	if password != confirmPassword {
		http.Redirect(w, r, "/admin/users?error=Passwords do not match", http.StatusSeeOther)
		return
	}

	// Check if username already exists
	_, err = m.DB.GetUserByUsername(username)
	if err == nil {
		http.Redirect(w, r, "/admin/users?error=Username already exists", http.StatusSeeOther)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Could not hash password", http.StatusInternalServerError)
		return
	}

	// Create user
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

// UpdateUser handles updating user information
func (m *Repository) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	id := r.URL.Path[len("/admin/users/update/"):]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	username := r.FormValue("username")

	// Basic validation
	if username == "" {
		http.Redirect(w, r, "/admin/users?error=Username is required", http.StatusSeeOther)
		return
	}

	// Check if username is already taken by another user
	existingUser, err := m.DB.GetUserByUsername(username)
	if err == nil && existingUser.ID != idInt {
		http.Redirect(w, r, "/admin/users?error=Username already exists", http.StatusSeeOther)
		return
	}

	// Update user
	user := models.User{
		ID:       idInt,
		Username: username,
	}

	err = m.DB.UpdateUser(user)
	if err != nil {
		http.Error(w, "Could not update user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users?success=User updated successfully", http.StatusSeeOther)
}

// DeleteUser handles user deletion
func (m *Repository) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	id := r.URL.Path[len("/admin/users/delete/"):]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Delete user
	err = m.DB.DeleteUser(idInt)
	if err != nil {
		http.Error(w, "Could not delete user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users?success=User deleted successfully", http.StatusSeeOther)
}

// ChangePassword handles password changes
func (m *Repository) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	// Basic validation
	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		http.Redirect(w, r, "/admin/users?error=All password fields are required", http.StatusSeeOther)
		return
	}

	if newPassword != confirmPassword {
		http.Redirect(w, r, "/admin/users?error=New passwords do not match", http.StatusSeeOther)
		return
	}

	// Get the user
	user, err := m.DB.GetUserByID(userID)
	if err != nil {
		http.Redirect(w, r, "/admin/users?error=User not found", http.StatusSeeOther)
		return
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		http.Redirect(w, r, "/admin/users?error=Current password is incorrect", http.StatusSeeOther)
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Could not hash password", http.StatusInternalServerError)
		return
	}

	// Update the password - use the regular UpdateUser method
	user.PasswordHash = string(hashedPassword)
	err = m.DB.UpdateUser(user)
	if err != nil {
		http.Error(w, "Could not update password", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users?success=Password changed successfully", http.StatusSeeOther)
}

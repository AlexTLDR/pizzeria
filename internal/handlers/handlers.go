package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/AlexTLDR/pizzeria/internal/auth"
	"github.com/AlexTLDR/pizzeria/internal/middleware"
	"github.com/AlexTLDR/pizzeria/internal/models"
)

// Repository is the repository type
type Repository struct {
	DB            *models.DBModel
	TemplateCache map[string]*template.Template
	OAuthConfig   *auth.OAuthConfig
}

// serverError logs the error and returns a generic 500 error to the user
func (m *Repository) serverError(w http.ResponseWriter, err error, source string) {
	log.Printf("SERVER ERROR (%s): %v", source, err)
	http.Error(w, "Sorry, we're having technical difficulties. Please try again later.", http.StatusInternalServerError)
}

// clientError sends a specific status code and message to the client
func (m *Repository) clientError(w http.ResponseWriter, status int, message string) {
	http.Error(w, message, status)
}

// adminError logs the error and returns the error details to admin users
func (m *Repository) adminError(w http.ResponseWriter, r *http.Request, err error, status int, source string) {
	// Check if user is authenticated and authorized (admin)
	userEmail, valid := middleware.VerifySecureSessionCookie(r)
	isAdmin := valid && m.OAuthConfig.IsAllowedEmail(userEmail)

	log.Printf("ERROR (%s): %v", source, err)

	if isAdmin {
		// For admin users, show the actual error message
		http.Error(w, fmt.Sprintf("Error in %s: %v", source, err), status)
	} else {
		// For regular users, show a generic message
		http.Error(w, "Sorry, we're having technical difficulties. Please try again later.", status)
	}
}

// Repo is the repository used by the handlers
var Repo *Repository

// NewRepo creates a new repository
func NewRepo(db *models.DBModel, tc map[string]*template.Template, oa *auth.OAuthConfig) *Repository {
	return &Repository{
		DB:            db,
		TemplateCache: tc,
		OAuthConfig:   oa,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

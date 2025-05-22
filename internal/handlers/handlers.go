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

// AppServices provides services needed by the handlers
type AppServices struct {
	DB            *models.DBModel
	TemplateCache map[string]*template.Template
	OAuthConfig   *auth.OAuthConfig
}

// serverError logs the error and returns a generic 500 error to the user
func (m *AppServices) serverError(w http.ResponseWriter, err error, source string) {
	log.Printf("SERVER ERROR (%s): %v", source, err)
	http.Error(w, "Sorry, we're having technical difficulties. Please try again later.", http.StatusInternalServerError)
}

// clientError sends a specific status code and message to the client
func (m *AppServices) clientError(w http.ResponseWriter, status int, message string) {
	http.Error(w, message, status)
}

// adminError logs the error and returns the error details to admin users
func (m *AppServices) adminError(w http.ResponseWriter, r *http.Request, err error, status int, source string) {
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

// Services is the global app services instance used by the handlers
var Services *AppServices

// NewAppServices creates a new app services container
func NewAppServices(db *models.DBModel, tc map[string]*template.Template, oa *auth.OAuthConfig) *AppServices {
	return &AppServices{
		DB:            db,
		TemplateCache: tc,
		OAuthConfig:   oa,
	}
}

// NewHandlers sets the app services for the handlers
func NewHandlers(r *AppServices) {
	Services = r
}

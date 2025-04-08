package handlers

import (
	"html/template"

	"github.com/AlexTLDR/pizzeria/internal/models"
)

// Repository is the repository type
type Repository struct {
	DB            *models.DBModel
	TemplateCache map[string]*template.Template
}

// Repo is the repository used by the handlers
var Repo *Repository

// NewRepo creates a new repository
func NewRepo(db *models.DBModel, tc map[string]*template.Template) *Repository {
	return &Repository{
		DB:            db,
		TemplateCache: tc,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

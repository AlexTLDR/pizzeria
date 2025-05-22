package app

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	"github.com/AlexTLDR/pizzeria/db"
	"github.com/AlexTLDR/pizzeria/internal/auth"
	"github.com/AlexTLDR/pizzeria/internal/handlers"
	"github.com/AlexTLDR/pizzeria/internal/middleware"
	"github.com/AlexTLDR/pizzeria/internal/models"
)

type Application struct {
	DB            *sql.DB
	DBModel       *models.DBModel
	TemplateCache map[string]*template.Template
	OAuthConfig   *auth.OAuthConfig
	IsProduction  bool
}

// New creates a new application instance
func New() (*Application, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Determine environment (development or production)
	appEnv := os.Getenv("APP_ENV")
	isProduction := appEnv == "production"

	if isProduction {
		log.Println("Running in PRODUCTION mode - debug endpoints disabled")
	} else {
		log.Println("Running in DEVELOPMENT mode - debug endpoints enabled")
	}

	// Initialize database
	dbPath := "db/pizzeria.db"
	if err := runMigrations(dbPath); err != nil {
		log.Printf("Migration error: %v", err)
	}

	database, err := db.InitDB(dbPath)
	if err != nil {
		return nil, err
	}

	// Initialize OAuth
	if err := middleware.InitializeOAuth(); err != nil {
		log.Printf("OAuth initialization error: %v", err)
	}

	oauthConfig := middleware.GetOAuthConfig()

	// Initialize templates
	templateCache, err := createTemplateCache()
	if err != nil {
		return nil, err
	}

	// Create the application
	app := &Application{
		DB:            database,
		DBModel:       &models.DBModel{DB: database},
		TemplateCache: templateCache,
		OAuthConfig:   oauthConfig,
		IsProduction:  isProduction,
	}

	return app, nil
}

// Close closes the application resources
func (app *Application) Close() {
	if app.DB != nil {
		if err := app.DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}
}

// SetupHandlers initializes the handlers
func (app *Application) SetupHandlers() http.Handler {
	// Initialize handlers
	appServices := handlers.NewAppServices(app.DBModel, app.TemplateCache, app.OAuthConfig)
	handlers.NewHandlers(appServices)

	// Create primary mux
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Auth related routes
	mux.HandleFunc("/login", handlers.Services.ShowLoginPage)
	mux.HandleFunc("/auth/google/login", handlers.Services.HandleGoogleLogin)
	mux.HandleFunc("/auth/google/callback", handlers.Services.HandleGoogleCallback)

	// Admin routes with custom handler that checks auth for all admin paths
	mux.HandleFunc("/admin", authenticatedRedirect)
	mux.HandleFunc("/admin/", authenticatedAdmin)

	// Home route MUST be registered LAST to avoid catching other routes
	mux.HandleFunc("/", handlers.Services.Home)

	return mux
}

// runMigrations runs the database migrations using goose
func runMigrations(dbPath string) error {
	// Get the absolute path to the migrations directory
	migrationsDir, err := filepath.Abs("db/migrations")
	if err != nil {
		return err
	}

	// Open the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Run the migrations
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	return goose.Up(db, migrationsDir)
}

// createTemplateCache creates a cache of templates
func createTemplateCache() (map[string]*template.Template, error) {
	funcMap := template.FuncMap{
		"deref": func(f *float64) float64 {
			if f != nil {
				return *f
			}
			return 0
		},
	}

	templates := map[string]*template.Template{
		"index.html":           template.Must(template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html", "templates/category-nav.html")),
		"login.html":           template.Must(template.New("login.html").Funcs(funcMap).ParseFiles("templates/login.html")),
		"admin-dashboard.html": template.Must(template.New("admin-dashboard.html").Funcs(funcMap).ParseFiles("templates/admin-dashboard.html")),
		"menu-form.html":       template.Must(template.New("menu-form.html").Funcs(funcMap).ParseFiles("templates/menu-form.html")),
	}

	return templates, nil
}

// authenticatedRedirect redirects /admin to /admin/
func authenticatedRedirect(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /admin redirect to /admin/")
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

// authenticatedAdmin handles all admin routes with auth protection
func authenticatedAdmin(w http.ResponseWriter, r *http.Request) {
	// Very important debug log
	log.Printf("Admin route requested: %s", r.URL.Path)

	// First check if user is authenticated
	userEmail, valid := middleware.VerifySecureSessionCookie(r)
	if !valid {
		log.Println("Invalid or expired session, redirecting to login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	// Check if the user is authorized
	if !middleware.GetOAuthConfig().IsAllowedEmail(userEmail) {
		log.Printf("Unauthorized access attempt by: %s", userEmail)
		middleware.ClearSecureSessionCookie(w)
		http.Error(w, "Unauthorized: Access denied", http.StatusForbidden)

		return
	}

	// User is authenticated and authorized, handle the route
	path := r.URL.Path

	// Log the exact path for debugging
	log.Printf("Processing authenticated admin path: %s", path)

	// Match the admin route and call the correct handler
	switch {
	case path == "/admin/" || path == "/admin":
		log.Println("Matched admin root path, calling AdminRoot handler")
		handlers.Services.AdminRoot(w, r)

	case path == "/admin/dashboard":
		handlers.Services.AdminDashboard(w, r)

	case path == "/admin/menu/new":
		handlers.Services.ShowCreateMenuItem(w, r)

	case path == "/admin/menu/create":
		handlers.Services.CreateMenuItem(w, r)

	case strings.HasPrefix(path, "/admin/menu/edit/"):
		handlers.Services.ShowEditMenuItem(w, r)

	case strings.HasPrefix(path, "/admin/menu/update/"):
		handlers.Services.UpdateMenuItem(w, r)

	case strings.HasPrefix(path, "/admin/menu/delete/"):
		handlers.Services.DeleteMenuItem(w, r)

	case path == "/admin/flash-message":
		handlers.Services.CreateFlashMessage(w, r)

	case strings.HasPrefix(path, "/admin/flash-message/delete/"):
		handlers.Services.DeleteFlashMessage(w, r)

	case path == "/admin/logout":
		handlers.Services.HandleLogout(w, r)

	default:
		log.Printf("404 Not Found for admin route: %s", path)
		http.NotFound(w, r)
	}
}

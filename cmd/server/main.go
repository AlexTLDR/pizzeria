package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexTLDR/pizzeria/db"
	"github.com/AlexTLDR/pizzeria/internal/handlers"
	"github.com/AlexTLDR/pizzeria/internal/middleware"
	"github.com/AlexTLDR/pizzeria/internal/models"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

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

func main() {
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

	dbPath := "db/pizzeria.db"
	if err := runMigrations(dbPath); err != nil {
		log.Printf("Migration error: %v", err)
	}

	// Initialize OAuth in middleware
	if err := middleware.InitializeOAuth(); err != nil {
		log.Printf("OAuth initialization error: %v", err)
	}

	// Get the OAuth config from middleware for handlers
	oauthConfig := middleware.GetOAuthConfig()

	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

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

	dbModel := models.DBModel{DB: database}

	repo := handlers.NewRepo(&dbModel, templates, oauthConfig)
	handlers.NewHandlers(repo)

	// Create primary mux
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Auth related routes
	mux.HandleFunc("/login", handlers.Repo.ShowLoginPage)
	mux.HandleFunc("/auth/google/login", handlers.Repo.HandleGoogleLogin)
	mux.HandleFunc("/auth/google/callback", handlers.Repo.HandleGoogleCallback)

	// Debug endpoints section removed as tests have been implemented

	// Admin routes with custom handler that checks auth for all admin paths
	mux.HandleFunc("/admin", authenticatedRedirect)
	mux.HandleFunc("/admin/", authenticatedAdmin)

	// Home route MUST be registered LAST to avoid catching other routes
	mux.HandleFunc("/", handlers.Repo.Home)

	// Start the server
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
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
		handlers.Repo.AdminRoot(w, r)

	case path == "/admin/dashboard":
		handlers.Repo.AdminDashboard(w, r)

	case path == "/admin/menu/new":
		handlers.Repo.ShowCreateMenuItem(w, r)

	case path == "/admin/menu/create":
		handlers.Repo.CreateMenuItem(w, r)

	case strings.HasPrefix(path, "/admin/menu/edit/"):
		handlers.Repo.ShowEditMenuItem(w, r)

	case strings.HasPrefix(path, "/admin/menu/update/"):
		handlers.Repo.UpdateMenuItem(w, r)

	case strings.HasPrefix(path, "/admin/menu/delete/"):
		handlers.Repo.DeleteMenuItem(w, r)

	case path == "/admin/flash-message":
		handlers.Repo.CreateFlashMessage(w, r)

	case strings.HasPrefix(path, "/admin/flash-message/delete/"):
		handlers.Repo.DeleteFlashMessage(w, r)

	case path == "/admin/logout":
		handlers.Repo.HandleLogout(w, r)

	default:
		log.Printf("404 Not Found for admin route: %s", path)
		http.NotFound(w, r)
	}
}

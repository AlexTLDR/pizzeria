package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/AlexTLDR/pizzeria/db"
	"github.com/AlexTLDR/pizzeria/internal/handlers"
	"github.com/AlexTLDR/pizzeria/internal/middleware"
	"github.com/AlexTLDR/pizzeria/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

func runMigrations(dbPath string) error {
	log.Println("Running database migrations...")
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	migrationsDir := filepath.Join("db", "migrations")

	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		return err
	}

	if err := goose.Status(sqlDB, migrationsDir); err != nil {
		log.Printf("Warning: Failed to print migration status: %v", err)
	}
	log.Println("Migrations completed successfully!")
	return nil
}

func main() {
	dbPath := "db/pizzeria.db"
	if err := runMigrations(dbPath); err != nil {
		log.Printf("Migration error: %v", err)
	}

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
		"user-management.html": template.Must(template.New("user-management.html").Funcs(funcMap).ParseFiles("templates/user-management.html")),
	}

	dbModel := models.DBModel{DB: database}

	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		log.Printf("Error checking for admin user: %v", err)
	}

	if count == 0 {
		_, err = database.Exec("INSERT INTO users (username, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?)",
			"admin",
			"$2a$12$UxK7WGHr.S1YnJRpMjXmFuQP8jDV8DZTFS1bngCrp2m4h.fMzm1bW",
			time.Now(),
			time.Now(),
		)
		if err != nil {
			log.Printf("Error creating admin user: %v", err)
		} else {
			log.Println("Admin user created successfully")
		}
	}

	repo := handlers.NewRepo(&dbModel, templates)
	handlers.NewHandlers(repo)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", handlers.Repo.Home)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handlers.Repo.Login(w, r)
		} else {
			handlers.Repo.ShowLogin(w, r)
		}
	})

	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/admin/dashboard", handlers.Repo.AdminDashboard)

	// Menu management routes
	adminMux.HandleFunc("/admin/menu/new", handlers.Repo.ShowCreateMenuItem)
	adminMux.HandleFunc("/admin/menu/create", handlers.Repo.CreateMenuItem)
	adminMux.HandleFunc("/admin/menu/edit/", handlers.Repo.ShowEditMenuItem)
	adminMux.HandleFunc("/admin/menu/update/", handlers.Repo.UpdateMenuItem)
	adminMux.HandleFunc("/admin/menu/delete/", handlers.Repo.DeleteMenuItem)

	// Flash message routes
	adminMux.HandleFunc("/admin/flash-message", handlers.Repo.CreateFlashMessage)
	adminMux.HandleFunc("/admin/flash-message/delete/", handlers.Repo.DeleteFlashMessage)

	// User management routes
	adminMux.HandleFunc("/admin/users", handlers.Repo.ShowUserManagement)
	adminMux.HandleFunc("/admin/users/create", handlers.Repo.CreateUser)
	adminMux.HandleFunc("/admin/users/update/", handlers.Repo.UpdateUser)
	adminMux.HandleFunc("/admin/users/delete/", handlers.Repo.DeleteUser)
	adminMux.HandleFunc("/admin/users/change-password", handlers.Repo.ChangePassword)

	// Logout route
	adminMux.HandleFunc("/admin/logout", handlers.Repo.Logout)

	http.Handle("/admin/", middleware.Auth(adminMux))

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

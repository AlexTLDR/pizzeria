package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/AlexTLDR/pizzeria/db"
	"github.com/AlexTLDR/pizzeria/internal/handlers"
	"github.com/AlexTLDR/pizzeria/internal/middleware"
	"github.com/AlexTLDR/pizzeria/internal/models"
)

func main() {
	database, err := db.InitDB("db/pizzeria.db")
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

	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		log.Printf("Error checking for admin user: %v", err)
	}

	if count == 0 {
		_, err = database.Exec(
			"INSERT INTO users (username, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?)",
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

	adminMux.HandleFunc("/admin/menu/new", handlers.Repo.ShowCreateMenuItem)
	adminMux.HandleFunc("/admin/menu/create", handlers.Repo.CreateMenuItem)
	adminMux.HandleFunc("/admin/menu/edit/", handlers.Repo.ShowEditMenuItem)
	adminMux.HandleFunc("/admin/menu/update/", handlers.Repo.UpdateMenuItem)
	adminMux.HandleFunc("/admin/menu/delete/", handlers.Repo.DeleteMenuItem)

	adminMux.HandleFunc("/admin/logout", handlers.Repo.Logout)

	http.Handle("/admin/", middleware.Auth(adminMux))

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

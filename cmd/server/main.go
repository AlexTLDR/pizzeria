package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

type MenuItem struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Category    string
	ImageURL    string
}

func main() {
	// Sample menu items
	pizzaMenu := []MenuItem{
		{ID: "1", Name: "Margherita", Description: "Classic tomato sauce, mozzarella, and basil", Price: 12.99, Category: "Pizza", ImageURL: "/static/images/margherita.jpg"},
		{ID: "2", Name: "Pepperoni", Description: "Tomato sauce, mozzarella, and pepperoni", Price: 14.99, Category: "Pizza", ImageURL: "/static/images/pepperoni.jpg"},
		{ID: "3", Name: "Spaghetti Carbonara", Description: "Creamy sauce with pancetta and parmesan", Price: 13.99, Category: "Pasta", ImageURL: "/static/images/carbonara.jpg"},
	}

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles(
			"templates/header.html",
			"templates/footer.html",
			"templates/index.html",
		))
		// Execute the index.html template specifically
		tmpl.ExecuteTemplate(w, "index.html", map[string]interface{}{
			"Title": "Pizzeria Ristorante - Authentic Italian Cuisine",
			"Year":  time.Now().Year(),
		})
	})

	// Menu page
	http.HandleFunc("/menu", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles(
			"templates/header.html",
			"templates/footer.html",
			"templates/menu.html",
		))
		// Execute the menu.html template specifically
		tmpl.ExecuteTemplate(w, "menu.html", map[string]interface{}{
			"Title": "Our Menu - Pizzeria Ristorante La piccola Sardegna",
			"Menu":  pizzaMenu,
			"Year":  time.Now().Year(),
		})
	})

	// Order page
	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles(
			"templates/header.html",
			"templates/footer.html",
			"templates/order.html",
		))
		// Execute the order.html template specifically
		tmpl.ExecuteTemplate(w, "order.html", map[string]interface{}{
			"Title": "Order Online - Pizzeria Ristorante La piccola Sardegna",
			"Menu":  pizzaMenu,
			"Year":  time.Now().Year(),
		})
	})

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

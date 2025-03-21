package main

import (
	"fmt"
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

// Template function map with the dict function
var funcMap = template.FuncMap{
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
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
		// Create a template with the function map
		tmpl := template.New("").Funcs(funcMap)

		// Parse the template files
		tmpl, err := tmpl.ParseFiles(
			"templates/header.html",
			"templates/footer.html",
			"templates/menu-section.html",
			"templates/index.html",
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Execute the index.html template specifically
		err = tmpl.ExecuteTemplate(w, "index.html", map[string]interface{}{
			"Title": "Pizzeria Ristorante - Authentic Italian Cuisine",
			"Year":  time.Now().Year(),
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Menu page
	http.HandleFunc("/menu", func(w http.ResponseWriter, r *http.Request) {
		// Create a template with the function map
		tmpl := template.New("").Funcs(funcMap)

		// Parse the template files
		tmpl, err := tmpl.ParseFiles(
			"templates/header.html",
			"templates/footer.html",
			"templates/menu.html",
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Execute the menu.html template specifically
		err = tmpl.ExecuteTemplate(w, "menu.html", map[string]interface{}{
			"Title": "Our Menu - Pizzeria Ristorante La piccola Sardegna",
			"Menu":  pizzaMenu,
			"Year":  time.Now().Year(),
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Order page
	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		// Create a template with the function map
		tmpl := template.New("").Funcs(funcMap)

		// Parse the template files
		tmpl, err := tmpl.ParseFiles(
			"templates/header.html",
			"templates/footer.html",
			"templates/order.html",
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Execute the order.html template specifically
		err = tmpl.ExecuteTemplate(w, "order.html", map[string]interface{}{
			"Title": "Order Online - Pizzeria Ristorante La piccola Sardegna",
			"Menu":  pizzaMenu,
			"Year":  time.Now().Year(),
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

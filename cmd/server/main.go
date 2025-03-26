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
	menuItems := []MenuItem{
		// Antipasti
		{ID: "a1", Name: "Antipasto di Mare", Description: "mit Meeresfrüchte", Price: 16.50, Category: "Antipasti", ImageURL: "/static/images/menu/antipasto-di-mare.jpeg"},
		{ID: "a2", Name: "Antipasto della Casa", Description: "Gemischte Vorspeisenplatte mit italienischen Wurstsorten und Caprese", Price: 14.50, Category: "Antipasti", ImageURL: "/static/images/menu/antipasto-della-casa.jpeg"},

		// Insalate
		{ID: "i1", Name: "Caesar Salad", Description: "Romaine lettuce, croutons, parmesan, and Caesar dressing", Price: 9.99, Category: "Insalate", ImageURL: "/static/images/caesar.jpg"},
		{ID: "i2", Name: "Mediterranean Salad", Description: "Mixed greens, olives, feta, and red onion", Price: 11.99, Category: "Insalate", ImageURL: "/static/images/mediterranean.jpg"},

		// Pizza
		{ID: "p1", Name: "Margherita", Description: "Classic tomato sauce, mozzarella, and basil", Price: 12.99, Category: "Pizza", ImageURL: "/static/images/margherita.jpg"},
		{ID: "p2", Name: "Pepperoni", Description: "Tomato sauce, mozzarella, and pepperoni", Price: 14.99, Category: "Pizza", ImageURL: "/static/images/pepperoni.jpg"},
		{ID: "p3", Name: "Quattro Formaggi", Description: "Four cheese pizza with mozzarella, gorgonzola, fontina, and parmesan", Price: 15.99, Category: "Pizza", ImageURL: "/static/images/quattro.jpg"},

		// Spaghetti
		{ID: "s1", Name: "Spaghetti Carbonara", Description: "Creamy sauce with pancetta and parmesan", Price: 13.99, Category: "Spaghetti", ImageURL: "/static/images/carbonara.jpg"},
		{ID: "s2", Name: "Spaghetti Bolognese", Description: "Rich meat sauce with ground beef and tomatoes", Price: 14.99, Category: "Spaghetti", ImageURL: "/static/images/bolognese.jpg"},

		// Penne
		{ID: "pe1", Name: "Penne Arrabbiata", Description: "Spicy tomato sauce with garlic and chili", Price: 12.99, Category: "Penne", ImageURL: "/static/images/arrabbiata.jpg"},
		{ID: "pe2", Name: "Penne alla Vodka", Description: "Creamy tomato sauce with a splash of vodka", Price: 14.99, Category: "Penne", ImageURL: "/static/images/vodka.jpg"},

		// Rigatoni
		{ID: "r1", Name: "Rigatoni al Forno", Description: "Baked rigatoni with meat sauce and cheese", Price: 15.99, Category: "Rigatoni", ImageURL: "/static/images/rigatoni.jpg"},

		// Pasta al Forno
		{ID: "pf1", Name: "Lasagna", Description: "Layers of pasta, meat sauce, and cheese", Price: 16.99, Category: "Pasta al Forno", ImageURL: "/static/images/lasagna.jpg"},
		{ID: "pf2", Name: "Cannelloni", Description: "Pasta tubes filled with ricotta and spinach", Price: 15.99, Category: "Pasta al Forno", ImageURL: "/static/images/cannelloni.jpg"},

		// Pesce Fritto
		{ID: "pesc1", Name: "Calamari Fritti", Description: "Fried squid rings with marinara sauce", Price: 13.99, Category: "Pesce Fritto", ImageURL: "/static/images/calamari.jpg"},

		// Carne
		{ID: "c1", Name: "Chicken Parmigiana", Description: "Breaded chicken with tomato sauce and melted cheese", Price: 17.99, Category: "Carne", ImageURL: "/static/images/parmigiana.jpg"},
		{ID: "c2", Name: "Veal Scaloppine", Description: "Thin slices of veal with lemon and capers", Price: 19.99, Category: "Carne", ImageURL: "/static/images/scaloppine.jpg"},
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.New("").Funcs(funcMap)

		tmpl, err := tmpl.ParseFiles(
			"templates/header.html",
			"templates/footer.html",
			"templates/category-nav.html",
			"templates/index.html",
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.ExecuteTemplate(w, "index.html", map[string]interface{}{
			"Title": "Pizzeria Ristorante - Authentic Italian Cuisine",
			"Menu":  menuItems,
			"Year":  time.Now().Year(),
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

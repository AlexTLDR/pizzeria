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
		{ID: "a3", Name: "Tomatensuppe", Description: "", Price: 7.60, Category: "Antipasti", ImageURL: "/static/images/menu/tomatensuppe.jpeg"},
		{ID: "a4", Name: "Bruscheta", Description: "Selbst gebakenes Brot mit frischen Tomaten und Knoblauch", Price: 8.50, Category: "Antipasti", ImageURL: "/static/images/menu/bruscheta.jpeg"},
		{ID: "a5", Name: "Caprese", Description: "Tomaten, Mozzarella mit Panini", Price: 9.50, Category: "Antipasti", ImageURL: "/static/images/menu/caprese.jpeg"},

		// Insalate
		{ID: "i1", Name: "Gemischter Salat", Description: "", Price: 4.90, Category: "Insalate", ImageURL: "/static/images/menu/gemischter-salat.jpeg"},
		{ID: "i2", Name: "Salat Tonno", Description: "Gemischter Salat mit Thunfisch, Panini", Price: 10.40, Category: "Insalate", ImageURL: "/static/images/menu/salat-tonno.jpeg"},
		{ID: "i1", Name: "Salat Capricciosa", Description: "Gemischter Salat mit Mozzarella, Landschinken, Panini", Price: 10.80, Category: "Insalate", ImageURL: "/static/images/menu/salat-capricciosa.jpeg"},
		{ID: "i1", Name: "Salat della Casa", Description: "Gemischter Salat mit Putenstreifen, panini", Price: 12.80, Category: "Insalate", ImageURL: "/static/images/menu/salat-della-casa.jpeg"},
		{ID: "i1", Name: "Salat Marinara", Description: "Gemischter Salat mit Meeresfrüchte, Panini", Price: 12.80, Category: "Insalate", ImageURL: "/static/images/menu/salat-marinara.jpeg"},

		// Pizza
		{ID: "p1", Name: "Pizzabrot", Description: "", Price: 12.99, Category: "Pizza", ImageURL: "/static/images/menu/pizzabrot.jpeg"},
		{ID: "p2", Name: "Pomodoro", Description: "Tomaten, Käse", Price: 9.60, Category: "Pizza", ImageURL: "/static/images/menu/pomodoro.jpeg"},
		{ID: "p3", Name: "Proscuitto", Description: "Vorderschinken", Price: 15.99, Category: "Pizza", ImageURL: "/static/images/menu/proscuitto.jpeg"},

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

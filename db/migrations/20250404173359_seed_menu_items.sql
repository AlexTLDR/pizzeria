-- +goose Up
-- SQL in this section is executed when the migration is applied.
INSERT INTO menu_items (name, description, price, small_price, category, image_url) VALUES
-- Antipasti
('Antipasto di Mare', 'mit Meeresfrüchte', 16.50, NULL, 'Antipasti', '/static/images/menu/antipasto-di-mare.jpeg'),
('Antipasto della Casa', 'Gemischte Vorspeisenplatte mit italienischen Wurstsorten und Caprese', 14.50, NULL, 'Antipasti', '/static/images/menu/antipasto-della-casa.jpeg'),
('Tomatensuppe', '', 7.60, NULL, 'Antipasti', '/static/images/menu/tomatensuppe.jpeg'),
('Bruscheta', 'Selbst gebakenes Brot mit frischen Tomaten und Knoblauch', 8.50, NULL, 'Antipasti', '/static/images/menu/bruscheta.jpeg'),
('Caprese', 'Tomaten, Mozzarella mit Panini', 9.50, NULL, 'Antipasti', '/static/images/menu/caprese.jpeg'),
-- Insalate
('Gemischter Salat', '', 4.90, NULL, 'Insalate', '/static/images/menu/gemischter-salat.jpeg'),
('Salat Tonno', 'Gemischter Salat mit Thunfisch, Panini', 10.40, NULL, 'Insalate', '/static/images/menu/salat-tonno.jpeg'),
('Salat Capricciosa', 'Gemischter Salat mit Mozzarella, Landschinken, Panini', 10.80, NULL, 'Insalate', '/static/images/menu/salat-capricciosa.jpeg'),
('Salat della Casa', 'Gemischter Salat mit Putenstreifen, panini', 12.80, NULL, 'Insalate', '/static/images/menu/salat-della-casa.jpeg'),
('Salat Marinara', 'Gemischter Salat mit Meeresfrüchte, Panini', 12.80, NULL, 'Insalate', '/static/images/menu/salat-marinara.jpeg'),
-- Pizza
('Pizzabrot', '', 12.99, NULL, 'Pizza', '/static/images/menu/pizzabrot.jpeg'),
('Pomodoro', 'Tomaten, Käse', 9.60, 4.60, 'Pizza', '/static/images/menu/pomodoro.jpeg'),
('Proscuitto', 'Vorderschinken', 15.99, 4.60, 'Pizza', '/static/images/menu/proscuitto.jpeg'),
-- Spaghetti
('Spaghetti Carbonara', 'Creamy sauce with pancetta and parmesan', 13.99, NULL, 'Spaghetti', '/static/images/carbonara.jpg'),
('Spaghetti Bolognese', 'Rich meat sauce with ground beef and tomatoes', 14.99, NULL, 'Spaghetti', '/static/images/bolognese.jpg'),
-- Penne
('Penne Arrabbiata', 'Spicy tomato sauce with garlic and chili', 12.99, NULL, 'Penne', '/static/images/arrabbiata.jpg'),
('Penne alla Vodka', 'Creamy tomato sauce with a splash of vodka', 14.99, NULL, 'Penne', '/static/images/vodka.jpg'),
-- Rigatoni
('Rigatoni al Forno', 'Baked rigatoni with meat sauce and cheese', 15.99, NULL, 'Rigatoni', '/static/images/rigatoni.jpg'),
-- Pasta al Forno
('Lasagna', 'Layers of pasta, meat sauce, and cheese', 16.99, NULL, 'Pasta al Forno', '/static/images/lasagna.jpg'),
('Cannelloni', 'Pasta tubes filled with ricotta and spinach', 15.99, NULL, 'Pasta al Forno', '/static/images/cannelloni.jpg'),
-- Pesce Fritto
('Calamari Fritti', 'Fried squid rings with marinara sauce', 13.99, NULL, 'Pesce Fritto', '/static/images/calamari.jpg'),
-- Carne
('Chicken Parmigiana', 'Breaded chicken with tomato sauce and melted cheese', 17.99, NULL, 'Carne', '/static/images/parmigiana.jpg'),
('Veal Scaloppine', 'Thin slices of veal with lemon and capers', 19.99, NULL, 'Carne', '/static/images/scaloppine.jpg');

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DELETE FROM menu_items WHERE name IN (
    'Antipasto di Mare', 'Antipasto della Casa', 'Tomatensuppe', 'Bruscheta', 'Caprese',
    'Gemischter Salat', 'Salat Tonno', 'Salat Capricciosa', 'Salat della Casa', 'Salat Marinara',
    'Pizzabrot', 'Pomodoro', 'Proscuitto',
    'Spaghetti Carbonara', 'Spaghetti Bolognese',
    'Penne Arrabbiata', 'Penne alla Vodka',
    'Rigatoni al Forno',
    'Lasagna', 'Cannelloni',
    'Calamari Fritti',
    'Chicken Parmigiana', 'Veal Scaloppine'
);

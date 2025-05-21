package models

import (
	"context"
	"time"
)

// MenuItem represents a menu item in the database
type MenuItem struct {
	ID          int
	Name        string
	Description string
	Price       float64
	SmallPrice  *float64
	Category    string
	ImageURL    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GetAllMenuItems retrieves all menu items from the database
func (m *DBModel) GetAllMenuItems() ([]MenuItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, name, description, price, small_price, category, image_url, 
              created_at, updated_at FROM menu_items ORDER BY category, name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MenuItem

	for rows.Next() {
		var item MenuItem
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.SmallPrice,
			&item.Category,
			&item.ImageURL,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	// Check for errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// GetMenuItemByID retrieves a menu item by its ID
func (m *DBModel) GetMenuItemByID(id int) (MenuItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, name, description, price, small_price, category, image_url, 
              created_at, updated_at FROM menu_items WHERE id = ?`

	var item MenuItem

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.SmallPrice,
		&item.Category,
		&item.ImageURL,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return item, err
	}

	return item, nil
}

// InsertMenuItem inserts a new menu item into the database
func (m *DBModel) InsertMenuItem(item MenuItem) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO menu_items (name, description, price, small_price, category, image_url, created_at, updated_at)
             VALUES (?, ?, ?, ?, ?, ?, ?, ?)
             RETURNING id`

	var newID int
	err := m.DB.QueryRowContext(ctx, stmt,
		item.Name,
		item.Description,
		item.Price,
		item.SmallPrice,
		item.Category,
		item.ImageURL,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// UpdateMenuItem updates an existing menu item in the database
func (m *DBModel) UpdateMenuItem(item MenuItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE menu_items SET
             name = ?,
             description = ?,
             price = ?,
             small_price = ?,
             category = ?,
             image_url = ?,
             updated_at = ?
             WHERE id = ?`

	_, err := m.DB.ExecContext(ctx, stmt,
		item.Name,
		item.Description,
		item.Price,
		item.SmallPrice,
		item.Category,
		item.ImageURL,
		time.Now(),
		item.ID,
	)

	return err
}

// DeleteMenuItem deletes a menu item from the database
func (m *DBModel) DeleteMenuItem(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM menu_items WHERE id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, id)

	return err
}

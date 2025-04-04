package models

import (
	"context"
	"database/sql"
	"time"
)

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

type User struct {
	ID           int
	Username     string
	PasswordHash string
	//IsAdmin      bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DBModel struct {
	DB *sql.DB
}

// Menu Item Methods
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

	return items, nil
}

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

func (m *DBModel) DeleteMenuItem(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM menu_items WHERE id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, id)
	return err
}

// User Methods
// User Methods
func (m *DBModel) GetUserByUsername(username string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, username, password_hash, created_at, updated_at 
             FROM users WHERE username = ?`

	var user User
	row := m.DB.QueryRowContext(ctx, query, username)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (m *DBModel) InsertUser(user User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO users (username, password_hash, created_at, updated_at)
            VALUES (?, ?, ?, ?)
            RETURNING id`

	var newID int
	err := m.DB.QueryRowContext(ctx, stmt,
		user.Username,
		user.PasswordHash,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

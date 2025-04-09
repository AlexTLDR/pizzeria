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

// User struct - kept for database migration compatibility
// This is being deprecated as we move to Google OAuth authentication
type User struct {
	ID           int
	Username     string
	PasswordHash string
	Email        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type FlashMessage struct {
	ID        int
	Type      string // "success" or "error"
	Message   string
	StartDate time.Time
	EndDate   time.Time
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Status    string
}

// GetStatus returns the appropriate status for the flash message based on date range
func (f *FlashMessage) GetStatus() string {
	if !f.Active {
		return "Inactive"
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Extract date components for comparison
	startDate := time.Date(f.StartDate.Year(), f.StartDate.Month(), f.StartDate.Day(), 0, 0, 0, 0, f.StartDate.Location())
	endDate := time.Date(f.EndDate.Year(), f.EndDate.Month(), f.EndDate.Day(), 0, 0, 0, 0, f.EndDate.Location())

	// Check if the dates are equal (comparing year, month, day only)
	startEqual := today.Year() == startDate.Year() && today.Month() == startDate.Month() && today.Day() == startDate.Day()
	endEqual := today.Year() == endDate.Year() && today.Month() == endDate.Month() && today.Day() == endDate.Day()

	// If today comes exactly on start date or end date, it's active
	if startEqual || endEqual {
		return "Active"
	}

	// If today is before the start date
	if today.Before(startDate) {
		return "Scheduled"
	} else if today.After(endDate) {
		return "Expired"
	}
	return "Active"
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

// DEPRECATED: User Methods - No longer used with Google OAuth authentication
// These methods are kept for reference but are no longer used in the application
/*
func (m *DBModel) GetUserByUsername(username string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, username, password_hash, email, created_at, updated_at
               FROM users WHERE username = ?`

	var user User
	row := m.DB.QueryRowContext(ctx, query, username)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
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

	stmt := `INSERT INTO users (username, password_hash, email, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?)
            RETURNING id`

	var newID int
	err := m.DB.QueryRowContext(ctx, stmt,
		user.Username,
		user.PasswordHash,
		user.Email,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}
	return newID, nil
}
*/

// Commented out user methods that are no longer used
/*
// Get all users
func (m *DBModel) GetAllUsers() ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, username, password_hash, email, created_at, updated_at
              FROM users ORDER BY username`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// Get user by ID
func (m *DBModel) GetUserByID(id int) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, username, password_hash, email, created_at, updated_at
              FROM users WHERE id = ?`

	var user User
	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return user, err
	}

	return user, nil
}

// Update user information
func (m *DBModel) UpdateUser(user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE users SET
             username = ?,
             password_hash = ?,
             email = ?,
             updated_at = ?
             WHERE id = ?`

	_, err := m.DB.ExecContext(ctx, stmt,
		user.Username,
		user.PasswordHash,
		user.Email,
		time.Now(),
		user.ID,
	)

	return err
}

// Delete user
func (m *DBModel) DeleteUser(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM users WHERE id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, id)
	return err
}

// Count total number of users
func (m *DBModel) GetUserCount() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	query := `SELECT COUNT(*) FROM users`
	err := m.DB.QueryRowContext(ctx, query).Scan(&count)

	return count, err
}
*/

// FlashMessage Methods - These are still used and active
func (m *DBModel) CreateFlashMessage(message FlashMessage) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO flash_messages (type, message, start_date, end_date, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id`

	var newID int
	err := m.DB.QueryRowContext(ctx, stmt,
		message.Type,
		message.Message,
		message.StartDate,
		message.EndDate,
		message.Active,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *DBModel) GetActiveFlashMessages() ([]FlashMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Get messages that are active AND the current date is between start_date and end_date
	currentDate := time.Now().Format("2006-01-02")
	query := `SELECT id, type, message, start_date, end_date, active, created_at, updated_at
		FROM flash_messages
		WHERE active = 1
		AND ? BETWEEN date(start_date) AND date(end_date)
		ORDER BY created_at DESC`

	rows, err := m.DB.QueryContext(ctx, query, currentDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []FlashMessage

	for rows.Next() {
		var msg FlashMessage
		err := rows.Scan(
			&msg.ID,
			&msg.Type,
			&msg.Message,
			&msg.StartDate,
			&msg.EndDate,
			&msg.Active,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Set the status based on date range
		msg.Status = msg.GetStatus()

		messages = append(messages, msg)
	}

	return messages, nil
}

func (m *DBModel) GetAllFlashMessages() ([]FlashMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, type, message, start_date, end_date, active, created_at, updated_at
		FROM flash_messages
		ORDER BY created_at DESC`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []FlashMessage

	for rows.Next() {
		var msg FlashMessage
		err := rows.Scan(
			&msg.ID,
			&msg.Type,
			&msg.Message,
			&msg.StartDate,
			&msg.EndDate,
			&msg.Active,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Set the status based on date range
		msg.Status = msg.GetStatus()

		messages = append(messages, msg)
	}

	return messages, nil
}

func (m *DBModel) DeleteFlashMessage(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM flash_messages WHERE id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, id)
	return err
}

package models

import (
	"context"
	"time"
)

// FlashMessage represents a flash message in the database
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

// CreateFlashMessage creates a new flash message in the database
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

// GetActiveFlashMessages retrieves all active flash messages from the database
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

// GetAllFlashMessages retrieves all flash messages from the database
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

// DeleteFlashMessage deletes a flash message from the database
func (m *DBModel) DeleteFlashMessage(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM flash_messages WHERE id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, id)
	return err
}
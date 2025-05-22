package models

import (
	"database/sql"
	"time"
)

// DBModel wraps the database connection
type DBModel struct {
	DB *sql.DB
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

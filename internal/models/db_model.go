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

// Note: All the methods for MenuItem and FlashMessage have been moved to their respective files.
// The User struct is kept here for reference but its methods are commented out in the original file
// and are not included here.

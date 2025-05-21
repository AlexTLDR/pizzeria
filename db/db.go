package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Initialize database connection
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Ping the database to ensure the connection is valid
	err = db.Ping()
	if err != nil {
		closeErr := db.Close() // Close the connection if ping fails
		if closeErr != nil {
			return nil, fmt.Errorf("ping error: %v, close error: %v", err, closeErr)
		}
		return nil, err
	}

	return db, nil
}

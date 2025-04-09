-- +goose Up
-- Drop the users table since we're moving to Google OAuth for authentication
DROP TABLE IF EXISTS users;

-- +goose Down
-- Recreate the users table if we need to roll back
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert the default admin user
INSERT INTO users (username, email, password) 
VALUES ('admin', 'admin@example.com', '$2a$12$Fm71kVRs9PgZrbAodRnJjeQw.VfIGCXQKJ97NtrW8a3GqQlzJ5nci');
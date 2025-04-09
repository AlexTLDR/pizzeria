-- +goose Up
-- SQL in this section is executed when the migration is applied.
-- For our transition to Google Auth, we're keeping the users table for now
-- but adding a new column to store authenticated email addresses

-- Add email column to users table if it doesn't exist
ALTER TABLE users ADD COLUMN email TEXT DEFAULT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE users DROP COLUMN email;
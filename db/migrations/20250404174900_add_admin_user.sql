-- +goose Up
-- SQL in this section is executed when the migration is applied.
INSERT INTO users (username, password_hash, created_at, updated_at)
SELECT 'admin',
       -- Hashed password for 'admin123' using bcrypt
       '$2a$12$UxK7WGHr.S1YnJRpMjXmFuQP8jDV8DZTFS1bngCrp2m4h.fMzm1bW',
       CURRENT_TIMESTAMP,
       CURRENT_TIMESTAMP
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE username = 'admin'
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DELETE FROM users WHERE username = 'admin';
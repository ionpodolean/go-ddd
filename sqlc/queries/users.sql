-- name: CreateUser :execresult
INSERT INTO users (email, password, first_name, last_name, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetUserByEmail :one
SELECT id, email, password, first_name, last_name, created_at, updated_at
FROM users
WHERE email = ? LIMIT 1;

-- name: GetUserByID :one
SELECT id, email, password, first_name, last_name, created_at, updated_at
FROM users
WHERE id = ? LIMIT 1;

-- name: ExistsUserByEmail :one
SELECT COUNT(1)
FROM users
WHERE email = ?;

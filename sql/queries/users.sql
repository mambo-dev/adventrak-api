-- name: CreateUser :one
INSERT INTO users (id, username, email, password, updated_at) 
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) 
RETURNING *;
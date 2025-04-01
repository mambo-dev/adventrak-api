-- name: CreateUser :one
INSERT INTO users (username, email, password, updated_at) 
VALUES (
    $1,
    $2,
    $3,
    $4
) 
RETURNING *;
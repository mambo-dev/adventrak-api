-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, updated_at) 
VALUES (
    $1,
    $2,
    $3,
    $4
) 
RETURNING *;

-- name: GetUser :one

SELECT id, username, email,password_hash, created_at
FROM USERS
WHERE username = $1;

-- name: UpdateUserDetails :one
UPDATE users
SET username = $1, email = $2, updated_at = $3
WHERE id = $1 
RETURNING *;

-- name: UpdatePassword :exec
UPDATE users
SET password_hash = $1
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
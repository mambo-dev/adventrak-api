-- name: CreateRefreshToken :one
INSERT INTO refresh_token(token,expires_at, user_id)
VALUES(
    $1,
    $2,
    $3 
)
RETURNING *;

-- name: RevokeRefreshToken :exec
UPDATE refresh_token
SET revoked_at = NOW()
WHERE token = $1;

-- name: GetRefreshToken :one 
SELECT * FROM refresh_token
WHERE user_id = $1;
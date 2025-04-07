-- name: CreateAccount :exec
INSERT INTO account (user_id)
VALUES (
    $1
);

-- name: VerifyAccount :exec
UPDATE account 
SET verified = true
WHERE user_id = $1;

-- name: DisableAccount :exec
UPDATE account
SET verified = NOW()
WHERE user_id = $1;

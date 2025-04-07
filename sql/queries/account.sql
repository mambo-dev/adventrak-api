-- name: CreateAccount :exec
INSERT INTO account (user_id)
VALUES (
    $1
);

-- name: GetUserAccount :one
SELECT * FROM account
WHERE user_id = $1;

-- name: SetVerificationCode :exec
UPDATE account 
SET verification_code = $1, verification_expires_at = $2
WHERE user_id = $3;


-- name: VerifyAccount :exec
UPDATE account 
SET verified = true, verification_code = NULL, verification_expires_at =  NOW()
WHERE user_id = $1;

-- name: DisableAccount :exec
UPDATE account
SET verified = NOW()
WHERE user_id = $1;


-- name: SetResetCode :exec
UPDATE account 
SET reset_code = $1, reset_code_expires_at = $2
WHERE user_id  = $3;



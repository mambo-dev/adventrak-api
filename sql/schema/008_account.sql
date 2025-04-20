-- +goose Up
ALTER TABLE account
ADD verification_code VARCHAR NOT NULL DEFAULT '';

ALTER TABLE account
ADD verification_expires_at TIMESTAMP ;

ALTER TABLE account
ADD reset_code_expires_at TIMESTAMP ;

ALTER TABLE account
ALTER COLUMN reset_code DROP NOT NULL;

-- +goose Down
ALTER TABLE account
DROP COLUMN verification_code;

ALTER TABLE account
DROP COLUMN verification_expires_at;

ALTER TABLE account
DROP COLUMN reset_code_expires_at;


ALTER TABLE account
ALTER COLUMN reset_code TYPE VARCHAR,
ALTER COLUMN reset_code SET NOT NULL,
ALTER COLUMN reset_code SET DEFAULT '';

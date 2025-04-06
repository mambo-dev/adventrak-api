-- +goose Up
CREATE TABLE account (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    verified BOOLEAN NOT NULL DEFAULT true,
    reset_code VARCHAR NOT NULL DEFAULT '',
    disabled_at TIMESTAMP,
    user_id uuid ,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE account;
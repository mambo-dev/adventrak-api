-- +goose Up

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    username VARCHAR UNIQUE NOT NULL,
    password VARCHAR NOT NULL,
    email  VARCHAR UNIQUE  NOT NULL
);

-- +goose Down
DROP TABLE users;
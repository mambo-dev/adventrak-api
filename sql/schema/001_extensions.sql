-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;


-- +goose Down
DROP EXTENSION  "uuid-ossp";
DROP EXTENSION  postgis;
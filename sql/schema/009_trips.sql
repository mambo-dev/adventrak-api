-- +goose Up
ALTER TABLE trips
ALTER COLUMN end_location DROP NOT NULL;

-- +goose Down
ALTER TABLE trips
ALTER COLUMN   end_location GEOGRAPHY(POINT, 4326) NOT NULL,
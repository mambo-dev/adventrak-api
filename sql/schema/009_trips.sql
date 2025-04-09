-- +goose Up
ALTER TABLE trips
ADD start_location_name VARCHAR(50) NOT NULL DEFAULT '';

ALTER TABLE trips
ADD end_location_name VARCHAR(50) ;

ALTER TABLE trips
ALTER COLUMN end_location DROP NOT NULL;

-- +goose Down
ALTER TABLE trips
ALTER COLUMN end_location TYPE geography(Point, 4326)
USING end_location::geography;

ALTER TABLE trips
DROP start_location_name;

ALTER TABLE trips
DROP end_location_name;
-- +goose Up
CREATE TABLE trip_stop(
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        trip_id uuid UNIQUE NOT NULL,
        FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE CASCADE,
        location_name VARCHAR NOT NULL,
        location_tag GEOGRAPHY(POINT, 4326) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(), 
        updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE trip_stop;
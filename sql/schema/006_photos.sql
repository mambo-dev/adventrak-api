-- +goose Up
CREATE TABLE trip_photos(
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        trip_id uuid UNIQUE ,
        FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE CASCADE,
        trip_stop_id uuid UNIQUE ,
        FOREIGN KEY (trip_stop_id) REFERENCES trip_stop(id) ON DELETE CASCADE,
        photo_url VARCHAR NOT NULL ,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(), 
        updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE trip_photos;
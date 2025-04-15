-- +goose Up
ALTER TABLE trip_stop
DROP CONSTRAINT trip_stop_trip_id_key;

-- +goose Down
ALTER TABLE trip_stop
ADD CONSTRAINT trip_stop_trip_id_key UNIQUE (trip_id);

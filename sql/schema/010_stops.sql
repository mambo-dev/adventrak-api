-- +goose Up

ALTER TABLE trip_stop
ADD user_id uuid  NOT NULL;

ALTER TABLE trip_stop
ADD CONSTRAINT fk_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;


-- +goose Down
ALTER TABLE trip_stop
DROP user_id;
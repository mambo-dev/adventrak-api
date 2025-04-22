-- +goose Up

ALTER TABLE trip_media
ADD user_id uuid  NOT NULL;

ALTER TABLE trip_media
ADD CONSTRAINT fk_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;



-- +goose Down
ALTER TABLE trip_media
DROP user_id;
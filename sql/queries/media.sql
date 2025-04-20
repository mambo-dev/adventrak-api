-- name: CreateTripMedia :one
INSERT INTO trip_media(trip_id, trip_stop_id, photo_url, video_url)
VALUES(
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: UpdateTripMedia :exec
UPDATE trip_media
SET photo_url = $1, video_url = $2, updated_at = NOW()
WHERE id = $3;

-- name: DeleteTripMedia :exec
DELETE FROM trip_media
WHERE id = $1;

-- name: GetTripMediaById :one
SELECT * FROM trip_media
WHERE id = $1;

-- name: GetTripMediaByTripId :many
SELECT * FROM trip_media
WHERE trip_id = $1;
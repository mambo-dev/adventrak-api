-- name: CreateTripMedia :one
INSERT INTO trip_media(trip_id, trip_stop_id, photo_url, video_url)
VALUES(
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: UpdateTripMedia :one
UPDATE trip_media
SET photo_url = $1, video_url = $2, updated_at = NOW()
WHERE id = $3
RETURNING id;

-- name: DeleteTripMedia :exec
DELETE FROM trip_media
WHERE id = $1;

-- name: GetTripMediaById :one
SELECT * FROM trip_media
WHERE id = $1;

-- name: GetTripMediaByTripOrStopID :many
SELECT * FROM trip_media
WHERE trip_id = $1 OR trip_stop_id = $2;


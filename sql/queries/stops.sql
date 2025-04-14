-- name: GetStops :many
SELECT
    id,
    location_name,
    created_at,
    ST_Y(location_tag::geometry) AS end_lat,
    ST_X(location_tag::geometry) AS end_lng
FROM trip_stop
WHERE trip_id = $1 AND user_id = $2;

-- name: GetStop :one
SELECT
    id,
    location_name,
    created_at,
    ST_Y(location_tag::geometry) AS end_lat,
    ST_X(location_tag::geometry) AS end_lng
FROM trip_stop
WHERE user_id = $1 AND id = $2;


-- name: CreateStop :one
INSERT INTO trip_stop (
    location_name,
    location_tag,
    trip_id
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING id;

-- name: UpdateStop :one
UPDATE trip_stop
SET location_name = $1, location_tag= $2
WHERE id = $3 AND user_id = $4
RETURNING id;

-- name: DeleteStop :exec
DELETE FROM trip_stop 
WHERE id = $1 AND user_id = $2; 


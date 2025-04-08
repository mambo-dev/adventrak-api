-- name: CreateTrip :one
INSERT INTO trips (
    start_date,
    start_location,
    end_location,
    end_date,
    distance_travelled,
    user_id
) VALUES (
$1,
$2,
$3,
$4,
$5,
$6
)
RETURNING *;



-- name: UpdateTrip :one
UPDATE trips
SET
    end_location = $1,
    end_date = $2,
    distance_travelled = $3,
    updated_at = $4
WHERE
    id = $5 
RETURNING *;

-- name: DeleteTrip :exec
DELETE FROM trips
WHERE  id = $1;


-- name: GetTrips :many
SELECT * FROM trips 
WHERE user_id = $1;

-- name: GetTrip :one
SELECT * FROM trips 
WHERE user_id = $1 AND  id = $2; 
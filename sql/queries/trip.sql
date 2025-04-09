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
RETURNING  id;



-- name: UpdateTrip :one
UPDATE trips
SET
    end_location = $1,
    end_date = $2,
    distance_travelled = $3,
    updated_at = $4
WHERE
    id = $5 
RETURNING  id;


-- name: DeleteTrip :exec
DELETE FROM trips
WHERE  id = $1;


-- name: GetTrips :many
SELECT   
  id,
  start_date,
  end_date,
  distance_travelled,
  created_at,
  updated_at,
  user_id,
  ST_Y(start_location::geometry) AS start_lat,
  ST_X(start_location::geometry) AS start_lng,
  ST_Y(end_location::geometry) AS end_lat,
  ST_X(end_location::geometry) AS end_lng
FROM trips 
WHERE user_id = $1;


-- name: GetTrip :one
SELECT 
  id,
  start_date,
  end_date,
  distance_travelled,
  created_at,
  updated_at,
  user_id,
  ST_Y(start_location::geometry) AS start_lat,
  ST_X(start_location::geometry) AS start_lng,
  ST_Y(end_location::geometry) AS end_lat,
  ST_X(end_location::geometry) AS end_lng
FROM trips
WHERE user_id = $1 AND id = $2;


-- name: GetTripDistance :one
SELECT ST_Distance(start_location, end_location) AS distance 
FROM trips WHERE id = $1;
-- name: CreateReservation :one
INSERT INTO reservations (user_id, room_id, start_time, end_time, status)
VALUES (
	$1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetReservation :one
SELECT * FROM reservations
WHERE id = $1 LIMIT 1;

-- name: ListReservationsByRoom :many
SELECT * FROM reservations
WHERE room_id = $1
ORDER BY start_time ASC;

-- name: ExistsOverlappingReservation :one
SELECT EXISTS (
    SELECT 1
    FROM reservations
    WHERE room_id = $1
      AND start_time < $3
      AND end_time > $2
);
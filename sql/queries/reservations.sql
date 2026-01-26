-- name: CreateReservation :one
INSERT INTO reservations (user_id, room_id, start_time, end_time, status)
VALUES (
	$1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetReservationByID :one
SELECT * FROM reservations
WHERE id = $1;

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

-- name: GetAllBetweenDates :many
SELECT 
    r.id,
    r.room_id,
    r.start_time,
    r.end_time,
    r.user_id as created_by_id,
    u.name as created_by_name,
    room.name as room_name
FROM reservations r
INNER JOIN users u ON r.user_id = u.id
INNER JOIN rooms room ON r.room_id = room.id
WHERE r.start_time >= $1
  AND r.end_time <= $2
ORDER BY r.room_id, r.start_time;

-- name: DeleteReservation :exec
DELETE FROM reservations
WHERE id = $1;
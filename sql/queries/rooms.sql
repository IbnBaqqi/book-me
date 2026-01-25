-- name: GetRoomByID :one
SELECT * FROM rooms
WHERE id = $1;
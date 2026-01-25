-- name: CreateUser :one
INSERT INTO users (email, name, role)
VALUES (
	$1, $2, $3
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- -- name: UpdateUser :one
-- UPDATE users
-- SET email = $2, name = $3, role = $4
-- WHERE id = $1
-- RETURNING *;

-- -- name: ListUsers :many
-- SELECT * FROM users
-- ORDER BY name;

-- -- name: DeleteUser :exec
-- DELETE FROM users
-- WHERE id = $1;
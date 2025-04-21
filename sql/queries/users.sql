-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET updated_at = NOW(),
    email = $2,
    hashed_password = $3
WHERE id = $1
RETURNING *;

-- name: UpgradeUser :one
-- This query is used to upgrade a user to a "chirpy red" status.
UPDATE users
SET updated_at = NOW(),
    is_chirpy_red = TRUE
WHERE id = $1
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: DeleteUsers :exec
DELETE FROM users;
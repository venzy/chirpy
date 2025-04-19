-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
    $1,         -- token
    NOW(),      -- created_at
    NOW(),      -- updated_at
    $2,         -- user_id
    $3          -- expires_at
)
RETURNING *;

-- name: GetUserIDWithRefreshToken :one
SELECT 
    users.id AS user_id,
    refresh_tokens.expires_at,
    refresh_tokens.revoked_at
FROM refresh_tokens
JOIN users ON refresh_tokens.user_id = users.id
WHERE refresh_tokens.token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE token = $1 AND revoked_at IS NULL;

-- name: DeleteRefreshTokenByToken :exec
DELETE FROM refresh_tokens WHERE token = $1;
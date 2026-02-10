-- name: CreateToken :one
INSERT INTO refresh_tokens (id, created_at, updated_at, user_id, expires_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3
)
RETURNING *;

-- name: GetToken :one
SELECT * FROM refresh_tokens WHERE id = $1;

-- name: MarkTokenRevoked :exec
UPDATE refresh_tokens
SET
    updated_at = $1,
    revoked_at = $1
WHERE id = $2;

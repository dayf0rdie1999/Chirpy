-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, revoked_at, user_id)
VALUES (
    $1,
    NOW(),
    NOW(),
    NOW() + interval '60 days',
    NULL,
    $2
)
RETURNING *;

-- name: RevokeRefreshToken :exec
UPDATE
    refresh_tokens
SET
    revoked_at = NOW(),
    updated_at = NOW()
WHERE
    refresh_tokens.token = $1;
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

-- name: DeleteUsers :exec
DELETE FROM users;


-- name: GetUserByEmail :one
SELECT
    *
FROM
    users
WHERE
    email = $1;

-- name: GetUserFromRefreshToken :one
SELECT
    *
FROM
    users
INNER JOIN refresh_tokens ON refresh_tokens.user_id = users.id
WHERE
    refresh_tokens.token = $1
    AND revoked_at IS NULL
    AND expires_at > NOW();

-- name: UpdateUserPassword :one
UPDATE users SET hashed_password = $1, email = $2, updated_at = NOW() WHERE id = $3
RETURNING *;

-- name: UpgradeChirpMembershipByuserId :one
UPDATE users SET is_chirpy_red = true, updated_at = NOW() WHERE id = $1
RETURNING *;
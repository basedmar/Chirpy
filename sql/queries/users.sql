-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, email, hashed_password)
VALUES (
gen_random_uuid(),
NOW(),
NOW(),
$1,
$2
)
RETURNING *;

-- name: DeleteUsers :exec
DELETE
FROM USERS;

-- name: GetUserViaEmail :one
SELECT *
FROM USERS
WHERE email = $1;

-- name: InsertToken :one
INSERT INTO refresh_token(token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
$1,
NOW(),
NOW(),
$2,
$3,
NULL
)
RETURNING *;


-- name: RevokeToken :exec
UPDATE refresh_token
SET revoked_at = NOW(), updated_at = NOW()
WHERE token = $1;

-- name: GetRefTokenUser :one
SELECT *
FROM refresh_token
WHERE token = $1;

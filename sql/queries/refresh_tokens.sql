-- name: CreateRefreshToken :one

INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at) VALUES($1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $2, $3) returning *;

-- name: GetRefreshToken :one
Select * from refresh_tokens where token = $1;

-- name: ExpireToken :exec
update refresh_tokens set revoked_at= CURRENT_TIMESTAMP,  expires_at = CURRENT_TIMESTAMP,   updated_at = CURRENT_TIMESTAMP where token = $1;

-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, email, hashed_password) values(gen_random_uuid(), NOW(), NOW(), $1, $2) returning id, email, created_at, updated_at, is_chirpy_red;

-- name: UpdateUser :one
UPDATE users set email = $2, hashed_password = $3, updated_at = CURRENT_TIMESTAMP where id = $1 returning id, email, created_at, updated_at, is_chirpy_red;

-- name: UpgradeUserToRed :exec
UPDATE users set is_chirpy_red = true where id = $1;

-- name: GetUser :one
Select * from users where email = $1;

-- name: ClearUsers :exec
DELETE from users;


-- name: CreateChirp :one
INSERT INTO chirps(id, created_at, updated_at, body, user_id) values(gen_random_uuid(), Now(), Now(), $1, $2) returning *;


-- name: GetChirps :many
select * from chirps order by created_at asc;


-- name: GetChirpsForAuthor :many
select * from chirps where user_id = $1 order by created_at asc;



-- name: GetChirp :one
select * from chirps where id = $1 limit 1;

-- name: DeleteChirp :exec
delete from chirps where id = $1 and user_id = $2;


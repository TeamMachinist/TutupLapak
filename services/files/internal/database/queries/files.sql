-- name: GetFile :one
SELECT * FROM files WHERE id = $1;

-- name: CreateFile :one
INSERT INTO files (id, user_id, file_uri, file_thumbnail_uri) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetFilesByUserID :many
SELECT * FROM files WHERE user_id = $1;

-- name: DeleteFile :exec
DELETE FROM files WHERE id = $1;

-- name: GetFilesByID :many
SELECT * FROM files WHERE id = ANY($1::uuid[]);
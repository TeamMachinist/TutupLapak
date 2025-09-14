-- name: GetFile :one
SELECT * FROM files WHERE id = $1;

-- name: CreateFile :one
INSERT INTO files (
  id, user_id, file_uri, file_thumbnail_uri
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetFilesByUser :many
SELECT * FROM files WHERE user_id = $1;

-- name: GetFileByIDAndUserID :one
SELECT * FROM files WHERE id = $1 AND user_id = $2;

-- name: DeleteFile :exec
DELETE FROM files WHERE id = $1;
CREATE TABLE files (
  fileid   text PRIMARY KEY,
  fileuri text      NOT NULL,
  filethumbnailuri  text NOT NULL
);

-- name: GetFiles :one
SELECT * FROM files
WHERE fileid = $1 LIMIT 1;

-- name: ListFiles :many
SELECT * FROM files;

-- name: CreateFiles :one
INSERT INTO files (
  fileid, fileuri, filethumbnailuri
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: DeleteFiles :exec
DELETE FROM files
WHERE fileid = $1;

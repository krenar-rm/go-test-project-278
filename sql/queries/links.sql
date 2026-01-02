-- name: GetAllLinks :many
SELECT id, original_url, short_name, created_at
FROM links
ORDER BY created_at DESC;

-- name: GetLinksWithPagination :many
SELECT id, original_url, short_name, created_at
FROM links
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountLinks :one
SELECT COUNT(*) as total FROM links;

-- name: GetLinkByID :one
SELECT id, original_url, short_name, created_at
FROM links
WHERE id = $1;

-- name: GetLinkByShortName :one
SELECT id, original_url, short_name, created_at
FROM links
WHERE short_name = $1;

-- name: CreateLink :one
INSERT INTO links (original_url, short_name)
VALUES ($1, $2)
RETURNING id, original_url, short_name, created_at;

-- name: UpdateLink :one
UPDATE links
SET original_url = $2, short_name = $3
WHERE id = $1
RETURNING id, original_url, short_name, created_at;

-- name: DeleteLink :exec
DELETE FROM links
WHERE id = $1;

-- name: CheckShortNameExists :one
SELECT COUNT(*) > 0 as exists
FROM links
WHERE short_name = $1 AND id != COALESCE($2, -1);


-- name: CreateLinkVisit :one
INSERT INTO link_visits (link_id, ip, user_agent, referer, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, link_id, ip, user_agent, referer, status, created_at;

-- name: GetAllLinkVisits :many
SELECT id, link_id, ip, user_agent, referer, status, created_at
FROM link_visits
ORDER BY created_at DESC;

-- name: GetLinkVisitsWithPagination :many
SELECT id, link_id, ip, user_agent, referer, status, created_at
FROM link_visits
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountLinkVisits :one
SELECT COUNT(*) as total FROM link_visits;

-- name: GetLinkVisitByID :one
SELECT id, link_id, ip, user_agent, referer, status, created_at
FROM link_visits
WHERE id = $1;

-- name: GetVisitsByLinkID :many
SELECT id, link_id, ip, user_agent, referer, status, created_at
FROM link_visits
WHERE link_id = $1
ORDER BY created_at DESC;

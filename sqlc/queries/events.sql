-- name: GetEvents :many
SELECT id, url, image, tag, title, "end" FROM events
ORDER BY "end", title;

-- name: CreateEvent :exec
INSERT INTO events (
  id, url, image, tag, title, "end"
)
SELECT $1, $2, $3, $4, $5, $6
WHERE NOT EXISTS (SELECT 1 FROM events WHERE id = $1);

-- name: DeleteExpiredEvents :exec
DELETE FROM events
WHERE "end" < CURRENT_DATE;
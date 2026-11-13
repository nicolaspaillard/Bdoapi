-- name: GetEvents :many
SELECT itemid, url, image, tag, title, "end" FROM events
ORDER BY "end", title;

-- name: CreateEvent :exec
INSERT INTO events (itemid, url, image, tag, title, "end")
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (itemid) DO UPDATE SET
  url = EXCLUDED.url,
  image = EXCLUDED.image,
  tag = EXCLUDED.tag,
  title = EXCLUDED.title,
  "end" = EXCLUDED."end";

-- name: DeleteExpiredEvents :exec
DELETE FROM events
WHERE "end" < CURRENT_DATE;
-- name: GetPearlItems :many
SELECT a.name,a.sold - b.sold AS sold,a.preorders FROM pearl_items a
INNER JOIN pearl_items b ON a.itemid = b.itemid
WHERE a.date = (SELECT MAX(date) FROM pearl_items)
AND b.date = (SELECT MAX(date) FROM pearl_items WHERE pearl_items.date <= $1)
ORDER BY a.name;

-- name: CreatePearlItem :exec
INSERT INTO pearl_items (
  itemid, name, date, sold, preorders
) 
SELECT $1, $2, $3, $4, $5
WHERE NOT EXISTS (SELECT 1 FROM pearl_items WHERE itemid = $1 AND date = $3);

-- name: DeleteOldPearlItems :exec
DELETE FROM pearl_items
WHERE date < now() - interval '3 weeks';
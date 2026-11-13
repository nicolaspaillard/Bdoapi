-- name: GetOutfits :many
SELECT a.itemid,a.name,a.sold - b.sold AS sold,a.preorders FROM outfits a
INNER JOIN outfits b ON a.itemid = b.itemid
WHERE a.date = (SELECT MAX(date) FROM outfits)
AND b.date = (SELECT MAX(date) FROM outfits WHERE outfits.date <= $1)
ORDER BY a.name;

-- name: CreateOutfit :exec
INSERT INTO outfits (itemid, name, date, sold, preorders)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (itemid, date) DO UPDATE SET
  name = EXCLUDED.name,
  sold = EXCLUDED.sold,
  preorders = EXCLUDED.preorders;

-- name: DeleteOldOutfits :exec
DELETE FROM outfits
WHERE date < now() - interval '4 weeks';
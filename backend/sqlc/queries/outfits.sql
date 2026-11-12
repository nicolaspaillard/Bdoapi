-- name: GetOutfits :many
SELECT a.name,a.sold - b.sold AS sold,a.preorders FROM outfits a
INNER JOIN outfits b ON a.itemid = b.itemid
WHERE a.date = (SELECT MAX(date) FROM outfits)
AND b.date = (SELECT MAX(date) FROM outfits WHERE outfits.date <= $1)
ORDER BY a.name;

-- name: CreateOutfit :exec
INSERT INTO outfits (
  itemid, name, date, sold, preorders
) 
SELECT $1, $2, $3, $4, $5
WHERE NOT EXISTS (SELECT 1 FROM outfits WHERE itemid = $1 AND date = $3);

-- name: DeleteOldOutfits :exec
DELETE FROM outfits
WHERE date < now() - interval '4 weeks';
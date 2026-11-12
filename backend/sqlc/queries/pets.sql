-- name: GetPets :many
SELECT a.name,a.sold - b.sold AS sold,a.preorders FROM pets a
INNER JOIN pets b ON a.itemid = b.itemid
WHERE a.date = (SELECT MAX(date) FROM pets)
AND b.date = (SELECT MAX(date) FROM pets WHERE pets.date <= $1)
ORDER BY a.name;

-- name: CreatePet :exec
INSERT INTO pets (
  itemid, name, date, sold, preorders
) 
SELECT $1, $2, $3, $4, $5
WHERE NOT EXISTS (SELECT 1 FROM pets WHERE itemid = $1 AND date = $3);

-- name: DeleteOldPets :exec
DELETE FROM pets
WHERE date < now() - interval '4 weeks';
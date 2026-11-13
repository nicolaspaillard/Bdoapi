-- name: UpsertLightstone :exec
INSERT INTO lightstones (itemid, name, current_stock, base_price)
VALUES ($1, $2, $3, $4)
ON CONFLICT (itemid) DO UPDATE SET
  name = EXCLUDED.name,
  current_stock = EXCLUDED.current_stock,
  base_price = EXCLUDED.base_price;

-- name: DeleteMissingLightstones :exec
DELETE FROM lightstones
WHERE NOT (itemid = ANY($1::bigint[]));

-- name: GetLightstones :many
SELECT itemid, name, current_stock, base_price FROM lightstones
ORDER BY name;
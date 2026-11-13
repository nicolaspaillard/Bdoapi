-- name: UpsertLightstone :exec
INSERT INTO lightstones (itemid, name, stock, price, silver)
VALUES (
  sqlc.arg(itemid)::bigint,
  sqlc.arg(name)::text,
  sqlc.arg(stock)::bigint,
  sqlc.arg(price)::bigint,
  (SELECT round(sqlc.arg(price)::bigint::numeric / q.quantity)::bigint
   FROM quantities q WHERE q.itemid = sqlc.arg(itemid)::bigint)
)
ON CONFLICT (itemid) DO UPDATE SET
  name = EXCLUDED.name,
  stock = EXCLUDED.stock,
  price = EXCLUDED.price,
  silver = EXCLUDED.silver;

-- name: DeleteMissingLightstones :exec
DELETE FROM lightstones
WHERE NOT (itemid = ANY($1::bigint[]));

-- name: GetLightstones :many
SELECT l.itemid, l.name, l.stock, l.price, q.quantity, l.silver
FROM lightstones l
LEFT JOIN quantities q ON q.itemid = l.itemid
ORDER BY l.name;
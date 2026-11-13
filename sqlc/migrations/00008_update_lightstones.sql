-- +goose Up
ALTER TABLE lightstones RENAME COLUMN current_stock TO stock;
ALTER TABLE lightstones RENAME COLUMN base_price TO price;
ALTER TABLE lightstones ADD COLUMN silver bigint;

CREATE TABLE IF NOT EXISTS quantities (
  itemid   bigint PRIMARY KEY,
  quantity bigint NOT NULL CHECK (quantity > 0)
);

-- +goose Down
DROP TABLE IF EXISTS quantities;
ALTER TABLE lightstones DROP COLUMN silver;
ALTER TABLE lightstones RENAME COLUMN price TO base_price;
ALTER TABLE lightstones RENAME COLUMN stock TO current_stock;
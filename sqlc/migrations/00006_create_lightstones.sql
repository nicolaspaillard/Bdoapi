-- +goose Up
CREATE TABLE IF NOT EXISTS lightstones (
  id            bigint PRIMARY KEY,
  name          text   NOT NULL,
  current_stock bigint NOT NULL,
  base_price    bigint NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS lightstones;
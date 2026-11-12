-- +goose up
CREATE TABLE IF NOT EXISTS pearl_items (
  id          bigserial PRIMARY KEY,
  itemid      bigint    NOT NULL,
  name        text      NOT NULL,
  date        timestamp NOT NULL,
  sold        bigint    NOT NULL,
  preorders   bigint    NOT NULL
);

-- +goose down
DROP TABLE IF EXISTS pearl_items;
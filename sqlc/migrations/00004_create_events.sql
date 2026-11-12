-- +goose Up
CREATE TABLE IF NOT EXISTS events (
  id      TEXT PRIMARY KEY,
  url     TEXT NOT NULL,
  image   TEXT NOT NULL,
  tag     TEXT NOT NULL,
  title   TEXT NOT NULL,
  "end"   DATE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS events;

-- +goose Up
ALTER TABLE outfits DROP COLUMN id;
ALTER TABLE outfits DROP CONSTRAINT outfits_itemid_date_key;
ALTER TABLE outfits ADD CONSTRAINT outfits_pkey PRIMARY KEY (itemid, date);

ALTER TABLE pets DROP COLUMN id;
ALTER TABLE pets DROP CONSTRAINT pets_itemid_date_key;
ALTER TABLE pets ADD CONSTRAINT pets_pkey PRIMARY KEY (itemid, date);

ALTER TABLE lightstones RENAME COLUMN id TO itemid;

ALTER TABLE events ALTER COLUMN id TYPE bigint USING id::bigint;
ALTER TABLE events RENAME COLUMN id TO itemid;

-- +goose Down
ALTER TABLE events RENAME COLUMN itemid TO id;
ALTER TABLE events ALTER COLUMN id TYPE text USING id::text;

ALTER TABLE lightstones RENAME COLUMN itemid TO id;

ALTER TABLE pets DROP CONSTRAINT pets_pkey;
ALTER TABLE pets ADD COLUMN id bigserial PRIMARY KEY;
ALTER TABLE pets ADD CONSTRAINT pets_itemid_date_key UNIQUE (itemid, date);

ALTER TABLE outfits DROP CONSTRAINT outfits_pkey;
ALTER TABLE outfits ADD COLUMN id bigserial PRIMARY KEY;
ALTER TABLE outfits ADD CONSTRAINT outfits_itemid_date_key UNIQUE (itemid, date);
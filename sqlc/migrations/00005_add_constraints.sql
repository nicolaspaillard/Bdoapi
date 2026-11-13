-- +goose Up
ALTER TABLE outfits ADD CONSTRAINT outfits_itemid_date_key UNIQUE (itemid, date);
ALTER TABLE pets ADD CONSTRAINT pets_itemid_date_key UNIQUE (itemid, date);

-- +goose Down
ALTER TABLE outfits DROP CONSTRAINT outfits_itemid_date_key;
ALTER TABLE pets DROP CONSTRAINT pets_itemid_date_key;
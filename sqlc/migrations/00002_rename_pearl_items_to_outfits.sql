-- +goose Up
ALTER TABLE pearl_items RENAME TO outfits;
ALTER SEQUENCE pearl_items_id_seq RENAME TO outfits_id_seq;
ALTER INDEX pearl_items_pkey RENAME TO outfits_pkey;

-- +goose Down
ALTER INDEX outfits_pkey RENAME TO pearl_items_pkey;
ALTER SEQUENCE outfits_id_seq RENAME TO pearl_items_id_seq;
ALTER TABLE outfits RENAME TO pearl_items;
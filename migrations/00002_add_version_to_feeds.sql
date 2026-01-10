-- +goose Up
ALTER TABLE feeds ADD COLUMN version integer NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE feeds DROP COLUMN version;

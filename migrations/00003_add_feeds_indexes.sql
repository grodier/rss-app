-- +goose Up
CREATE INDEX IF NOT EXISTS feeds_title_idx ON feeds USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS feeds_site_url_idx ON feeds (site_url);

-- +goose Down
DROP INDEX IF EXISTS feeds_title_idx;
DROP INDEX IF EXISTS feeds_site_url_idx;

-- +goose Up
CREATE TABLE IF NOT EXISTS feeds (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  title text NOT NULL,
  description text NOT NULL,
  url text NOT NULL UNIQUE,
  site_url text NOT NULL,
  language text
);

-- +goose Down
DROP TABLE IF EXISTS feeds;

package pgsql

import (
	"database/sql"

	"github.com/grodier/rss-app/internal/models"
)

type FeedService struct {
	db DBTX
}

func NewFeedService(db DBTX) *FeedService {
	return &FeedService{db: db}
}

func (fs *FeedService) Create(feed *models.Feed) error {
	query := `
    INSERT INTO feeds (title, description, url, site_url, language)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id, created_at
  `

	args := []any{feed.Title, feed.Description, feed.URL, feed.SiteURL, feed.Language}

	return fs.db.QueryRow(query, args...).Scan(&feed.ID, &feed.CreatedAt)
}

func (fs *FeedService) Get(id int64) (*models.Feed, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
    SELECT id, title, description, url, site_url, language, created_at
    FROM feeds
    WHERE id = $1
  `

	var feed models.Feed

	err := fs.db.QueryRow(query, id).Scan(
		&feed.ID,
		&feed.Title,
		&feed.Description,
		&feed.URL,
		&feed.SiteURL,
		&feed.Language,
		&feed.CreatedAt,
	)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &feed, nil
}

func (fs *FeedService) Update(feed *models.Feed) error {
	return nil
}

func (fs *FeedService) Delete(id int64) error {
	return nil
}

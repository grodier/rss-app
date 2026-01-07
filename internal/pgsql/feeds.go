package pgsql

import "github.com/grodier/rss-app/internal/models"

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
	return nil, nil
}

func (fs *FeedService) Update(feed *models.Feed) error {
	return nil
}

func (fs *FeedService) Delete(id int64) error {
	return nil
}

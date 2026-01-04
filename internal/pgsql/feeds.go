package pgsql

import "github.com/grodier/rss-app/internal/models"

type FeedService struct {
	db *DB
}

func NewFeedService(db *DB) *FeedService {
	return &FeedService{db: db}
}

func (fs *FeedService) Create(feed *models.Feed) error {
	return nil
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

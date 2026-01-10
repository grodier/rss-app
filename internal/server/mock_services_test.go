package server

import (
	"errors"
	"time"

	"github.com/grodier/rss-app/internal/models"
)

// mockFeedService is a mock implementation of models.FeedService for testing
type mockFeedService struct {
	createFn func(feed *models.Feed) error
	getFn    func(id int64) (*models.Feed, error)
	updateFn func(feed *models.Feed) error
	deleteFn func(id int64) error
}

func (m *mockFeedService) Create(feed *models.Feed) error {
	if m.createFn != nil {
		return m.createFn(feed)
	}
	// Default behavior: simulate successful creation with ID, timestamp, and version
	feed.ID = 1
	feed.CreatedAt = time.Now()
	feed.Version = 1
	return nil
}

func (m *mockFeedService) Get(id int64) (*models.Feed, error) {
	if m.getFn != nil {
		return m.getFn(id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockFeedService) Update(feed *models.Feed) error {
	if m.updateFn != nil {
		return m.updateFn(feed)
	}
	return errors.New("not implemented")
}

func (m *mockFeedService) Delete(id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return errors.New("not implemented")
}

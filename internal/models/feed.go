package models

import (
	"time"

	"github.com/grodier/rss-app/internal/validator"
)

type Feed struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	SiteURL     string    `json:"site_url"`
	CreatedAt   time.Time `json:"-"`
	Language    string    `json:"language,omitzero"`
}

type FeedService interface {
	Create(feed *Feed) error
	Get(id int64) (*Feed, error)
	Update(feed *Feed) error
	Delete(id int64) error
}

func ValidateFeed(v *validator.Validator, feed *Feed) {
	v.Check(feed.Title != "", "title", "must be provided")
	v.Check(len(feed.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(feed.Description != "", "description", "must be provided")
	v.Check(feed.URL != "", "url", "must be provided")
	v.Check(feed.SiteURL != "", "site_url", "must be provided")
}

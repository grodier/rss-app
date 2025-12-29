package models

import (
	"time"
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

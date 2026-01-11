package pgsql

import (
	"context"
	"database/sql"
	"time"

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
    RETURNING id, created_at, version`

	args := []any{feed.Title, feed.Description, feed.URL, feed.SiteURL, feed.Language}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return fs.db.QueryRowContext(ctx, query, args...).Scan(&feed.ID, &feed.CreatedAt, &feed.Version)
}

func (fs *FeedService) Get(id int64) (*models.Feed, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
    SELECT id, title, description, url, site_url, language, created_at, version
    FROM feeds
    WHERE id = $1`

	var feed models.Feed

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := fs.db.QueryRowContext(ctx, query, id).Scan(
		&feed.ID,
		&feed.Title,
		&feed.Description,
		&feed.URL,
		&feed.SiteURL,
		&feed.Language,
		&feed.CreatedAt,
		&feed.Version,
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

func (fs *FeedService) GetAll(title, url string, filters models.Filters) ([]*models.Feed, error) {
	query := `
    SELECT id, title, description, url, site_url, language, created_at, version
    FROM feeds
    ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := fs.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feeds := []*models.Feed{}
	for rows.Next() {
		var feed models.Feed
		err := rows.Scan(
			&feed.ID,
			&feed.Title,
			&feed.Description,
			&feed.URL,
			&feed.SiteURL,
			&feed.Language,
			&feed.CreatedAt,
			&feed.Version,
		)
		if err != nil {
			return nil, err
		}

		feeds = append(feeds, &feed)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return feeds, nil
}

func (fs *FeedService) Update(feed *models.Feed) error {
	if feed.ID < 1 {
		return ErrRecordNotFound
	}

	query := `
    UPDATE feeds
    SET title = $1, description = $2, url = $3, site_url = $4, language = $5, version = version + 1
    WHERE id = $6 AND version = $7
    RETURNING version`

	args := []any{
		feed.Title,
		feed.Description,
		feed.URL,
		feed.SiteURL,
		feed.Language,
		feed.ID,
		feed.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := fs.db.QueryRowContext(ctx, query, args...).Scan(&feed.Version)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (fs *FeedService) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM feeds
        WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := fs.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

package pgsql

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/grodier/rss-app/internal/models"
)

func TestFeedService_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	expectedID := int64(1)
	expectedCreatedAt := time.Now()

	mock.ExpectQuery(`INSERT INTO feeds`).
		WithArgs("Test Feed", "A test description", "https://example.com/feed.xml", "https://example.com", "en").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(expectedID, expectedCreatedAt))

	fs := NewFeedService(db)

	feed := &models.Feed{
		Title:       "Test Feed",
		Description: "A test description",
		URL:         "https://example.com/feed.xml",
		SiteURL:     "https://example.com",
		Language:    "en",
	}

	err = fs.Create(feed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if feed.ID != expectedID {
		t.Errorf("expected ID %d, got %d", expectedID, feed.ID)
	}

	if !feed.CreatedAt.Equal(expectedCreatedAt) {
		t.Errorf("expected CreatedAt %v, got %v", expectedCreatedAt, feed.CreatedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFeedService_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`INSERT INTO feeds`).
		WithArgs("Test Feed", "A test description", "https://example.com/feed.xml", "https://example.com", "en").
		WillReturnError(sqlmock.ErrCancelled)

	fs := NewFeedService(db)

	feed := &models.Feed{
		Title:       "Test Feed",
		Description: "A test description",
		URL:         "https://example.com/feed.xml",
		SiteURL:     "https://example.com",
		Language:    "en",
	}

	err = fs.Create(feed)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

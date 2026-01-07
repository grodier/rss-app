package pgsql

import (
	"database/sql"
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

func TestFeedService_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	expectedID := int64(1)
	expectedCreatedAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "title", "description", "url", "site_url", "language", "created_at"}).
		AddRow(expectedID, "Test Feed", "A test description", "https://example.com/feed.xml", "https://example.com", "en", expectedCreatedAt)

	mock.ExpectQuery(`SELECT .+ FROM feeds WHERE id = \$1`).
		WithArgs(expectedID).
		WillReturnRows(rows)

	fs := NewFeedService(db)

	feed, err := fs.Get(expectedID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if feed.ID != expectedID {
		t.Errorf("got ID %d, want %d", feed.ID, expectedID)
	}
	if feed.Title != "Test Feed" {
		t.Errorf("got Title %q, want %q", feed.Title, "Test Feed")
	}
	if feed.Description != "A test description" {
		t.Errorf("got Description %q, want %q", feed.Description, "A test description")
	}
	if feed.URL != "https://example.com/feed.xml" {
		t.Errorf("got URL %q, want %q", feed.URL, "https://example.com/feed.xml")
	}
	if feed.SiteURL != "https://example.com" {
		t.Errorf("got SiteURL %q, want %q", feed.SiteURL, "https://example.com")
	}
	if !feed.CreatedAt.Equal(expectedCreatedAt) {
		t.Errorf("got CreatedAt %v, want %v", feed.CreatedAt, expectedCreatedAt)
	}
	if feed.Language != "en" {
		t.Errorf("got Language %q, want %q", feed.Language, "en")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFeedService_Get_Errors(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		mockError error // nil means no DB call expected
		wantError error
	}{
		{"invalid id zero", 0, nil, ErrRecordNotFound},
		{"invalid id negative", -1, nil, ErrRecordNotFound},
		{"record not found", 999, sql.ErrNoRows, ErrRecordNotFound},
		{"database error", 1, sqlmock.ErrCancelled, sqlmock.ErrCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			if tt.mockError != nil {
				mock.ExpectQuery(`SELECT .+ FROM feeds WHERE id = \$1`).
					WithArgs(tt.id).
					WillReturnError(tt.mockError)
			}

			fs := NewFeedService(db)

			feed, err := fs.Get(tt.id)

			if feed != nil {
				t.Error("expected nil feed")
			}
			if err != tt.wantError {
				t.Errorf("got error %v, want %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

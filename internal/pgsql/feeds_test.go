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
	expectedVersion := int32(1)

	mock.ExpectQuery(`INSERT INTO feeds`).
		WithArgs("Test Feed", "A test description", "https://example.com/feed.xml", "https://example.com", "en").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "version"}).AddRow(expectedID, expectedCreatedAt, expectedVersion))

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

	if feed.Version != expectedVersion {
		t.Errorf("expected Version %d, got %d", expectedVersion, feed.Version)
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
	expectedVersion := int32(1)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "url", "site_url", "language", "created_at", "version"}).
		AddRow(expectedID, "Test Feed", "A test description", "https://example.com/feed.xml", "https://example.com", "en", expectedCreatedAt, expectedVersion)

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
	if feed.Version != expectedVersion {
		t.Errorf("got Version %d, want %d", feed.Version, expectedVersion)
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

func TestFeedService_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`UPDATE feeds SET .+ WHERE id = \$6 AND version = \$7`).
		WithArgs("Updated Feed", "Updated description", "https://example.com/updated.xml", "https://example.com/updated", "es", int64(1), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(int32(2)))

	fs := NewFeedService(db)

	feed := &models.Feed{
		ID:          1,
		Title:       "Updated Feed",
		Description: "Updated description",
		URL:         "https://example.com/updated.xml",
		SiteURL:     "https://example.com/updated",
		Language:    "es",
		Version:     1,
	}

	err = fs.Update(feed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if feed.Version != 2 {
		t.Errorf("expected Version 2, got %d", feed.Version)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFeedService_Update_Errors(t *testing.T) {
	tests := []struct {
		name        string
		feedID      int64
		feedVersion int32
		mockError   error // nil means no DB call expected (invalid ID)
		wantError   error
	}{
		{"invalid id zero", 0, 1, nil, ErrRecordNotFound},
		{"invalid id negative", -1, 1, nil, ErrRecordNotFound},
		{"edit conflict", 1, 1, sql.ErrNoRows, ErrEditConflict},
		{"database error", 1, 1, sqlmock.ErrCancelled, sqlmock.ErrCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			// Only set up mock expectation if ID is valid (DB call will be made)
			if tt.feedID >= 1 {
				mock.ExpectQuery(`UPDATE feeds SET .+ WHERE id = \$6 AND version = \$7`).
					WithArgs("Test Feed", "A test description", "https://example.com/feed.xml", "https://example.com", "en", tt.feedID, tt.feedVersion).
					WillReturnError(tt.mockError)
			}

			fs := NewFeedService(db)

			feed := &models.Feed{
				ID:          tt.feedID,
				Title:       "Test Feed",
				Description: "A test description",
				URL:         "https://example.com/feed.xml",
				SiteURL:     "https://example.com",
				Language:    "en",
				Version:     tt.feedVersion,
			}

			err = fs.Update(feed)

			if err != tt.wantError {
				t.Errorf("got error %v, want %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestFeedService_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`DELETE FROM feeds WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	fs := NewFeedService(db)

	err = fs.Delete(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFeedService_Delete_Errors(t *testing.T) {
	tests := []struct {
		name         string
		id           int64
		mockError    error // nil means no DB call expected (invalid ID)
		rowsAffected int64 // 0 means record not found
		wantError    error
	}{
		{"invalid id zero", 0, nil, 0, ErrRecordNotFound},
		{"invalid id negative", -1, nil, 0, ErrRecordNotFound},
		{"record not found", 999, nil, 0, ErrRecordNotFound},
		{"database error", 1, sqlmock.ErrCancelled, 0, sqlmock.ErrCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			// Only set up mock expectation if ID is valid (DB call will be made)
			if tt.id >= 1 {
				if tt.mockError != nil {
					mock.ExpectExec(`DELETE FROM feeds WHERE id = \$1`).
						WithArgs(tt.id).
						WillReturnError(tt.mockError)
				} else {
					mock.ExpectExec(`DELETE FROM feeds WHERE id = \$1`).
						WithArgs(tt.id).
						WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))
				}
			}

			fs := NewFeedService(db)

			err = fs.Delete(tt.id)

			if err != tt.wantError {
				t.Errorf("got error %v, want %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

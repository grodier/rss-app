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

func TestFeedService_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	expectedCreatedAt := time.Now()

	rows := sqlmock.NewRows([]string{"count", "id", "title", "description", "url", "site_url", "language", "created_at", "version"}).
		AddRow(2, int64(1), "Test Feed 1", "Description 1", "https://example1.com/feed.xml", "https://example1.com", "en", expectedCreatedAt, int32(1)).
		AddRow(2, int64(2), "Test Feed 2", "Description 2", "https://example2.com/feed.xml", "https://example2.com", "es", expectedCreatedAt, int32(1))

	mock.ExpectQuery(`SELECT count\(\*\) OVER\(\), id, title, description, url, site_url, language, created_at, version FROM feeds`).
		WithArgs("", "", 20, 0).
		WillReturnRows(rows)

	fs := NewFeedService(db)

	filters := models.Filters{
		Page:         1,
		PageSize:     20,
		Sort:         "id",
		SortSafelist: []string{"id", "title", "url", "-id", "-title", "-url"},
	}

	feeds, metadata, err := fs.GetAll("", "", filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(feeds) != 2 {
		t.Fatalf("expected 2 feeds, got %d", len(feeds))
	}

	// Check first feed
	if feeds[0].ID != 1 {
		t.Errorf("got ID %d, want 1", feeds[0].ID)
	}
	if feeds[0].Title != "Test Feed 1" {
		t.Errorf("got Title %q, want %q", feeds[0].Title, "Test Feed 1")
	}
	if feeds[0].Description != "Description 1" {
		t.Errorf("got Description %q, want %q", feeds[0].Description, "Description 1")
	}
	if feeds[0].URL != "https://example1.com/feed.xml" {
		t.Errorf("got URL %q, want %q", feeds[0].URL, "https://example1.com/feed.xml")
	}
	if feeds[0].SiteURL != "https://example1.com" {
		t.Errorf("got SiteURL %q, want %q", feeds[0].SiteURL, "https://example1.com")
	}
	if feeds[0].Language != "en" {
		t.Errorf("got Language %q, want %q", feeds[0].Language, "en")
	}

	// Check second feed
	if feeds[1].ID != 2 {
		t.Errorf("got ID %d, want 2", feeds[1].ID)
	}
	if feeds[1].Title != "Test Feed 2" {
		t.Errorf("got Title %q, want %q", feeds[1].Title, "Test Feed 2")
	}

	// Check metadata
	if metadata.TotalRecords != 2 {
		t.Errorf("got TotalRecords %d, want 2", metadata.TotalRecords)
	}
	if metadata.CurrentPage != 1 {
		t.Errorf("got CurrentPage %d, want 1", metadata.CurrentPage)
	}
	if metadata.PageSize != 20 {
		t.Errorf("got PageSize %d, want 20", metadata.PageSize)
	}
	if metadata.FirstPage != 1 {
		t.Errorf("got FirstPage %d, want 1", metadata.FirstPage)
	}
	if metadata.LastPage != 1 {
		t.Errorf("got LastPage %d, want 1", metadata.LastPage)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFeedService_GetAll_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"count", "id", "title", "description", "url", "site_url", "language", "created_at", "version"})

	mock.ExpectQuery(`SELECT count\(\*\) OVER\(\), id, title, description, url, site_url, language, created_at, version FROM feeds`).
		WithArgs("", "", 20, 0).
		WillReturnRows(rows)

	fs := NewFeedService(db)

	filters := models.Filters{
		Page:         1,
		PageSize:     20,
		Sort:         "id",
		SortSafelist: []string{"id", "title", "url", "-id", "-title", "-url"},
	}

	feeds, metadata, err := fs.GetAll("", "", filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feeds))
	}

	// Metadata should be empty when no records
	if metadata.TotalRecords != 0 {
		t.Errorf("got TotalRecords %d, want 0", metadata.TotalRecords)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFeedService_GetAll_WithFilters(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		url          string
		page         int
		pageSize     int
		sort         string
		wantLimit    int
		wantOffset   int
	}{
		{
			name:       "filter by title",
			title:      "technology",
			url:        "",
			page:       1,
			pageSize:   20,
			sort:       "id",
			wantLimit:  20,
			wantOffset: 0,
		},
		{
			name:       "filter by url",
			title:      "",
			url:        "https://example.com",
			page:       1,
			pageSize:   20,
			sort:       "id",
			wantLimit:  20,
			wantOffset: 0,
		},
		{
			name:       "custom pagination page 2",
			title:      "",
			url:        "",
			page:       2,
			pageSize:   10,
			sort:       "id",
			wantLimit:  10,
			wantOffset: 10,
		},
		{
			name:       "custom pagination page 3",
			title:      "",
			url:        "",
			page:       3,
			pageSize:   5,
			sort:       "id",
			wantLimit:  5,
			wantOffset: 10,
		},
		{
			name:       "combined title and url filter",
			title:      "tech",
			url:        "https://example.com",
			page:       1,
			pageSize:   20,
			sort:       "id",
			wantLimit:  20,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			rows := sqlmock.NewRows([]string{"count", "id", "title", "description", "url", "site_url", "language", "created_at", "version"})

			mock.ExpectQuery(`SELECT count\(\*\) OVER\(\), id, title, description, url, site_url, language, created_at, version FROM feeds`).
				WithArgs(tt.title, tt.url, tt.wantLimit, tt.wantOffset).
				WillReturnRows(rows)

			fs := NewFeedService(db)

			filters := models.Filters{
				Page:         tt.page,
				PageSize:     tt.pageSize,
				Sort:         tt.sort,
				SortSafelist: []string{"id", "title", "url", "-id", "-title", "-url"},
			}

			_, _, err = fs.GetAll(tt.title, tt.url, filters)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestFeedService_GetAll_Pagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	expectedCreatedAt := time.Now()

	// Simulate 25 total records, page 2 with page_size 10
	rows := sqlmock.NewRows([]string{"count", "id", "title", "description", "url", "site_url", "language", "created_at", "version"}).
		AddRow(25, int64(11), "Test Feed 11", "Description 11", "https://example11.com/feed.xml", "https://example11.com", "en", expectedCreatedAt, int32(1)).
		AddRow(25, int64(12), "Test Feed 12", "Description 12", "https://example12.com/feed.xml", "https://example12.com", "en", expectedCreatedAt, int32(1))

	mock.ExpectQuery(`SELECT count\(\*\) OVER\(\), id, title, description, url, site_url, language, created_at, version FROM feeds`).
		WithArgs("", "", 10, 10). // limit 10, offset 10 for page 2
		WillReturnRows(rows)

	fs := NewFeedService(db)

	filters := models.Filters{
		Page:         2,
		PageSize:     10,
		Sort:         "id",
		SortSafelist: []string{"id", "title", "url", "-id", "-title", "-url"},
	}

	feeds, metadata, err := fs.GetAll("", "", filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(feeds) != 2 {
		t.Fatalf("expected 2 feeds, got %d", len(feeds))
	}

	// Check metadata for pagination
	if metadata.TotalRecords != 25 {
		t.Errorf("got TotalRecords %d, want 25", metadata.TotalRecords)
	}
	if metadata.CurrentPage != 2 {
		t.Errorf("got CurrentPage %d, want 2", metadata.CurrentPage)
	}
	if metadata.PageSize != 10 {
		t.Errorf("got PageSize %d, want 10", metadata.PageSize)
	}
	if metadata.FirstPage != 1 {
		t.Errorf("got FirstPage %d, want 1", metadata.FirstPage)
	}
	if metadata.LastPage != 3 {
		t.Errorf("got LastPage %d, want 3", metadata.LastPage)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFeedService_GetAll_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT count\(\*\) OVER\(\), id, title, description, url, site_url, language, created_at, version FROM feeds`).
		WithArgs("", "", 20, 0).
		WillReturnError(sqlmock.ErrCancelled)

	fs := NewFeedService(db)

	filters := models.Filters{
		Page:         1,
		PageSize:     20,
		Sort:         "id",
		SortSafelist: []string{"id", "title", "url", "-id", "-title", "-url"},
	}

	feeds, metadata, err := fs.GetAll("", "", filters)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if feeds != nil {
		t.Errorf("expected nil feeds, got %v", feeds)
	}

	if metadata != (models.Metadata{}) {
		t.Errorf("expected empty metadata, got %v", metadata)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFeedService_GetAll_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Return invalid data that will cause a scan error (string where int expected)
	rows := sqlmock.NewRows([]string{"count", "id", "title", "description", "url", "site_url", "language", "created_at", "version"}).
		AddRow("invalid", int64(1), "Test Feed", "Description", "https://example.com/feed.xml", "https://example.com", "en", time.Now(), int32(1))

	mock.ExpectQuery(`SELECT count\(\*\) OVER\(\), id, title, description, url, site_url, language, created_at, version FROM feeds`).
		WithArgs("", "", 20, 0).
		WillReturnRows(rows)

	fs := NewFeedService(db)

	filters := models.Filters{
		Page:         1,
		PageSize:     20,
		Sort:         "id",
		SortSafelist: []string{"id", "title", "url", "-id", "-title", "-url"},
	}

	feeds, metadata, err := fs.GetAll("", "", filters)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if feeds != nil {
		t.Errorf("expected nil feeds, got %v", feeds)
	}

	if metadata != (models.Metadata{}) {
		t.Errorf("expected empty metadata, got %v", metadata)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

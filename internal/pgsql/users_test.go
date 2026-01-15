package pgsql

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/grodier/rss-app/internal/models"
)

func TestUserService_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	expectedID := int64(1)
	expectedCreatedAt := time.Now()
	expectedVersion := 1

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("Test User", "test@example.com", sqlmock.AnyArg(), false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "version"}).AddRow(expectedID, expectedCreatedAt, expectedVersion))

	us := NewUserService(db)

	user := &models.User{
		Name:      "Test User",
		Email:     "test@example.com",
		Activated: false,
	}
	err = user.Password.Set("password123")
	if err != nil {
		t.Fatalf("failed to set password: %v", err)
	}

	err = us.Create(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != expectedID {
		t.Errorf("expected ID %d, got %d", expectedID, user.ID)
	}

	if !user.CreatedAt.Equal(expectedCreatedAt) {
		t.Errorf("expected CreatedAt %v, got %v", expectedCreatedAt, user.CreatedAt)
	}

	if user.Version != expectedVersion {
		t.Errorf("expected Version %d, got %d", expectedVersion, user.Version)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUserService_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("Test User", "test@example.com", sqlmock.AnyArg(), false).
		WillReturnError(sqlmock.ErrCancelled)

	us := NewUserService(db)

	user := &models.User{
		Name:      "Test User",
		Email:     "test@example.com",
		Activated: false,
	}
	err = user.Password.Set("password123")
	if err != nil {
		t.Fatalf("failed to set password: %v", err)
	}

	err = us.Create(user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUserService_Create_DuplicateEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	duplicateEmailErr := &mockPqError{message: `pq: duplicate key value violates unique constraint "users_email_key"`}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("Test User", "test@example.com", sqlmock.AnyArg(), false).
		WillReturnError(duplicateEmailErr)

	us := NewUserService(db)

	user := &models.User{
		Name:      "Test User",
		Email:     "test@example.com",
		Activated: false,
	}
	err = user.Password.Set("password123")
	if err != nil {
		t.Fatalf("failed to set password: %v", err)
	}

	err = us.Create(user)
	if err != ErrDuplicateEmail {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// mockPqError is a helper type to simulate pq error messages
type mockPqError struct {
	message string
}

func (e *mockPqError) Error() string {
	return e.message
}

func TestUserService_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	expectedID := int64(1)
	expectedCreatedAt := time.Now()
	expectedName := "Test User"
	expectedEmail := "test@example.com"
	expectedPasswordHash := []byte("hashedpassword")
	expectedActivated := true
	expectedVersion := 1

	rows := sqlmock.NewRows([]string{"id", "created_at", "name", "email", "password_hash", "activated", "version"}).
		AddRow(expectedID, expectedCreatedAt, expectedName, expectedEmail, expectedPasswordHash, expectedActivated, expectedVersion)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE email = \$1`).
		WithArgs(expectedEmail).
		WillReturnRows(rows)

	us := NewUserService(db)

	user, err := us.GetByEmail(expectedEmail)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != expectedID {
		t.Errorf("got ID %d, want %d", user.ID, expectedID)
	}
	if !user.CreatedAt.Equal(expectedCreatedAt) {
		t.Errorf("got CreatedAt %v, want %v", user.CreatedAt, expectedCreatedAt)
	}
	if user.Name != expectedName {
		t.Errorf("got Name %q, want %q", user.Name, expectedName)
	}
	if user.Email != expectedEmail {
		t.Errorf("got Email %q, want %q", user.Email, expectedEmail)
	}
	if string(user.Password.Hash) != string(expectedPasswordHash) {
		t.Errorf("got Password.Hash %q, want %q", user.Password.Hash, expectedPasswordHash)
	}
	if user.Activated != expectedActivated {
		t.Errorf("got Activated %v, want %v", user.Activated, expectedActivated)
	}
	if user.Version != expectedVersion {
		t.Errorf("got Version %d, want %d", user.Version, expectedVersion)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUserService_GetByEmail_Errors(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		mockError error
		wantError error
	}{
		{"record not found", "notfound@example.com", sql.ErrNoRows, ErrRecordNotFound},
		{"database error", "test@example.com", sqlmock.ErrCancelled, sqlmock.ErrCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			mock.ExpectQuery(`SELECT .+ FROM users WHERE email = \$1`).
				WithArgs(tt.email).
				WillReturnError(tt.mockError)

			us := NewUserService(db)

			user, err := us.GetByEmail(tt.email)

			if user != nil {
				t.Error("expected nil user")
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

package pgsql

import (
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

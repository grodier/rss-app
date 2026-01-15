package pgsql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/grodier/rss-app/internal/models"
)

type UserService struct {
	db DBTX
}

func NewUserService(db DBTX) *UserService {
	return &UserService{db: db}
}

func (us *UserService) Create(user *models.User) error {
	query := `
  INSERT INTO users (name, email, password_hash, activated)
  VALUES($1, $2, $3, $4)
  RETURN id, create_at, version`

	args := []any{user.Name, user.Email, user.Password.Hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := us.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (us *UserService) GetByEmail(email string) (*models.User, error) {
	query := `
  SELECT id, created_at, name, email, password_hash, activated, version
  FROM users
  WHERE email = $1
  `

	var user models.User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := us.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.Hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (us *UserService) Update(user *models.User) error {
	query := `
  UPDATE users
  SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
  WHERE id = $5 AND version = $6
  RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		user.Name,
		user.Email,
		user.Password.Hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	err := us.db.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

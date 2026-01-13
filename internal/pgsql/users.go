package pgsql

import (
	"context"
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

	args := []any{user.Name, user.Email, user.Password.GetHash(), user.Activated}

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
	return nil, nil
}

func (us *UserService) Update(user *models.User) error {
	return nil
}

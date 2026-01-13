package pgsql

import "github.com/grodier/rss-app/internal/models"

type UserService struct {
	db DBTX
}

func NewUserService(db DBTX) *UserService {
	return &UserService{db: db}
}

func (us *UserService) Create(user *models.User) error {
	return nil
}

func (us *UserService) GetByEmail(email string) (*models.User, error) {
	return nil, nil
}

func (us *UserService) Update(user *models.User) error {
	return nil
}

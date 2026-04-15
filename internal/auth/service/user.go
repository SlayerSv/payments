package service

import (
	"context"

	"github.com/SlayerSv/payments/internal/auth/models"
	"github.com/SlayerSv/payments/internal/auth/repo"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	DB repo.User
}

func NewUser(db repo.User) *User {
	return &User{DB: db}
}

func (us *User) Create(ctx context.Context, email string) (uuid.UUID, error) {
	return us.DB.Create(ctx, email)
}

func (us *User) Get(ctx context.Context, id uuid.UUID) (models.User, error) {
	return us.DB.Get(ctx, id)
}

func (us *User) GetByEmail(ctx context.Context, email string) (models.User, error) {
	return us.DB.GetByEmail(ctx, email)
}

func (us *User) UpdateName(ctx context.Context, id uuid.UUID, newName string) (models.User, error) {
	return us.DB.UpdateName(ctx, id, newName)
}

func (us *User) UpdatePassword(ctx context.Context, id uuid.UUID, newPassword string) (models.User, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	return us.DB.UpdatePassword(ctx, id, string(hashedPass))
}

func (us *User) Delete(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return us.DB.Delete(ctx, id)
}

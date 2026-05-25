package service

import (
	"context"

	"github.com/SlayerSv/payments/internal/shared/models"
)

type UserClient interface {
	Get(ctx context.Context) (models.UserDTO, error)
	Update(ctx context.Context, name, password *string) (models.UserDTO, error)
}

type User struct {
	userClient UserClient
}

func NewUser(userClient UserClient) *User {
	return &User{userClient: userClient}
}

func (u *User) Get(ctx context.Context) (models.UserDTO, error) {
	return u.userClient.Get(ctx)
}

func (u *User) Update(ctx context.Context, name, password *string) (models.UserDTO, error) {
	return u.userClient.Update(ctx, name, password)
}

package service

import (
	"context"

	"github.com/SlayerSv/payments/internal/auth/models"
)

type UserClient interface {
	Get(ctx context.Context) (models.User, error)
	Update(ctx context.Context, name, password *string) (models.User, error)
}

type User struct {
	userClient UserClient
}

func NewUser(authClient AuthClient, userClient UserClient) *User {
	return &User{userClient: userClient}
}

func (u *User) Get(ctx context.Context) (models.User, error) {
	return u.userClient.Get(ctx)
}

func (u *User) Update(ctx context.Context, name, password *string) (models.User, error) {
	return u.userClient.Update(ctx, name, password)
}

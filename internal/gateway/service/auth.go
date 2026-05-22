package service

import "context"

type AuthClient interface {
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email string) error
	Restore(ctx context.Context, email string) error
}

type Auth struct {
	authClient AuthClient
}

func NewAuth(authClient AuthClient) *Auth {
	return &Auth{authClient: authClient}
}

func (a *Auth) Login(ctx context.Context, email, password string) (string, error) {
	return a.authClient.Login(ctx, email, password)
}

func (a *Auth) Register(ctx context.Context, email string) error {
	return a.authClient.Register(ctx, email)
}

func (a *Auth) Restore(ctx context.Context, email string) error {
	return a.authClient.Restore(ctx, email)
}

package adapter

import (
	"context"
	"fmt"

	"github.com/SlayerSv/payments/gen/auth"
	"github.com/SlayerSv/payments/internal/shared/errs"
)

type Auth struct {
	authClient auth.AuthServiceClient
}

func NewAuth(authClient auth.AuthServiceClient) *Auth {
	return &Auth{authClient: authClient}
}

func (a *Auth) Login(ctx context.Context, email, password string) (string, error) {
	resp, err := a.authClient.Login(ctx, &auth.LoginRequest{Email: email, Password: password})
	if err != nil {
		return "", fmt.Errorf("%w: error login: %w", errs.Internal, err)
	}
	return resp.GetToken(), nil
}

func (a *Auth) Register(ctx context.Context, email string) error {
	_, err := a.authClient.Register(ctx, &auth.RegisterRequest{Email: email})
	if err != nil {
		return fmt.Errorf("%w: error register grpc client call: %w", errs.Internal, err)
	}
	return nil
}

func (a *Auth) Restore(ctx context.Context, email string) error {
	_, err := a.authClient.Restore(ctx, &auth.RestoreRequest{Email: email})
	if err != nil {
		return fmt.Errorf("%w: error register grpc client call: %w", errs.Internal, err)
	}
	return nil
}

package repo

import (
	"context"

	"github.com/SlayerSv/payments/internal/auth/models"
)

type OTP interface {
	Create(ctx context.Context, otp models.OTP) (models.OTP, error)
	Get(ctx context.Context, email string) (models.OTP, error)
	Delete(ctx context.Context, email string) (string, error)
}

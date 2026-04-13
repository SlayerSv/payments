package repo

import (
	"context"

	"github.com/SlayerSv/payments/internal/auth/models"
	"github.com/google/uuid"
)

type Code interface {
	Create(ctx context.Context, otp models.OTPCode) (models.OTPCode, error)
	Get(ctx context.Context, userID uuid.UUID) (models.OTPCode, error)
	Delete(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

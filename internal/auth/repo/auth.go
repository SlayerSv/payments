package repo

import (
	"context"

	"github.com/SlayerSv/payments/internal/auth/models"
	"github.com/google/uuid"
)

type Auth interface {
	Register(ctx context.Context, otp models.OTP) (uuid.UUID, error)
}

package repo

import (
	"context"

	"github.com/SlayerSv/payments/internal/auth/models"
	"github.com/google/uuid"
)

type User interface {
	Create(ctx context.Context, user models.User) (uuid.UUID, error)
	Get(ctx context.Context, id uuid.UUID) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
	UpdateName(ctx context.Context, id uuid.UUID, newName string) (models.User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, newPassword string) (models.User, error)
	Delete(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

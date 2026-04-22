package repository

import (
	"context"

	"github.com/SlayerSv/payments/internal/trans/models"
	"github.com/google/uuid"
)

type Transaction interface {
	Create(ctx context.Context, tx models.Transaction) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (models.Transaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.TransactionStatus) error
	GetAccHistory(ctx context.Context, accountID uuid.UUID) ([]models.Transaction, error)
}

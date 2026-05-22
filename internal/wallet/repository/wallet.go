package repository

import (
	"context"

	"github.com/SlayerSv/payments/internal/wallet/models"
	"github.com/google/uuid"
)

type Wallet interface {
	Create(ctx context.Context, ownerID uuid.UUID) (uuid.UUID, error)
	Get(ctx context.Context, ownerID, ID uuid.UUID) (models.Wallet, error)
	GetAll(ctx context.Context, ownerID uuid.UUID) ([]models.Wallet, error)
	// Проверка ключа идемпотентности
	GetIdempotencyResponse(ctx context.Context, key string) ([]byte, bool, error)
	// обновление баланса: делает всё в одной DB-транзакции
	UpdateBalanceAtomic(ctx context.Context, args models.UpdateBalanceParams) error
	Delete(ctx context.Context, ownerID, walletID uuid.UUID) error
}

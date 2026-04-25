package repository

import (
	"context"

	"github.com/SlayerSv/payments/internal/wallet/models"
	"github.com/google/uuid"
)

type Wallet interface {
	CreateAccount(ctx context.Context, ownerID uuid.UUID) (uuid.UUID, error)
	GetAccount(ctx context.Context, ownerID, ID uuid.UUID) (models.Account, error)
	GetAccounts(ctx context.Context, ownerID uuid.UUID) ([]models.Account, error)
	// Проверка ключа идемпотентности
	GetIdempotencyResponse(ctx context.Context, key string) ([]byte, bool, error)
	// обновление баланса: делает всё в одной DB-транзакции
	UpdateBalanceAtomic(ctx context.Context, args models.UpdateBalanceParams) error
	DeleteAccount(ctx context.Context, ownerID, accountID uuid.UUID) error
}

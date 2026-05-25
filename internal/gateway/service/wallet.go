package service

import (
	"context"

	"github.com/SlayerSv/payments/internal/shared/models"
)

type WalletClient interface {
	Get(ctx context.Context, ID string) (models.WalletDTO, error)
	GetAll(ctx context.Context) ([]models.WalletDTO, error)
	Create(ctx context.Context) (string, error)
	Delete(ctx context.Context, ID string) error
}

type Wallet struct {
	WalletClient WalletClient
}

func NewWallet(WalletClient WalletClient) *Wallet {
	return &Wallet{WalletClient: WalletClient}
}

func (w *Wallet) Get(ctx context.Context, ID string) (models.WalletDTO, error) {
	return w.WalletClient.Get(ctx, ID)
}

func (w *Wallet) GetAll(ctx context.Context) ([]models.WalletDTO, error) {
	return w.WalletClient.GetAll(ctx)
}

func (w *Wallet) Create(ctx context.Context) (string, error) {
	return w.WalletClient.Create(ctx)
}

func (w *Wallet) Delete(ctx context.Context, ID string) error {
	return w.WalletClient.Delete(ctx, ID)
}

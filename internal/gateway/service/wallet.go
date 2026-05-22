package service

import (
	"context"

	"github.com/SlayerSv/payments/internal/wallet/models"
)

type WalletClient interface {
	Get(ctx context.Context) (models.Wallet, error)
	Update(ctx context.Context, name, password *string) (models.Wallet, error)
}

type Wallet struct {
	WalletClient WalletClient
}

func NewWallet(authClient AuthClient, WalletClient WalletClient) *Wallet {
	return &Wallet{WalletClient: WalletClient}
}

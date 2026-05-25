package adapter

import (
	"context"
	"fmt"

	pb "github.com/SlayerSv/payments/gen/wallet"
	"github.com/SlayerSv/payments/internal/shared/models"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Wallet struct {
	WalletClient pb.WalletServiceClient
}

func NewWallet(walletClient pb.WalletServiceClient) *Wallet {
	return &Wallet{WalletClient: walletClient}
}

func (w *Wallet) Get(ctx context.Context, ID string) (models.WalletDTO, error) {
	resp, err := w.WalletClient.Get(ctx, &pb.GetRequest{Id: ID})
	if err != nil {
		return models.WalletDTO{}, fmt.Errorf("error calling wallet client: %w", err)
	}
	return pbToWallet(resp.GetWallet()), nil
}

func (w *Wallet) GetAll(ctx context.Context) ([]models.WalletDTO, error) {
	resp, err := w.WalletClient.GetAll(ctx, &emptypb.Empty{})
	if err != nil {
		return []models.WalletDTO{}, fmt.Errorf("error calling wallet client: %w", err)
	}
	wallets := []models.WalletDTO{}
	for _, wallet := range resp.GetWallets() {
		wallets = append(wallets, pbToWallet(wallet))
	}
	return wallets, nil
}

func (w *Wallet) Create(ctx context.Context) (string, error) {
	resp, err := w.WalletClient.Create(ctx, &emptypb.Empty{})
	if err != nil {
		return "", fmt.Errorf("error calling wallet client: %w", err)
	}
	return resp.GetWalletId(), nil
}

func (w *Wallet) Delete(ctx context.Context, ID string) error {
	_, err := w.WalletClient.Delete(ctx, &pb.DeleteRequest{Id: ID})
	if err != nil {
		return fmt.Errorf("error calling wallet client: %w", err)
	}
	return nil
}

func pbToWallet(wallet *pb.Wallet) models.WalletDTO {
	return models.WalletDTO{
		ID:        wallet.Id,
		OwnerID:   wallet.OwnerId,
		Balance:   wallet.Balance,
		CreatedAt: wallet.CreateAt.String(),
	}
}

package service

import (
	"context"

	"github.com/google/uuid"
)

type TransClient interface {
	Deposit(ctx context.Context, walletID uuid.UUID, amount int64) error
	Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error
	Transfer(ctx context.Context, donorWalletID, receiverWalletID uuid.UUID, amount int64) error
}

type Trans struct {
	transClient TransClient
}

func NewTrans(transClient TransClient) *Trans {
	return &Trans{transClient: transClient}
}

func (t *Trans) Deposit(ctx context.Context, walletID uuid.UUID, amount int64) error {
	return t.transClient.Deposit(ctx, walletID, amount)
}

func (t *Trans) Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error {
	return t.transClient.Withdraw(ctx, walletID, amount)
}

func (t *Trans) Transfer(ctx context.Context, donorWalletID, receiverWalletID uuid.UUID, amount int64) error {
	return t.transClient.Transfer(ctx, donorWalletID, receiverWalletID, amount)
}

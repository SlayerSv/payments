package adapter

import (
	"context"
	"fmt"

	pb "github.com/SlayerSv/payments/gen/trans"
	"github.com/SlayerSv/payments/internal/shared/models"
)

type Trans struct {
	transClient pb.TransServiceClient
}

func NewTrans(transClient pb.TransServiceClient) *Trans {
	return &Trans{transClient: transClient}
}

func (t *Trans) Deposit(ctx context.Context, walletID string, amount int64) (int64, error) {
	resp, err := t.transClient.Deposit(ctx, &pb.DepositRequest{WalletId: walletID, Amount: amount})
	if err != nil {
		return 0, fmt.Errorf("error calling trans client: %w", err)
	}
	return resp.GetNewBalance(), nil
}

func (t *Trans) Withdraw(ctx context.Context, walletID string, amount int64) (int64, error) {
	resp, err := t.transClient.Withdraw(ctx, &pb.WithdrawRequest{WalletId: walletID, Amount: amount})
	if err != nil {
		return 0, fmt.Errorf("error calling trans client: %w", err)
	}
	return resp.GetNewBalance(), nil
}

func (t *Trans) Transfer(ctx context.Context, donorWalletID, receiverWalletID string, amount int64) (int64, error) {
	resp, err := t.transClient.Transfer(ctx, &pb.TransferRequest{
		DonorWalletId:    donorWalletID,
		ReceiverWalletId: receiverWalletID,
		Amount:           amount})
	if err != nil {
		return 0, fmt.Errorf("error calling trans client: %w", err)
	}
	return resp.GetNewBalance(), nil
}

func (t *Trans) GetTransactionHistory(ctx context.Context, walletID string) (models.TransactionHistory, error) {
	resp, err := t.transClient.GetTransactionHistory(ctx, &pb.GetTransactionHistoryRequest{WalletId: walletID})
	if err != nil {
		return models.TransactionHistory{}, fmt.Errorf("error calling trans client: %w", err)
	}
	transHistory := models.TransactionHistory{Transactions: []models.TransactionDTO{}}
	for _, trans := range resp.Transactions {
		transHistory.Transactions = append(transHistory.Transactions, pbToTrans(trans))
	}
	return transHistory, nil
}

func pbToTrans(trans *pb.Transaction) models.TransactionDTO {
	tr := models.TransactionDTO{}
	tr.ID = trans.Id
	tr.OpType = trans.OpType.String()
	tr.Amount = trans.Amount
	tr.CompletedAt = trans.CreatedAt.String()
	if trans.OpType == pb.OperationType_DEPOSIT || trans.OpType == pb.OperationType_TRANSFER {
		tr.ReceiverWalletID = trans.ReceiverWalletId
	}
	if trans.OpType == pb.OperationType_WITHDRAW || trans.OpType == pb.OperationType_TRANSFER {
		tr.DonorWalletID = trans.DonorWalletId
	}
	return tr
}

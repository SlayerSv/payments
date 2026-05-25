package service

import (
	"context"

	"github.com/SlayerSv/payments/internal/shared/models"
)

type TransClient interface {
	Deposit(ctx context.Context, walletID string, amount int64) (int64, error)
	Withdraw(ctx context.Context, walletID string, amount int64) (int64, error)
	Transfer(ctx context.Context, donorWalletID, receiverWalletID string, amount int64) (int64, error)
	GetTransactionHistory(ctx context.Context, walletID string) (models.TransactionHistory, error)
}

type Trans struct {
	transClient TransClient
}

func NewTrans(transClient TransClient) *Trans {
	return &Trans{transClient: transClient}
}

func (t *Trans) Deposit(ctx context.Context, walletID string, amount int64) (int64, error) {
	return t.transClient.Deposit(ctx, walletID, amount)
}

func (t *Trans) Withdraw(ctx context.Context, walletID string, amount int64) (int64, error) {
	return t.transClient.Withdraw(ctx, walletID, amount)
}

func (t *Trans) Transfer(ctx context.Context, donorWalletID, receiverWalletID string, amount int64) (int64, error) {
	return t.transClient.Transfer(ctx, donorWalletID, receiverWalletID, amount)
}

func (t *Trans) GetTransactionHistory(ctx context.Context, walletID string) (models.TransactionHistory, error) {

	// ids := map[string]string{}
	// for _, trans := range resp.Transactions {
	// 	if trans.DonorWalletId != nil {
	// 		ids[*trans.DonorWalletId] = ""
	// 	}
	// 	if trans.ReceiverWalletId != nil {
	// 		ids[*trans.ReceiverWalletId] = ""
	// 	}
	// }
	// idreq := &authpb.GetEmailsRequest{}
	// for id := range ids {
	// 	idreq.Ids = append(idreq.Ids, id)
	// }
	// idemail, err := app.Clients.User.GetEmails(ctx, idreq)
	// if err != nil {
	// 	app.ErrorJSON(w, r, fmt.Errorf("%w: error getting emails: %w", errs.Internal, err))
	// 	return
	// }
	return t.transClient.GetTransactionHistory(ctx, walletID)
}

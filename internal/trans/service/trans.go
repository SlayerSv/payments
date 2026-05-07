package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	pb "github.com/SlayerSv/payments/gen/auth"
	walletpb "github.com/SlayerSv/payments/gen/wallet"
	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/trans/models"
	"github.com/SlayerSv/payments/internal/trans/repository"
)

type Transaction struct {
	repo   repository.Transaction
	user   pb.UserServiceClient
	wallet walletpb.WalletServiceClient
}

func NewTransaction(repo repository.Transaction, userClient pb.UserServiceClient, walletClient walletpb.WalletServiceClient) *Transaction {
	return &Transaction{
		repo:   repo,
		user:   userClient,
		wallet: walletClient,
	}
}

func (s *Transaction) Deposit(ctx context.Context, userID, accountID uuid.UUID, amount int64) (int64, error) {
	acc, err := s.wallet.Get(ctx, &walletpb.GetRequest{Id: accountID.String()})
	if err != nil {
		return 0, fmt.Errorf("%w: error getting account: %w", errs.Internal, err)
	}
	if acc.Account.OwnerId != userID.String() {
		return 0, fmt.Errorf("%w: account's owner id does not match user id", errs.Forbidden)
	}
	tx := models.Transaction{
		ReceiverAccountID: &accountID,
		Amount:            amount,
		OpType:            models.OperationDeposit,
	}
	id, err := s.repo.Create(ctx, tx)
	if err != nil {
		return 0, err
	}
	req := &walletpb.ProcessOperationRequest{}
	req.IdempotencyKey = id.String() + " " + "CREDIT"
	req.TransactionId = id.String()
	req.AccountId = accountID.String()
	req.Amount = amount
	resp, err := s.wallet.ProcessOperation(ctx, req)
	if err != nil {
		s.repo.UpdateStatus(ctx, id, models.StatusFailed)
		return 0, err
	}
	err = s.repo.UpdateStatus(ctx, id, models.StatusCompleted)
	if err != nil {
		return resp.NewBalance, err
	}
	return resp.NewBalance, nil
}

func (s *Transaction) Withdraw(ctx context.Context, userID, accountID uuid.UUID, amount int64) (int64, error) {
	acc, err := s.wallet.Get(ctx, &walletpb.GetRequest{Id: accountID.String()})
	if err != nil {
		return 0, fmt.Errorf("%w: error getting account: %w", errs.Internal, err)
	}
	if acc.Account.OwnerId != userID.String() {
		return 0, fmt.Errorf("%w: account's owner id does not match user id", errs.Forbidden)
	}
	tx := models.Transaction{
		DonorAccountID: &accountID,
		Amount:         amount,
		OpType:         models.OperationWithdraw,
	}
	id, err := s.repo.Create(ctx, tx)
	if err != nil {
		return 0, err
	}
	req := &walletpb.ProcessOperationRequest{}
	req.IdempotencyKey = id.String() + " " + "DEBIT"
	req.TransactionId = id.String()
	req.AccountId = accountID.String()
	req.Amount = -amount
	resp, err := s.wallet.ProcessOperation(ctx, req)
	if err != nil {
		s.repo.UpdateStatus(ctx, id, models.StatusFailed)
		return 0, err
	}
	err = s.repo.UpdateStatus(ctx, id, models.StatusCompleted)
	return resp.NewBalance, err
}

func (s *Transaction) Transfer(ctx context.Context, userID uuid.UUID, trans models.Transfer) (int64, error) {
	acc, err := s.wallet.Get(ctx, &walletpb.GetRequest{Id: trans.DonorAccountID.String()})
	if err != nil {
		return 0, fmt.Errorf("%w: error getting account: %w", errs.Internal, err)
	}
	if acc.Account.OwnerId != userID.String() {
		return 0, fmt.Errorf("%w: donor account's owner id does not match user id", errs.Forbidden)
	}
	tx := models.Transaction{
		DonorAccountID:    &trans.DonorAccountID,
		ReceiverAccountID: &trans.ReceiverAccountID,
		Amount:            trans.Amount,
		OpType:            models.OperationTransfer,
	}
	id, err := s.repo.Create(ctx, tx)
	if err != nil {
		return 0, err
	}
	debreq := &walletpb.ProcessOperationRequest{}
	debreq.IdempotencyKey = id.String() + " " + "DEBIT"
	debreq.TransactionId = id.String()
	debreq.AccountId = trans.DonorAccountID.String()
	debreq.Amount = -trans.Amount
	resp, err := s.wallet.ProcessOperation(ctx, debreq)
	if err != nil {
		s.repo.UpdateStatus(ctx, id, models.StatusFailed)
		return 0, err
	}
	err = s.repo.UpdateStatus(ctx, id, models.StatusDebitSuccess)

	credreq := &walletpb.ProcessOperationRequest{}
	credreq.IdempotencyKey = id.String() + " " + "CREDIT"
	credreq.TransactionId = id.String()
	credreq.AccountId = trans.ReceiverAccountID.String()
	credreq.Amount = trans.Amount
	_, err = s.wallet.ProcessOperation(ctx, credreq)
	if err != nil {
		err = s.repo.UpdateStatus(ctx, id, models.StatusRollbackPending)
		debreq.IdempotencyKey = id.String() + " " + "ROLLBACK"
		debreq.Amount = -debreq.Amount
		rollbackresp, err := s.wallet.ProcessOperation(ctx, debreq)
		s.repo.UpdateStatus(ctx, id, models.StatusFailed)
		return rollbackresp.NewBalance, err
	}
	err = s.repo.UpdateStatus(ctx, id, models.StatusCompleted)
	return resp.NewBalance, err
}

func (s *Transaction) GetTransactionHistory(ctx context.Context, userID, accountID uuid.UUID) ([]models.Transaction, error) {
	acc, err := s.wallet.Get(ctx, &walletpb.GetRequest{Id: accountID.String()})
	if err != nil {
		return nil, fmt.Errorf("%w: error getting account: %w", errs.Internal, err)
	}
	if acc.Account.OwnerId != userID.String() {
		return nil, fmt.Errorf("%w: donor account's owner id does not match user id", errs.Forbidden)
	}
	return s.repo.GetTransactionHistory(ctx, accountID)
}

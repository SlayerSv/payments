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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Объявляем счетчик (Counter) с лейблом "status"
var (
	SagaCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "saga_transactions",
			Help: "Общее количество транзакций саги по статусам",
		},
		[]string{"status"}, // TOTAL, COMPLETED, FAILED, ROLLBACK
	)
	DepositCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "deposits",
			Help: "Общее количество депозитов по статусам",
		},
		[]string{"status"}, // TOTAL, SUCCESS
	)
	WithdrawCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "withdraws",
			Help: "Общее количество выводов средств по статусам",
		},
		[]string{"status"}, // TOTAL, SUCCESS
	)
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

func (s *Transaction) Deposit(ctx context.Context, userID, walletID uuid.UUID, amount int64) (int64, error) {
	DepositCounter.WithLabelValues("TOTAL").Inc()
	wallet, err := s.wallet.Get(ctx, &walletpb.GetRequest{Id: walletID.String()})
	if err != nil {
		return 0, fmt.Errorf("%w: error getting wallet: %w", errs.Internal, err)
	}
	if wallet.Wallet.OwnerId != userID.String() {
		return 0, fmt.Errorf("%w: wallet's owner id does not match user id", errs.Forbidden)
	}
	tx := models.Transaction{
		ReceiverWalletID: &walletID,
		Amount:           amount,
		OpType:           models.OperationDeposit,
	}
	id, err := s.repo.Create(ctx, tx)
	if err != nil {
		return 0, err
	}
	req := &walletpb.ProcessOperationRequest{}
	req.IdempotencyKey = id.String() + " " + "CREDIT"
	req.TransactionId = id.String()
	req.WalletId = walletID.String()
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
	DepositCounter.WithLabelValues("SUCCESS").Inc()
	return resp.NewBalance, nil
}

func (s *Transaction) Withdraw(ctx context.Context, userID, walletID uuid.UUID, amount int64) (int64, error) {
	WithdrawCounter.WithLabelValues("TOTAL").Inc()
	wallet, err := s.wallet.Get(ctx, &walletpb.GetRequest{Id: walletID.String()})
	if err != nil {
		return 0, fmt.Errorf("%w: error getting wallet: %w", errs.Internal, err)
	}
	if wallet.Wallet.OwnerId != userID.String() {
		return 0, fmt.Errorf("%w: wallet's owner id does not match user id", errs.Forbidden)
	}
	tx := models.Transaction{
		DonorWalletID: &walletID,
		Amount:        amount,
		OpType:        models.OperationWithdraw,
	}
	id, err := s.repo.Create(ctx, tx)
	if err != nil {
		return 0, err
	}
	req := &walletpb.ProcessOperationRequest{}
	req.IdempotencyKey = id.String() + " " + "DEBIT"
	req.TransactionId = id.String()
	req.WalletId = walletID.String()
	req.Amount = -amount
	resp, err := s.wallet.ProcessOperation(ctx, req)
	if err != nil {
		s.repo.UpdateStatus(ctx, id, models.StatusFailed)
		return 0, err
	}
	err = s.repo.UpdateStatus(ctx, id, models.StatusCompleted)
	DepositCounter.WithLabelValues("SUCCESS").Inc()
	return resp.NewBalance, err
}

func (s *Transaction) Transfer(ctx context.Context, userID uuid.UUID, trans models.Transfer) (int64, error) {
	wallet, err := s.wallet.Get(ctx, &walletpb.GetRequest{Id: trans.DonorWalletID.String()})
	if err != nil {
		return 0, fmt.Errorf("%w: error getting wallet: %w", errs.Internal, err)
	}
	if wallet.Wallet.OwnerId != userID.String() {
		return 0, fmt.Errorf("%w: donor wallet's owner id does not match user id", errs.Forbidden)
	}
	tx := models.Transaction{
		DonorWalletID:    &trans.DonorWalletID,
		ReceiverWalletID: &trans.ReceiverWalletID,
		Amount:           trans.Amount,
		OpType:           models.OperationTransfer,
	}
	id, err := s.repo.Create(ctx, tx)
	if err != nil {
		return 0, err
	}
	SagaCounter.WithLabelValues("TOTAL").Inc()
	debreq := &walletpb.ProcessOperationRequest{}
	debreq.IdempotencyKey = id.String() + " " + "DEBIT"
	debreq.TransactionId = id.String()
	debreq.WalletId = trans.DonorWalletID.String()
	debreq.Amount = -trans.Amount
	resp, err := s.wallet.ProcessOperation(ctx, debreq)
	if err != nil {
		SagaCounter.WithLabelValues("FAILED").Inc()
		s.repo.UpdateStatus(ctx, id, models.StatusFailed)
		return 0, err
	}
	err = s.repo.UpdateStatus(ctx, id, models.StatusDebitSuccess)

	credreq := &walletpb.ProcessOperationRequest{}
	credreq.IdempotencyKey = id.String() + " " + "CREDIT"
	credreq.TransactionId = id.String()
	credreq.WalletId = trans.ReceiverWalletID.String()
	credreq.Amount = trans.Amount
	_, err = s.wallet.ProcessOperation(ctx, credreq)
	if err != nil {
		SagaCounter.WithLabelValues("ROLLBACK").Inc()
		err = s.repo.UpdateStatus(ctx, id, models.StatusRollbackPending)
		debreq.IdempotencyKey = id.String() + " " + "ROLLBACK"
		debreq.Amount = -debreq.Amount
		rollbackresp, err := s.wallet.ProcessOperation(ctx, debreq)
		s.repo.UpdateStatus(ctx, id, models.StatusFailed)
		return rollbackresp.NewBalance, err
	}
	err = s.repo.UpdateStatus(ctx, id, models.StatusCompleted)
	SagaCounter.WithLabelValues("COMPLETED").Inc()
	return resp.NewBalance, err
}

func (s *Transaction) GetTransactionHistory(ctx context.Context, userID, walletID uuid.UUID) ([]models.Transaction, error) {
	wallet, err := s.wallet.Get(ctx, &walletpb.GetRequest{Id: walletID.String()})
	if err != nil {
		return nil, fmt.Errorf("%w: error getting wallet: %w", errs.Internal, err)
	}
	if wallet.Wallet.OwnerId != userID.String() {
		return nil, fmt.Errorf("%w: donor wallet's owner id does not match user id", errs.Forbidden)
	}
	return s.repo.GetTransactionHistory(ctx, walletID)
}

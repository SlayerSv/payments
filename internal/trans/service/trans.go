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

func (s *Transaction) Deposit(ctx context.Context, user_id uuid.UUID, accType models.AccountType, amount int64) (uuid.UUID, error) {
	if amount <= 0 {
		return uuid.Nil, fmt.Errorf("%w: amount must be positive", errs.BadRequest)
	}
	if accType == models.AccountInvalid {
		return uuid.Nil, fmt.Errorf("%w: invalid account type", errs.BadRequest)
	}
	tx := models.Transaction{
		ReceiverID:   user_id,
		ReceiverType: accType,
		Amount:       amount,
		OpType:       models.OperationDeposit,
	}
	return s.repo.Create(ctx, tx)
}

func (s *Transaction) Withdraw(ctx context.Context, user_id uuid.UUID, accType models.AccountType, amount int64) (uuid.UUID, error) {
	if amount <= 0 {
		return uuid.Nil, fmt.Errorf("%w: amount must be positive", errs.BadRequest)
	}
	if accType == models.AccountInvalid {
		return uuid.Nil, fmt.Errorf("%w: invalid account type", errs.BadRequest)
	}
	tx := models.Transaction{
		SenderID:   user_id,
		SenderType: accType,
		Amount:     amount,
		OpType:     models.OperationWithdraw,
	}
	return s.repo.Create(ctx, tx)
}

func (s *Transaction) Transfer(
	ctx context.Context,
	senderID uuid.UUID,
	senderAccType models.AccountType,
	receiverEmail string,
	receiverAccType models.AccountType,
	amount int64) (uuid.UUID, error) {
	if amount <= 0 {
		return uuid.Nil, fmt.Errorf("%w: amount must be positive", errs.BadRequest)
	}
	if senderAccType == models.AccountInvalid {
		return uuid.Nil, fmt.Errorf("%w: invalid account type", errs.BadRequest)
	}
	receiver, err := s.user.GetByEmail(ctx, &pb.GetByEmailRequest{Email: receiverEmail})
	if err != nil {
		return uuid.Nil, fmt.Errorf("error getting receiver id: %w", err)
	}
	receiverID, err := uuid.Parse(receiver.GetId())
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: error parsing user id: %w", errs.Internal, err)
	}
	tx := models.Transaction{
		SenderID:     senderID,
		SenderType:   senderAccType,
		ReceiverID:   receiverID,
		ReceiverType: receiverAccType,
		Amount:       amount,
		OpType:       models.OperationTransfer,
	}
	return s.repo.Create(ctx, tx)
}

func (s *Transaction) GetAccHistory(ctx context.Context, accountID uuid.UUID) ([]models.Transaction, error) {
	return s.repo.GetAccHistory(ctx, accountID)
}

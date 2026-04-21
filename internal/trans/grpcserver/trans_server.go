package grpcserver

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/SlayerSv/payments/gen/trans"
	"github.com/SlayerSv/payments/internal/trans/models"
	"github.com/SlayerSv/payments/internal/trans/service"
)

type Trans struct {
	trans.UnimplementedTransServiceServer
	service *service.Transaction
}

func NewTrans(svc *service.Transaction) *Trans {
	return &Trans{service: svc}
}

// Вспомогательная функция для маппинга енамов
func mapProtoType(p trans.AccountType) models.AccountType {
	switch p {
	case trans.AccountType_SAVINGS:
		return models.AccountSavings
	case trans.AccountType_WALLET:
		return models.AccountWallet
	}
	return models.AccountInvalid
}

func (s *Trans) Deposit(ctx context.Context, req *trans.DepositRequest) (*trans.DepositResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}
	_, err = s.service.Deposit(ctx, userID, mapProtoType(req.AccType), req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deposit failed: %v", err)
	}
	return &trans.DepositResponse{Status: "ok"}, nil
}

func (s *Trans) Withdraw(ctx context.Context, req *trans.WithdrawRequest) (*trans.WithdrawResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}
	_, err = s.service.Withdraw(ctx, userID, mapProtoType(req.AccType), req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "withdraw failed: %v", err)
	}
	return &trans.WithdrawResponse{Status: "ok"}, nil
}

func (s *Trans) Transfer(ctx context.Context, req *trans.TransferRequest) (*trans.TransferResponse, error) {
	senderID, err := uuid.Parse(req.SenderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid sender_id format")
	}
	_, err = s.service.Transfer(ctx, senderID, mapProtoType(req.SenderAccType), req.ReceiverEmail, mapProtoType(req.ReceiverAccType), req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "transfer failed: %v", err)
	}
	return &trans.TransferResponse{Status: "ok"}, nil
}

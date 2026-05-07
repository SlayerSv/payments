package grpcserver

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/SlayerSv/payments/gen/trans"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"github.com/SlayerSv/payments/internal/trans/models"
	"github.com/SlayerSv/payments/internal/trans/service"
)

type Trans struct {
	pb.UnimplementedTransServiceServer
	service *service.Transaction
}

func NewTrans(svc *service.Transaction) *Trans {
	return &Trans{service: svc}
}

// Вспомогательная функция для маппинга енамов
func mapAccType(p pb.AccountType) models.AccountType {
	switch p {
	case pb.AccountType_SAVINGS:
		return models.AccountSavings
	case pb.AccountType_WALLET:
		return models.AccountWallet
	}
	return models.AccountUnspecified
}

func mapAccToPbType(m models.AccountType) pb.AccountType {
	switch m {
	case models.AccountWallet:
		return pb.AccountType_WALLET
	case models.AccountSavings:
		return pb.AccountType_SAVINGS
	default:
		return pb.AccountType_ACCOUNT_UNSPECIFIED
	}
}

func mapOpType(t models.OperationType) pb.OperationType {
	switch t {
	case models.OperationDeposit:
		return pb.OperationType_DEPOSIT
	case models.OperationWithdraw:
		return pb.OperationType_WITHDRAW
	case models.OperationTransfer:
		return pb.OperationType_TRANSFER
	default:
		return pb.OperationType_OPERATION_UNSPECIFIED
	}
}

func (s *Trans) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.NewBalanceResponse, error) {
	userID, err := uuid.Parse(ctx.Value(interceptors.UserID).(string))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id format: %v", err)
	}
	accID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing account id: %v", err)
	}
	newBalance, err := s.service.Deposit(ctx, userID, accID, mapAccType(req.AccountType), req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deposit failed: %v", err)
	}
	return &pb.NewBalanceResponse{NewBalance: newBalance}, nil
}

func (s *Trans) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*pb.NewBalanceResponse, error) {
	userID, err := uuid.Parse(ctx.Value(interceptors.UserID).(string))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id format: %v", err)
	}
	accID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing account id: %v", err)
	}
	newBalance, err := s.service.Withdraw(ctx, userID, accID, mapAccType(req.AccountType), req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "withdraw failed: %v", err)
	}
	return &pb.NewBalanceResponse{NewBalance: newBalance}, nil
}

func (s *Trans) Transfer(ctx context.Context, req *pb.TransferRequest) (*pb.NewBalanceResponse, error) {
	userID, err := uuid.Parse(ctx.Value(interceptors.UserID).(string))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id format: %v", err)
	}
	donorAccID, err := uuid.Parse(req.DonorAccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid donor account id format: %v", err)
	}
	receiverAccID, err := uuid.Parse(req.ReceiverAccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid receiver account id format: %v", err)
	}
	transreq := models.Transfer{
		DonorID:           userID,
		DonorAccountID:    donorAccID,
		ReceiverAccountID: receiverAccID,
		Amount:            req.Amount,
	}
	newBalance, err := s.service.Transfer(ctx, userID, transreq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "transfer failed: %v", err)
	}
	return &pb.NewBalanceResponse{NewBalance: newBalance}, nil
}

func (s *Trans) GetTransactionHistory(ctx context.Context, req *pb.GetTransactionHistoryRequest) (*pb.GetTransactionHistoryResponse, error) {
	userID, err := uuid.Parse(ctx.Value(interceptors.UserID).(string))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id format: %v", err)
	}
	accID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_id format: %v", err)
	}
	trans, err := s.service.GetTransactionHistory(ctx, userID, accID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting history: %v", err)
	}
	resp := &pb.GetTransactionHistoryResponse{}
	for _, tran := range trans {
		tr := &pb.Transaction{}
		tr.Id = tran.ID.String()
		tr.OpType = mapOpType(tran.OpType)
		tr.Amount = tran.Amount
		tr.CreatedAt = timestamppb.New(tran.UpdatedAt)
		if tr.OpType == pb.OperationType_DEPOSIT || tr.OpType == pb.OperationType_TRANSFER {
			raccid := tran.ReceiverAccountID.String()
			tr.ReceiverAccountId = &raccid
			racctype := mapAccToPbType(*tran.ReceiverAccountType)
			tr.ReceiverAccountType = &racctype
		}
		if tr.OpType == pb.OperationType_WITHDRAW || tr.OpType == pb.OperationType_TRANSFER {
			daccid := tran.DonorAccountID.String()
			tr.DonorAccountId = &daccid
			dacctype := mapAccToPbType(*tran.DonorAccountType)
			tr.DonorAccountType = &dacctype
		}
		resp.Transactions = append(resp.Transactions, tr)
	}
	return resp, nil
}

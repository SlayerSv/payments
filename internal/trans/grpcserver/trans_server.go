package grpcserver

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *Trans) Deposit(ctx context.Context, req *pb.DepositRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(ctx.Value(interceptors.UserID).(string))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id format: %v", err)
	}
	_, err = s.service.Deposit(ctx, userID, mapAccType(req.AccountType), req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deposit failed: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Trans) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(ctx.Value(interceptors.UserID).(string))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id format: %v", err)
	}
	_, err = s.service.Withdraw(ctx, userID, mapAccType(req.AccountType), req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "withdraw failed: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Trans) Transfer(ctx context.Context, req *pb.TransferRequest) (*emptypb.Empty, error) {
	senderID, err := uuid.Parse(req.SenderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid sender_id format: %v", err)
	}
	_, err = s.service.Transfer(ctx, senderID, mapAccType(req.SenderType), req.ReceiverEmail, mapAccType(req.ReceiverType), req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "transfer failed: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Trans) GetTransactionHistory(ctx context.Context, req *emptypb.Empty) (*pb.GetTransactionHistoryResponse, error) {
	userID, err := uuid.Parse(ctx.Value(interceptors.UserID).(string))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id format: %v", err)
	}
	trans, err := s.service.GetTransactionHistory(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting history: %v", err)
	}
	resp := &pb.GetTransactionHistoryResponse{}
	for _, tran := range trans {
		resp.Transactions = append(resp.Transactions, &pb.Transaction{
			Id:           tran.ID.String(),
			OpType:       mapOpType(tran.OpType),
			SenderId:     tran.SenderID.String(),
			SenderType:   mapAccToPbType(tran.SenderType),
			ReceiverId:   tran.ReceiverID.String(),
			ReceiverType: mapAccToPbType(tran.ReceiverType),
			Amount:       tran.Amount,
			CreatedAt:    timestamppb.New(tran.UpdatedAt),
		})
	}
	return resp, nil
}

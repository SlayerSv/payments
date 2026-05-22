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
	accID, err := uuid.Parse(req.WalletId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing wallet id: %v", err)
	}
	newBalance, err := s.service.Deposit(ctx, userID, accID, req.Amount)
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
	accID, err := uuid.Parse(req.WalletId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing wallet id: %v", err)
	}
	newBalance, err := s.service.Withdraw(ctx, userID, accID, req.Amount)
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
	donorAccID, err := uuid.Parse(req.DonorWalletId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid donor wallet id format: %v", err)
	}
	receiverAccID, err := uuid.Parse(req.ReceiverWalletId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid receiver wallet id format: %v", err)
	}
	transreq := models.Transfer{
		DonorWalletID:    donorAccID,
		ReceiverWalletID: receiverAccID,
		Amount:           req.Amount,
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
	accID, err := uuid.Parse(req.WalletId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid wallet_id format: %v", err)
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
			raccid := tran.ReceiverWalletID.String()
			tr.ReceiverWalletId = &raccid
		}
		if tr.OpType == pb.OperationType_WITHDRAW || tr.OpType == pb.OperationType_TRANSFER {
			daccid := tran.DonorWalletID.String()
			tr.DonorWalletId = &daccid
		}
		resp.Transactions = append(resp.Transactions, tr)
	}
	return resp, nil
}

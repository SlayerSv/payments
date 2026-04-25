package grpcserver

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/SlayerSv/payments/gen/wallet"
	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"github.com/SlayerSv/payments/internal/wallet/models"
	"github.com/SlayerSv/payments/internal/wallet/service"
)

type Wallet struct {
	pb.UnimplementedWalletServiceServer
	walletService *service.Wallet
}

func NewWallet(ws *service.Wallet) *Wallet {
	return &Wallet{
		walletService: ws,
	}
}

func (s *Wallet) ProcessOperation(ctx context.Context, req *pb.ProcessOperationRequest) (*pb.ProcessOperationResponse, error) {
	// 1. Валидация и парсинг UUID
	txID, err := uuid.Parse(req.TransactionId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid transaction_id format: %v", err)
	}

	accID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_id format: %v", err)
	}

	if req.IdempotencyKey == "" {
		return nil, status.Errorf(codes.InvalidArgument, "idempotency_key is required")
	}

	// 2. Формируем запрос для бизнес-логики
	opReq := models.OperationRequest{
		IdempotencyKey: req.IdempotencyKey,
		TransactionID:  txID,
		AccountID:      accID,
		AmountDelta:    req.AmountDelta,
	}

	// 3. Вызываем сервис
	resp, err := s.walletService.ProcessOperation(ctx, opReq)
	if err != nil {
		// Маппинг бизнес-ошибок в gRPC-ошибки
		if errors.Is(err, errs.InsufficientFunds) {
			// FailedPrecondition идеально подходит для бизнес-отказов (нет денег)
			return nil, status.Errorf(codes.FailedPrecondition, "insufficient funds")
		}
		if errors.Is(err, errs.MaxRetriesReached) {
			// Aborted значит, что транзакция отвалилась из-за конкуренции, клиент может попробовать еще раз
			return nil, status.Errorf(codes.Aborted, "system is busy, please retry")
		}

		// Любая другая непредвиденная ошибка (упала БД и т.д.)
		return nil, status.Errorf(codes.Internal, "internal server error: %v", err)
	}

	// 4. Возвращаем успешный ответ
	return &pb.ProcessOperationResponse{
		Id:         resp.AccountID.String(),
		NewBalance: resp.NewBalance,
		Status:     resp.Status,
	}, nil
}

// CreateWallet — регистрация кошелька
func (s *Wallet) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	ownerID, err := uuid.Parse(ctx.Value(interceptors.UserID).(string))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid owner_id format")
	}

	accountID, err := s.walletService.Create(ctx, ownerID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create wallet: %v", err)
	}

	return &pb.CreateResponse{
		AccountId: accountID.String(),
	}, nil
}

func (s *Wallet) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	ownerID, err := uuid.Parse(ctx.Value(interceptors.UserID).(string))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid owner_id format")
	}
	accID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_id format")
	}

	acc, err := s.walletService.GetAccount(ctx, ownerID, accID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get account: %v", err)
	}
	return &pb.GetResponse{
		Account: toPbAccount(acc),
	}, nil
}

func (s *Wallet) GetAll(ctx context.Context, req *pb.GetAllRequest) (*pb.GetAllResponse, error) {
	ownerID, err := uuid.Parse(ctx.Value(interceptors.UserID).(string))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid owner_id format")
	}

	accs, err := s.walletService.GetAccounts(ctx, ownerID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get accounts: %v", err)
	}
	resp := &pb.GetAllResponse{}
	for _, acc := range accs {
		resp.Accounts = append(resp.Accounts, toPbAccount(acc))
	}
	return resp, nil
}

func (s *Wallet) Delete(ctx context.Context, req *pb.DeleteRequest) (*emptypb.Empty, error) {
	ownerID, err := uuid.Parse(ctx.Value(interceptors.UserID).(string))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid owner_id format")
	}
	accID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_id format")
	}

	err = s.walletService.DeleteAccount(ctx, ownerID, accID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete account: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func toPbAccount(acc models.Account) *pb.Account {
	return &pb.Account{
		Id:       acc.ID.String(),
		OwnerId:  acc.OwnerID.String(),
		Balance:  acc.Balance,
		CreateAt: timestamppb.New(acc.CreatedAt),
	}
}

package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/SlayerSv/payments/gen/wallet" // Импорт сгенерированного кода
	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/wallet/models"
	"github.com/SlayerSv/payments/internal/wallet/service" // Наша бизнес-лоinternal/wallet/service
)

// Wallet реализует сгенерированный интерфейс WalletServiceServer
type Wallet struct {
	pb.UnimplementedWalletServiceServer
	walletService service.Wallet
}

func NewWallet(ws service.Wallet) *Wallet {
	return &Wallet{
		walletService: ws,
	}
}

// ProcessOperation — тот самый метод, в который будет стучаться сервис Транзакций
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
		AccountId:  resp.AccountID.String(),
		NewBalance: resp.NewBalance,
		Status:     resp.Status,
	}, nil
}

// CreateWallet — регистрация кошелька
func (s *Wallet) CreateWallet(ctx context.Context, req *pb.CreateWalletRequest) (*pb.CreateWalletResponse, error) {
	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid owner_id format")
	}

	walletID, err := s.walletService.CreateWallet(ctx, ownerID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create wallet: %v", err)
	}

	return &pb.CreateWalletResponse{
		AccountId: walletID.String(),
	}, nil
}

// GetBalance — узнать текущий баланс
func (s *Wallet) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	accID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_id format")
	}

	balance, err := s.walletService.GetBalance(ctx, accID)
	if err != nil {
		// Здесь можно добавить проверку: если кошелек не найден, возвращать codes.NotFound
		return nil, status.Errorf(codes.Internal, "failed to get balance: %v", err)
	}

	return &pb.GetBalanceResponse{
		Balance: balance,
	}, nil
}

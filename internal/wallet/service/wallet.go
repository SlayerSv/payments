package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/wallet/models"
	"github.com/SlayerSv/payments/internal/wallet/repository"
)

const (
	MaxOptimisticRetries = 5 // Сколько раз пытаться обновить баланс при конфликтах
	OutboxTopicWalletOp  = "wallet.operation.applied"
)

type Wallet struct {
	repo repository.Wallet
}

func NewWallet(repo repository.Wallet) *Wallet {
	return &Wallet{repo: repo}
}

// CreateWallet — регистрация нового кошелька
func (s *Wallet) CreateWallet(ctx context.Context, ownerID uuid.UUID) (uuid.UUID, error) {
	return s.repo.CreateAccount(ctx, ownerID)
}

func (s *Wallet) GetAccount(ctx context.Context, accountID uuid.UUID) (models.Account, error) {
	return s.repo.GetAccount(ctx, accountID)
}

func (s *Wallet) GetAccounts(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	return s.repo.GetAccounts(ctx, userID)
}

// ProcessOperation — Изменение баланса с ретраями и идемпотентностью (ГЛАВНЫЙ МЕТОД)
func (s *Wallet) ProcessOperation(ctx context.Context, req models.OperationRequest) (models.OperationResponse, error) {
	// 1. ПРОВЕРКА ИДЕМПОТЕНТНОСТИ
	// Если мы уже обрабатывали этот запрос (например, сеть моргнула, и клиент послал gRPC запрос еще раз),
	// мы не списываем деньги снова, а просто отдаем сохраненный ответ.
	if req.IdempotencyKey != "" {
		cachedRespBytes, exists, err := s.repo.GetIdempotencyResponse(ctx, req.IdempotencyKey)
		if err != nil {
			return models.OperationResponse{}, fmt.Errorf("failed to check idempotency: %w", err)
		}
		if exists {
			var cachedResp models.OperationResponse
			_ = json.Unmarshal(cachedRespBytes, &cachedResp)
			return cachedResp, nil // Возвращаем старый успешный ответ
		}
	}

	// 2. ЦИКЛ РЕТРАЕВ (Оптимистичная блокировка)
	// Если параллельный процесс успеет изменить баланс быстрее нас, версия БД изменится,
	// и repo вернет ErrConcurrentUpdate. В этом случае мы идем на новый круг.
	for attempt := 0; attempt < MaxOptimisticRetries; attempt++ {

		// 2.1 Получаем актуальное состояние кошелька (баланс и ВЕРСИЮ)
		acc, err := s.repo.GetAccount(ctx, req.AccountID)
		if err != nil {
			return models.OperationResponse{}, err
		}

		// 2.2 Бизнес-валидация: Проверка на отрицательный баланс при списании
		newBalance := acc.Balance + req.AmountDelta
		if req.AmountDelta < 0 && newBalance < 0 {
			return models.OperationResponse{}, errs.InsufficientFunds
		}

		// 2.3 Готовим успешный ответ
		response := models.OperationResponse{
			AccountID:  acc.ID,
			NewBalance: newBalance,
			Status:     "APPLIED",
		}
		respBytes, _ := json.Marshal(response)

		// 2.4 Готовим событие для Outbox (чтобы Kafka узнала об изменении)
		outboxPayload := map[string]interface{}{
			"transaction_id": req.TransactionID,
			"account_id":     acc.ID,
			"amount_delta":   req.AmountDelta,
			"new_balance":    newBalance,
		}
		outboxBytes, _ := json.Marshal(outboxPayload)

		// 2.5 Пытаемся применить транзакцию в БД
		updateParams := models.UpdateBalanceParams{
			AccountID:           acc.ID,
			TransactionID:       req.TransactionID,
			AmountDelta:         req.AmountDelta,
			ExpectedVersion:     acc.Version, // Важно! Передаем версию, которую прочитали
			OutboxTopic:         OutboxTopicWalletOp,
			OutboxPayload:       outboxBytes,
			IdempotencyKey:      req.IdempotencyKey,
			IdempotencyResponse: respBytes,
		}

		err = s.repo.UpdateBalanceAtomic(ctx, updateParams)

		// 2.6 Обработка результата
		if err == nil {
			// Успех! Выходим из функции
			return response, nil
		}

		// Если это ошибка конфликта версий — идем на следующий круг цикла (ретрай)
		if errors.Is(err, errs.ConcurrentUpdate) {
			continue // ВАЖНО: Пытаемся снова (прочитаем новые данные и попробуем обновить)
		}

		// Если это какая-то другая ошибка БД (упала сеть, нет места на диске) — возвращаем её
		return models.OperationResponse{}, err
	}

	// 3. ЕСЛИ ИСЧЕРПАЛИ ВСЕ ПОПЫТКИ
	// Такое бывает при огромной нагрузке (сотни конкурентных списаний с 1 кошелька в секунду).
	return models.OperationResponse{}, errs.MaxRetriesReached
}

func (s *Wallet) DeleteAccount(ctx context.Context, accountID uuid.UUID) error {
	return s.repo.DeleteAccount(ctx, accountID)
}

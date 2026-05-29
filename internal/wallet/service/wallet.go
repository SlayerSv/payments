package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/wallet/models"
	"github.com/SlayerSv/payments/internal/wallet/repository"
)

const (
	MaxOptimisticRetries = 5 // Сколько раз пытаться обновить баланс при конфликтах
	OutboxTopicWalletOp  = "wallet.operation.applied"
)

var (
	OperationCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "operations",
			Help: "Общее количество операций с кошельками по статусам",
		},
		[]string{"status"}, // TOTAL, SUCCESS, CACHED
	)
)

type Wallet struct {
	repo repository.Wallet
}

func NewWallet(repo repository.Wallet) *Wallet {
	return &Wallet{repo: repo}
}

// CreateWallet — регистрация нового кошелька
func (s *Wallet) Create(ctx context.Context, ownerID uuid.UUID) (uuid.UUID, error) {
	return s.repo.Create(ctx, ownerID)
}

func (s *Wallet) Get(ctx context.Context, ownerID, walletID uuid.UUID) (models.Wallet, error) {
	return s.repo.Get(ctx, ownerID, walletID)
}

func (s *Wallet) GetAll(ctx context.Context, userID uuid.UUID) ([]models.Wallet, error) {
	return s.repo.GetAll(ctx, userID)
}

// ProcessOperation — Изменение баланса с ретраями и идемпотентностью (ГЛАВНЫЙ МЕТОД)
func (s *Wallet) ProcessOperation(ctx context.Context, req models.OperationRequest) (models.OperationResponse, error) {
	OperationCounter.WithLabelValues("TOTAL").Inc()
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
			OperationCounter.WithLabelValues("CACHED").Inc()
			return cachedResp, nil // Возвращаем старый успешный ответ
		}
	}

	// 2. ЦИКЛ РЕТРАЕВ (Оптимистичная блокировка)
	// Если параллельный процесс успеет изменить баланс быстрее нас, версия БД изменится,
	// и repo вернет ErrConcurrentUpdate. В этом случае мы идем на новый круг.
	for attempt := 0; attempt < MaxOptimisticRetries; attempt++ {
		// 2.1 Получаем актуальное состояние кошелька (баланс и ВЕРСИЮ)
		wallet, err := s.repo.Get(ctx, req.OwnerID, req.WalletID)
		if err != nil {
			return models.OperationResponse{}, err
		}

		// 2.2 Бизнес-валидация: Проверка на отрицательный баланс при списании
		newBalance := wallet.Balance + req.Amount
		if req.Amount < 0 && newBalance < 0 {
			return models.OperationResponse{}, errs.InsufficientFunds
		}

		// 2.3 Готовим успешный ответ
		response := models.OperationResponse{
			WalletID:   wallet.ID,
			NewBalance: newBalance,
			Status:     "APPLIED",
		}
		respBytes, _ := json.Marshal(response)

		// 2.4 Готовим событие для Outbox (чтобы Kafka узнала об изменении)
		outboxPayload := map[string]interface{}{
			"transaction_id": req.TransactionID,
			"wallet_id":      wallet.ID,
			"amount":         req.Amount,
			"new_balance":    newBalance,
		}
		outboxBytes, _ := json.Marshal(outboxPayload)

		// 2.5 Пытаемся применить транзакцию в БД
		updateParams := models.UpdateBalanceParams{
			WalletID:            wallet.ID,
			TransactionID:       req.TransactionID,
			Amount:              req.Amount,
			ExpectedVersion:     wallet.Version, // Важно! Передаем версию, которую прочитали
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
		OperationCounter.WithLabelValues("SUCCESS").Inc()
		// Если это какая-то другая ошибка БД (упала сеть, нет места на диске) — возвращаем её
		return models.OperationResponse{}, err
	}

	// 3. ЕСЛИ ИСЧЕРПАЛИ ВСЕ ПОПЫТКИ
	// Такое бывает при огромной нагрузке (сотни конкурентных списаний с 1 кошелька в секунду).
	return models.OperationResponse{}, errs.MaxRetriesReached
}

func (s *Wallet) Delete(ctx context.Context, ownerID, walletID uuid.UUID) error {
	return s.repo.Delete(ctx, ownerID, walletID)
}

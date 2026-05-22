package models

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Balance   int64
	Version   int64
	CreatedAt time.Time
}

type LedgerEntry struct {
	ID            uuid.UUID
	WalletID      uuid.UUID
	TransactionID uuid.UUID
	Amount        int64
	CreatedAt     time.Time
}

type OutboxMessage struct {
	ID      int64
	Topic   string
	Payload []byte
	Status  string
}

type UpdateBalanceParams struct {
	WalletID            uuid.UUID
	TransactionID       uuid.UUID // ID из сервиса транзакций (Correlation ID)
	Amount              int64     // Сколько прибавить (положительное) или отнять (отрицательное)
	ExpectedVersion     int64     // Текущая версия для Optimistic Lock
	OutboxTopic         string    // Топик для кафки/брокера
	OutboxPayload       []byte    // JSON payload
	IdempotencyKey      string    // Ключ для защиты от дублей
	IdempotencyResponse []byte    // Что вернуть при повторном запросе
}

// OperationRequest — DTO для изменения баланса
type OperationRequest struct {
	IdempotencyKey string    // Уникальный ключ запроса (от Transaction Service)
	TransactionID  uuid.UUID // ID транзакции для связи
	OwnerID        uuid.UUID // ID владельца кошелька
	WalletID       uuid.UUID // Кошелек, который меняем
	Amount         int64     // Сумма: положительная (пополнение) или отрицательная (списание)
}

// OperationResponse — DTO ответа
type OperationResponse struct {
	WalletID   uuid.UUID `json:"wallet_id"`
	NewBalance int64     `json:"new_balance"`
	Status     string    `json:"status"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	AccountInvalid AccountType = "INVALID"
	AccountWallet  AccountType = "WALLET"
	AccountSavings AccountType = "SAVINGS"
	AccountSystem  AccountType = "SYSTEM"
)

type TransactionStatus string

const (
	StatusCreated         TransactionStatus = "CREATED"
	StatusDebitSuccess    TransactionStatus = "DEBIT_SUCCESS"
	StatusCompleted       TransactionStatus = "COMPLETED"
	StatusRollbackPending TransactionStatus = "ROLLBACK_PENDING"
	StatusFailed          TransactionStatus = "FAILED"
)

type OperationType string

const (
	OperationDeposit  OperationType = "DEBIT"
	OperationWithdraw OperationType = "WITHDRAW"
	OperationTransfer OperationType = "TRANSFER"
)

type Transaction struct {
	ID           uuid.UUID
	OpType       OperationType
	SenderID     uuid.UUID
	SenderType   AccountType
	ReceiverID   uuid.UUID
	ReceiverType AccountType
	Amount       int64
	Status       TransactionStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

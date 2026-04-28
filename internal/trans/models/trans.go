package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	AccountUnspecified AccountType = "UNSPECIFIED"
	AccountWallet      AccountType = "WALLET"
	AccountSavings     AccountType = "SAVINGS"
)

func GetAccountType(acctype string) AccountType {
	acc := strings.TrimSpace(strings.ToLower(acctype))
	switch acc {
	case "wallet":
		return AccountWallet
	case "savings":
		return AccountSavings
	default:
		return AccountUnspecified
	}
}

type TransactionStatus string

const (
	StatusUnspecified     TransactionStatus = "UNSPECIFIED"
	StatusCreated         TransactionStatus = "CREATED"
	StatusDebitSuccess    TransactionStatus = "DEBIT_SUCCESS"
	StatusCompleted       TransactionStatus = "COMPLETED"
	StatusRollbackPending TransactionStatus = "ROLLBACK_PENDING"
	StatusFailed          TransactionStatus = "FAILED"
)

type OperationType string

const (
	OperationUnspecified OperationType = "UNSPECIFIED"
	OperationDeposit     OperationType = "DEBIT"
	OperationWithdraw    OperationType = "WITHDRAW"
	OperationTransfer    OperationType = "TRANSFER"
)

func GetOperationType(acctype string) OperationType {
	acc := strings.TrimSpace(strings.ToLower(acctype))
	switch acc {
	case "debit":
		return OperationDeposit
	case "withdraw":
		return OperationWithdraw
	case "transfer":
		return OperationTransfer
	default:
		return OperationUnspecified
	}
}

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

type TransactionDTO struct {
	ID            string    `json:"id"`
	OpType        string    `json:"op_type"`
	SenderID      string    `json:"sender_id"`
	SenderEmail   string    `json:"sender_email"`
	SenderType    string    `json:"sender_type"`
	ReceiverID    string    `json:"receiver_id"`
	ReceiverEmail string    `json:"receiver_email"`
	ReceiverType  string    `json:"receiver_type"`
	Amount        int64     `json:"amount"`
	CompletedAt   time.Time `json:"completed_at"`
}

type DepositRequest struct {
	AccountType string `json:"account_type" validate:"required,account_type"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
}

type WithdrawRequest struct {
	AccountType string `json:"account_type" validate:"required,account_type"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
}

type TransferRequest struct {
	SenderID      string `json:"sender_id" validate:"required,uuid"`
	SenderType    string `json:"sender_type" validate:"required,account_type"`
	ReceiverEmail string `json:"receiver_email" validate:"required,email"`
	ReceiverType  string `json:"receiver_type" validate:"required,account_type"`
	Amount        int64  `json:"amount" validate:"required,gt=0"`
}

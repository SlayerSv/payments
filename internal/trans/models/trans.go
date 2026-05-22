package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

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
	OperationDeposit     OperationType = "DEPOSIT"
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
	ID               uuid.UUID
	OpType           OperationType
	DonorWalletID    *uuid.UUID
	ReceiverWalletID *uuid.UUID
	Amount           int64
	Status           TransactionStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Transfer struct {
	DonorWalletID    uuid.UUID `json:"donor_wallet_id" validate:"required,uuid"`
	ReceiverWalletID uuid.UUID `json:"receiver_wallet_id" validate:"required,uuid"`
	Amount           int64     `json:"amount" validate:"required,gt=0"`
}

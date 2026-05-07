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
	ID                uuid.UUID
	OpType            OperationType
	DonorAccountID    *uuid.UUID
	ReceiverAccountID *uuid.UUID
	Amount            int64
	Status            TransactionStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type TransactionDTO struct {
	ID                string    `json:"id"`
	OpType            string    `json:"op_type"`
	DonorAccountID    *string   `json:"donor_account_id,omitzero"`
	DonorEmail        *string   `json:"donor_email,omitzero"`
	DonorName         *string   `json:"donor_name,omitzero"`
	ReceiverAccountID *string   `json:"receiver_account_id,omitzero"`
	ReceiverEmail     *string   `json:"receiver_email,omitzero"`
	ReceiverName      *string   `json:"receiver_name,omitzero"`
	Amount            int64     `json:"amount"`
	CompletedAt       time.Time `json:"completed_at"`
}

type TransactionHistory struct {
	Transactions []TransactionDTO `json:"transactions"`
}

type DepositRequest struct {
	Amount int64 `json:"amount" validate:"required,gt=0"`
}

type WithdrawRequest struct {
	Amount int64 `json:"amount" validate:"required,gt=0"`
}

type TransferRequest struct {
	ReceiverAccountID string `json:"receiver_account_id" validate:"required,uuid"`
	Amount            int64  `json:"amount" validate:"required,gt=0"`
}

type Transfer struct {
	DonorAccountID    uuid.UUID `json:"donor_account_id" validate:"required,uuid"`
	ReceiverAccountID uuid.UUID `json:"receiver_account_id" validate:"required,uuid"`
	Amount            int64     `json:"amount" validate:"required,gt=0"`
}

type NewBalanceResponse struct {
	NewBalance int64 `json:"new_balance"`
}

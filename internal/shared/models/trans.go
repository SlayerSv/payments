package models

type TransactionDTO struct {
	ID               string  `json:"id"`
	OpType           string  `json:"op_type"`
	DonorWalletID    *string `json:"donor_wallet_id,omitzero"`
	DonorEmail       *string `json:"donor_email,omitzero"`
	DonorName        *string `json:"donor_name,omitzero"`
	ReceiverWalletID *string `json:"receiver_wallet_id,omitzero"`
	ReceiverEmail    *string `json:"receiver_email,omitzero"`
	ReceiverName     *string `json:"receiver_name,omitzero"`
	Amount           int64   `json:"amount"`
	CompletedAt      string  `json:"completed_at"`
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
	ReceiverWalletID string `json:"receiver_wallet_id" validate:"required,uuid"`
	Amount           int64  `json:"amount" validate:"required,gt=0"`
}

type NewBalanceResponse struct {
	NewBalance int64 `json:"new_balance"`
}

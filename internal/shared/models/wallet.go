package models

type WalletID struct {
	ID string `json:"id"`
}

type WalletDTO struct {
	ID        string `json:"id"`
	OwnerID   string `json:"owner_id"`
	Balance   int64  `json:"balance"`
	CreatedAt string `json:"created_at"`
}

type WalletsDTO struct {
	Wallets []WalletDTO `json:"wallets"`
}

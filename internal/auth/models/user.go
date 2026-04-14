package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Name         *string   `json:"name,omitzero"`
	PasswordHash *string   `json:"-"` // Скрываем из JSON, используем указатель для nullable
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

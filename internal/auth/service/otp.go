package service

import (
	"context"

	"github.com/SlayerSv/payments/internal/auth/models"
	"github.com/SlayerSv/payments/internal/auth/repo"
)

type OTP struct {
	DB repo.OTP
}

func NewOTP(db repo.OTP) *OTP {
	return &OTP{DB: db}
}

func (us *OTP) Create(ctx context.Context, otp models.OTP) (models.OTP, error) {
	return us.DB.Create(ctx, otp)
}

func (us *OTP) Get(ctx context.Context, email string) (models.OTP, error) {
	return us.DB.Get(ctx, email)
}

func (us *OTP) Delete(ctx context.Context, email string) (string, error) {
	return us.DB.Delete(ctx, email)
}

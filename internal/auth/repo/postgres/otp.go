package postgres

import (
	"context"

	"github.com/SlayerSv/payments/internal/auth/models"
	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OTP struct {
	pool *pgxpool.Pool
}

func NewOTP(pool *pgxpool.Pool) *OTP {
	return &OTP{pool: pool}
}

func (r *OTP) Create(ctx context.Context, otp models.OTP) (models.OTP, error) {
	query := `
		INSERT INTO otp (email, code, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) 
		DO UPDATE SET 
			code = EXCLUDED.code,
			expires_at = EXCLUDED.expires_at,
			created_at = NOW()
		RETURNING email, code, expires_at, created_at;
	`
	err := r.pool.QueryRow(ctx, query, otp.Email, otp.Code, otp.ExpiresAt).
		Scan(&otp.Email, &otp.Code, &otp.ExpiresAt, &otp.CreatedAt)
	return otp, errs.WrapErr(err)
}

func (r *OTP) Get(ctx context.Context, email string) (models.OTP, error) {
	otp := models.OTP{}
	query := `SELECT email, code, expires_at, created_at 
	          FROM otp 
	          WHERE email = $1`

	err := r.pool.QueryRow(ctx, query, email).
		Scan(&otp.Email, &otp.Code, &otp.ExpiresAt, &otp.CreatedAt)
	return otp, errs.WrapErr(err)
}

func (r *OTP) Delete(ctx context.Context, email string) (string, error) {
	err := r.pool.QueryRow(ctx, "DELETE FROM otp WHERE email = $1 RETURNING email", email).Scan(&email)
	return email, errs.WrapErr(err)
}

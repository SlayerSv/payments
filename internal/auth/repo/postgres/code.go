package postgres

import (
	"context"

	"github.com/SlayerSv/payments/internal/auth/models"
	"github.com/SlayerSv/payments/pkg/errs"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"
)

type OTPCode struct {
	pool *pgxpool.Pool
}

func NewOTPCode(pool *pgxpool.Pool) *OTPCode {
	return &OTPCode{pool: pool}
}

func (r *OTPCode) Create(ctx context.Context, otp models.OTPCode) (models.OTPCode, error) {
	query := `INSERT INTO otp_codes (user_id, code, expires_at) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := r.pool.QueryRow(ctx, query, otp.UserID, otp.Code, otp.ExpiresAt).Scan(&otp.ID, &otp.CreatedAt)
	return otp, errs.WrapErr(err)
}

func (r *OTPCode) Get(ctx context.Context, userID uuid.UUID) (models.OTPCode, error) {
	otp := models.OTPCode{}
	query := `SELECT id, user_id, code, expires_at, created_at 
	          FROM otp_codes 
	          WHERE user_id = $1`

	err := r.pool.QueryRow(ctx, query, userID).Scan(&otp.ID, &otp.UserID, &otp.Code, &otp.ExpiresAt, &otp.CreatedAt)
	return otp, errs.WrapErr(err)
}

func (r *OTPCode) Delete(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var newID uuid.UUID
	err := r.pool.QueryRow(ctx, "DELETE FROM otp_codes WHERE id = $1 RETURNING id", id).Scan(&newID)
	return newID, errs.WrapErr(err)
}

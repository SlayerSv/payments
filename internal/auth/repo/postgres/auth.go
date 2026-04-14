package postgres

import (
	"context"

	"github.com/SlayerSv/payments/internal/auth/models"
	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Auth struct {
	pool *pgxpool.Pool
}

func NewAuth(pool *pgxpool.Pool) *Auth {
	return &Auth{pool: pool}
}

func (r *Auth) Register(ctx context.Context, otp models.OTP) (uuid.UUID, error) {
	var id uuid.UUID
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return id, errs.WrapErr(err)
	}
	defer tx.Rollback(ctx)
	queryUser := `INSERT INTO users (email) VALUES ($1) RETURNING id`
	err = tx.QueryRow(ctx, queryUser, otp.Email).Scan(&id)
	if err != nil {
		return id, errs.WrapErr(err)
	}
	queryOTP := `INSERT INTO otp (email, code, expires_at) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, queryOTP, otp.Email, otp.Code, otp.ExpiresAt)
	if err != nil {
		return id, errs.WrapErr(err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return id, errs.WrapErr(err)
	}
	return id, nil
}
